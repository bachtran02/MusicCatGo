package main

import (
	"MusicCatGo/commands"
	"MusicCatGo/handlers"
	"MusicCatGo/musicbot"
	"context"
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
	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()

	slog.Info("starting MusicCat...")
	slog.Info("disgo version", slog.String("version", disgo.Version))

	var err error
	b := &musicbot.Bot{PlayerManager: *musicbot.NewPlayerManager()}
	cmds := &commands.Commands{Bot: b}

	r := handler.New()
	r.Use(middleware.Go)
	r.Route("/bot", func(r handler.Router) {
		r.SlashCommand("/join", cmds.Connect)
		r.SlashCommand("/leave", cmds.Disconnect)
		r.SlashCommand("/ping", cmds.Ping)
	})
	r.Route("/music", func(r handler.Router) {
		r.SlashCommand("/play", cmds.Play)
		r.SlashCommand("/search", cmds.Play)
		r.Autocomplete("/search", cmds.SearchAutocomplete)
		r.SlashCommand("/resume", cmds.Resume)
		r.SlashCommand("/pause", cmds.Pause)
		r.SlashCommand("/queue", cmds.Queue)
	})

	hdlr := &handlers.Handlers{Bot: b}

	b.Client, err = disgo.New(os.Getenv("TOKEN"),
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

	if err = b.Start(); err != nil {
		slog.Error("failed to start bot", slog.Any("err", err))
		os.Exit(-1)
	}

	defer b.Client.Close(context.TODO())

	slog.Info("MusicCat is now running.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
