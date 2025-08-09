package handlers

import (
	"MusicCatGo/musicbot"
	"context"
	"fmt"
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

			h.PlayerManager.Stop(&h.Lavalink, ctx, event.VoiceState.GuildID)

			if err := h.Client.UpdateVoiceState(ctx, event.VoiceState.GuildID, nil, false, false); err != nil {
				slog.Error("failed to disconnect from voice channel", slog.Any("error", err))
			}
		} else if voiceUsers == 2 {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if userDeafened {
				h.PlayerManager.Pause(&h.Lavalink, ctx, event.VoiceState.GuildID)
			} else if userUndeafened {
				h.PlayerManager.Resume(&h.Lavalink, ctx, event.VoiceState.GuildID)
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

	state, ok := h.PlayerManager.GetState(p.GuildID())
	if !ok || state.ChannelID() == 0 {
		return
	}

	playerEmbed := createPlayerEmbed(event.Track, state)
	playerMessage, err := h.Client.Rest().CreateMessage(state.ChannelID(), playerEmbed)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to send message %s", err.Error()))
		return
	}
	state.SetMessageID(playerMessage.ID)
}

func (h *Handlers) OnTrackEnd(p disgolink.Player, event lavalink.TrackEndEvent) {

	state, ok := h.PlayerManager.GetState(p.GuildID())
	if !ok || state.ChannelID() == 0 || state.MessageID() == 0 {
		slog.Error("failed to fetch old player message data")
		return
	}

	if err := h.Client.Rest().DeleteMessage(state.ChannelID(), state.MessageID()); err != nil {
		slog.Error("failed to delete old player message")
	}

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
		slog.Error("failed to send message")
	}
}
