package handlers

import (
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	/* user updating voice state */
	if event.VoiceState.UserID != h.Client.ApplicationID() {
		botVoiceState, ok := h.Client.Caches().VoiceState(event.VoiceState.GuildID, h.Client.ApplicationID())
		if !ok || event.OldVoiceState.ChannelID == nil {
			/* bot isn't in voice channel or just joined */
			return
		}

		var (
			voiceUsers       int
			userDeafened     bool
			userUndeafened   bool
			player, playerOk = h.PlayerManager.GetPlayer(event.VoiceState.GuildID)
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
			/* there is only bot left in voice chat */
			if playerOk {
				player.ClearState()
				if err := player.StopAudio(ctx); err != nil {
					slog.Error("failed to stop audio",
						slog.Any("error", err), slog.Any("guild_id", event.VoiceState.GuildID))
				}
			}

			if err := h.Client.UpdateVoiceState(ctx, event.VoiceState.GuildID, nil, false, false); err != nil {
				slog.Error("failed to disconnect from voice channel",
					slog.Any("error", err), slog.Any("guild_id", event.VoiceState.GuildID))
			}
		} else if voiceUsers == 2 {
			/* there is bot and single user
			-> user owns control of playback with "Deafen" button */
			if playerOk && userDeafened {
				player.Pause(ctx)
			} else if playerOk && userUndeafened {
				player.Resume(ctx)
			}
		}
	} else {
		h.Lavalink.OnVoiceStateUpdate(
			ctx, event.VoiceState.GuildID,
			event.VoiceState.ChannelID, event.VoiceState.SessionID)
	}
}

func (h *Handlers) OnVoiceServerUpdate(event *events.VoiceServerUpdate) {
	if event.Endpoint != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		h.Lavalink.OnVoiceServerUpdate(ctx, event.GuildID, event.Token, *event.Endpoint)
	}
}

func (h *Handlers) OnTrackStart(p disgolink.Player, event lavalink.TrackStartEvent) {

	player, ok := h.PlayerManager.GetPlayer(p.GuildID())
	if !ok || player.ChannelID() == 0 {
		/* player is not playing or channel ID not set */
		return
	}

	playerEmbed := createPlayerEmbed(event.Track, player.IsPaused(), player.Shuffle(), player.Loop())
	playerMessage, err := h.Client.Rest().CreateMessage(player.ChannelID(), playerEmbed)
	if err != nil {
		slog.Error("failed to send player embed",
			slog.Any("error", err.Error()), slog.Any("guild_id", p.GuildID()))
		return
	}
	player.SetMessage(playerMessage) /* update message */
}

func (h *Handlers) OnTrackEnd(p disgolink.Player, event lavalink.TrackEndEvent) {

	player, ok := h.PlayerManager.GetPlayer(p.GuildID())
	if !ok || player.PlayerMessage() == nil {
		slog.Error("failed to fetch old player message data",
			slog.Any("guild_id", p.GuildID()))
		return
	}

	var (
		playerMessage = player.PlayerMessage()
		channelId     = playerMessage.ChannelID
		messageId     = playerMessage.ID
	)
	if err := h.Client.Rest().DeleteMessage(channelId, messageId); err != nil {
		slog.Error("failed to delete old player message",
			slog.Any("error", err.Error()),
			slog.Any("guild_id", p.GuildID()),
			slog.Any("channel_id", channelId),
			slog.Any("message_id", messageId),
		)
	}

	if event.Reason.MayStartNext() {
		/* starting next track in queue */
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := player.PlayNext(ctx); err != nil {
			slog.Error("failed to play next track in queue",
				slog.Any("error", err.Error()),
				slog.Any("guild_id", p.GuildID()))
		}
	}
}
