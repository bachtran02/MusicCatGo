package main

import (
	"MusicCatGo/commands"
	"MusicCatGo/handlers"
	"MusicCatGo/musicbot"

	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
	"github.com/disgoorg/disgolink/v3/disgolink"
)

//go:embed db/schema.sql
var DBschema string

func main() {

	cfg, err := musicbot.ReadConfig("config.yml")
	if err != nil {
		slog.Error("failed to read config file", slog.Any("err", err))
	}

	slog.Info("starting MusicCat...")
	slog.Info("disgo version", slog.String("version", disgo.Version))
	slog.Info("disgolink version", slog.String("version", disgolink.Version))

	b := &musicbot.Bot{
		Cfg:           cfg,
		PlayerManager: *musicbot.NewPlayerManager(),
	}
	cmds := &commands.Commands{Bot: b}

	r := handler.New()
	r.Use(middleware.Go)
	r.Route("/bot", func(r handler.Router) {
		r.SlashCommand("/join", cmds.Connect)
		r.SlashCommand("/leave", cmds.Disconnect)
		r.SlashCommand("/ping", cmds.Ping)
	})
	r.Route("/music", func(r handler.Router) {
		r.SlashCommand("/loop", cmds.Loop)
		r.SlashCommand("/now", cmds.Now)
		r.SlashCommand("/pause", cmds.Pause)
		r.SlashCommand("/play", cmds.Play)
		r.SlashCommand("/playlist", cmds.PlayPlaylist)
		r.Autocomplete("/playlist", cmds.PlaylistAutocomplete)
		r.SlashCommand("/queue", cmds.Queue)
		r.SlashCommand("/remove", cmds.RemoveQueueTrack)
		r.Autocomplete("/remove", cmds.RemoveQueueTrackAutocomplete)
		r.SlashCommand("/resume", cmds.Resume)
		r.SlashCommand("/search", cmds.Play)
		r.Autocomplete("/search", cmds.SearchAutocomplete)
		r.SlashCommand("/seek", cmds.Seek)
		r.SlashCommand("/shuffle", cmds.Shuffle)
		r.SlashCommand("/skip", cmds.Skip)
		r.SlashCommand("/stop", cmds.Stop)
	})
	r.Route("/list", func(r handler.Router) {
		r.SlashCommand("/add", cmds.AddPlaylistTrack)
		r.Autocomplete("/add", cmds.AddPlaylistTrackAutocomplete)
		r.SlashCommand("/create", cmds.CreatePlaylist)
		r.SlashCommand("/delete", cmds.DeletePlaylist)
		r.Autocomplete("/delete", cmds.PlaylistAutocomplete)
		r.SlashCommand("/list", cmds.ListPlaylists)
		r.SlashCommand("/remove", cmds.RemovePlaylistTrack)
		r.Autocomplete("/remove", cmds.RemovePlaylistTrackAutocomplete)
	})

	hdlr := &handlers.Handlers{Bot: b}

	b.Client, err = disgo.New(cfg.Bot.Token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildVoiceStates,
			)),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagVoiceStates),
		),
		bot.WithEventListeners(r),
		bot.WithEventListenerFunc(hdlr.OnVoiceStateUpdate),
		bot.WithEventListenerFunc(hdlr.OnVoiceServerUpdate),
		bot.WithEventListenerFunc(hdlr.OnPlayerInteraction),
	)
	if err != nil {
		slog.Error("failed to create disgo client", slog.Any("err", err))
		return
	}

	if err = handler.SyncCommands(b.Client, commands.CommandCreates, nil); err != nil {
		slog.Error("failed to sync commands", slog.Any("err", err))
	}

	if b.Lavalink = disgolink.New(b.Client.ApplicationID(),
		disgolink.WithListenerFunc(hdlr.OnTrackStart),
		disgolink.WithListenerFunc(hdlr.OnTrackEnd),
		// disgolink.WithListenerFunc(hdlr.OnTrackException),
		// disgolink.WithListenerFunc(hdlr.OnTrackStuck),
	); err != nil {
		slog.Error("failed to create disgolink client", slog.Any("err", err))
		os.Exit(-1)
	}

	b.Db, err = musicbot.NewDB(cfg.DB, DBschema)
	if err != nil {
		slog.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(-1)
	}
	slog.Info("Connected to database")
	defer b.Db.Close()

	if err = b.Start(); err != nil {
		slog.Error("failed to start bot", slog.Any("err", err))
		os.Exit(-1)
	}
	defer b.Client.Close(context.TODO())

	slog.Info("MusicCat is now running.")

	if b.Cfg.MusicTracker.Enabled {
		/* enabling tracker server */
		wsServer := musicbot.NewWsServer(b.Cfg.MusicTracker.AllowedOrigins)
		go wsServer.Run()

		trackerHandler := handlers.TrackerHandler{
			ChannelID: b.Cfg.MusicTracker.ChannelID,
			GuildID:   b.Cfg.MusicTracker.GuildID,
			WsServer:  wsServer,
		}

		// run tracker server to serve current track
		trackerServer := musicbot.NewTrackerServer(
			wsServer,
			trackerHandler.ServeHTTP,
			b.Cfg.MusicTracker.HostAddress,
			b.Cfg.MusicTracker.HttpPath,
			b.Cfg.MusicTracker.WebsocketPath)

		go trackerServer.Start()
		defer trackerServer.Close(context.TODO())

		// register lavalink client listener
		b.Lavalink.AddListeners(
			disgolink.NewListenerFunc(trackerHandler.OnTrackStart),
			disgolink.NewListenerFunc(trackerHandler.OnTrackEnd),
			disgolink.NewListenerFunc(trackerHandler.OnPlayerUpdate),
		)
		slog.Info(
			"MusicCat music tracker is enabled",
			slog.String("http", fmt.Sprintf("%s/%s",
				b.Cfg.MusicTracker.HostAddress,
				b.Cfg.MusicTracker.HttpPath)),
			slog.String("ws", fmt.Sprintf("%s/%s",
				b.Cfg.MusicTracker.HostAddress,
				b.Cfg.MusicTracker.WebsocketPath)),
		)
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
