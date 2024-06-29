package handlers

import (
	"MusicCatGo/commands"
	"MusicCatGo/musicbot"
	"context"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

type Handlers struct {
	*musicbot.Bot
}

func (h *Handlers) OnVoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	if event.VoiceState.UserID != h.Client.ApplicationID() {
		botVoiceState, ok := h.Client.Caches().VoiceState(event.VoiceState.GuildID, h.Client.ApplicationID())
		if !ok || event.OldVoiceState.ChannelID == nil {
			return
		}

		var (
			voiceUsers     int
			userDeafened   bool
			userUndeafened bool
		)

		h.Client.Caches().VoiceStatesForEach(event.VoiceState.GuildID, func(vs discord.VoiceState) {
			if *vs.ChannelID == *botVoiceState.ChannelID {
				voiceUsers++
				if vs.UserID == event.VoiceState.UserID {
					if event.VoiceState.SelfDeaf && !event.OldVoiceState.SelfDeaf {
						userDeafened = true
					} else if !event.VoiceState.SelfDeaf && event.OldVoiceState.SelfDeaf {
						userUndeafened = true
					}
				}
			}
		})
		if voiceUsers <= 1 {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := h.Client.UpdateVoiceState(ctx, event.VoiceState.GuildID, nil, false, false); err != nil {
				slog.Error("failed to disconnect from voice channel", slog.Any("error", err))
			}
		} else if voiceUsers == 2 {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if userDeafened {
				commands.Pause(h.Lavalink, ctx, event.VoiceState.GuildID)
			} else if userUndeafened {
				commands.Resume(h.Lavalink, ctx, event.VoiceState.GuildID)
			}
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	h.Lavalink.OnVoiceStateUpdate(ctx, event.VoiceState.GuildID, event.VoiceState.ChannelID, event.VoiceState.SessionID)
}

func (h *Handlers) OnVoiceServerUpdate(event *events.VoiceServerUpdate) {
	if event.Endpoint == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	h.Lavalink.OnVoiceServerUpdate(ctx, event.GuildID, event.Token, *event.Endpoint)
}

func (h *Handlers) OnTrackStart(p disgolink.Player, event lavalink.TrackStartEvent) {
	slog.Info("Track started")

}

func (h *Handlers) OnTrackEnd(p disgolink.Player, event lavalink.TrackEndEvent) {
	if !event.Reason.MayStartNext() {
		return
	}
	track, ok := h.PlayerManager.Next(p.GuildID())
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := p.Update(ctx, lavalink.WithTrack(track)); err != nil {
		// channelID := h.PlayerManager.Player(p.GuildID())
		// if channelID == 0 {
		// 	return
		// }
		// if _, err = h.Client.Rest().CreateMessage(channelID, discord.MessageCreate{
		// 	Content:         "failed to start next track: " + err.Error(),
		// 	AllowedMentions: &discord.AllowedMentions{},
		// }); err != nil {
		// 	slog.Error("failed to send message", tint.Err(err))
		// }
		slog.Error("failed to send message")
	}
}
