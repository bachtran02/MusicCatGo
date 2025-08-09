package handlers

import (
	"MusicCatGo/commands"
	"MusicCatGo/musicbot"
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type ButtonID string

const (
	PlayPrevious ButtonID = "play_previous"
	PlayNext     ButtonID = "play_next"
	PausePlayer  ButtonID = "pause_player"
	ResumePlayer ButtonID = "resume_player"
	StopPlayer   ButtonID = "stop_player"
	LoopQueue    ButtonID = "loop_queue"
	LoopTrack    ButtonID = "loop_track"
	LoopOff      ButtonID = "loop_off"
	ShuffleOn    ButtonID = "shuffle_on"
	ShuffleOff   ButtonID = "shuffle_off"
)

func (h *Handlers) OnPlayerInteraction(event *events.ComponentInteractionCreate) {

	state, ok := h.PlayerManager.GetState(*event.GuildID())
	if !ok || state.MessageID() == 0 {
		return
	}

	ctx := context.TODO()

	switch ButtonID(event.ComponentInteraction.ButtonInteractionData().CustomID()) {
	case PlayNext:
		h.PlayerManager.Skip(&h.Lavalink, ctx, *event.GuildID())
	case StopPlayer:
		h.PlayerManager.Stop(&h.Lavalink, ctx, *event.GuildID())
	case ResumePlayer:
		h.PlayerManager.Resume(&h.Lavalink, ctx, *event.GuildID())
		updatePlayerEmbed(state, event)
	case PausePlayer:
		h.PlayerManager.Pause(&h.Lavalink, ctx, *event.GuildID())
		updatePlayerEmbed(state, event)
	case PlayPrevious:
		h.PlayerManager.PlayPrevious(&h.Lavalink, ctx, *event.GuildID())
		updatePlayerEmbed(state, event)
	case ShuffleOn:
		h.PlayerManager.SetShuffle(*event.GuildID(), musicbot.ShuffleOn)
		updatePlayerEmbed(state, event)
	case ShuffleOff:
		h.PlayerManager.SetShuffle(*event.GuildID(), musicbot.ShuffleOff)
		updatePlayerEmbed(state, event)
	case LoopOff:
		h.PlayerManager.SetLoop(*event.GuildID(), musicbot.LoopNone)
		updatePlayerEmbed(state, event)
	case LoopTrack:
		h.PlayerManager.SetLoop(*event.GuildID(), musicbot.LoopTrack)
		updatePlayerEmbed(state, event)
	case LoopQueue:
		h.PlayerManager.SetLoop(*event.GuildID(), musicbot.LoopQueue)
		updatePlayerEmbed(state, event)
	}
}

func updatePlayerEmbed(state *musicbot.PlayerState, event *events.ComponentInteractionCreate) {
	buttons := createButtons(state)
	messageBuilder := discord.NewMessageUpdateBuilder()
	messageBuilder.
		AddActionRow(buttons[0], buttons[1], buttons[2]).
		AddActionRow(buttons[3], buttons[4], buttons[5])

	event.UpdateMessage(messageBuilder.Build())
}

func createPlayerEmbed(track lavalink.Track, state *musicbot.PlayerState) discord.MessageCreate {

	embedBuilder := createEmbed(track)
	messageBuilder := discord.NewMessageCreateBuilder().SetEmbeds(embedBuilder.Build())
	return addButtonsNew(state, messageBuilder).Build()
}

func addButtonsNew(state *musicbot.PlayerState, messageBuilder *discord.MessageCreateBuilder) *discord.MessageCreateBuilder {

	buttons := createButtons(state)

	return messageBuilder.
		AddActionRow(buttons[0], buttons[1], buttons[2]).
		AddActionRow(buttons[3], buttons[4], buttons[5])
}

func createButtons(state *musicbot.PlayerState) []discord.ButtonComponent {
	var (
		playPauseButton discord.ButtonComponent
		repeatButton    discord.ButtonComponent
		shuffleButton   discord.ButtonComponent

		playPreviousButton = discord.NewSecondaryButton("", string(PlayPrevious)).WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(musicbot.PLAYER_PREVIOUS_EMOJI_ID)})
		playNextButton     = discord.NewSecondaryButton("", string(PlayNext)).WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(musicbot.PLAYER_NEXT_EMOJI_ID)})
		stopButton         = discord.NewSecondaryButton("", string(StopPlayer)).WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(musicbot.STOP_PLAYER_EMOJI_ID)})
	)

	if state.Paused() {
		playPauseButton = discord.NewSecondaryButton("", string(ResumePlayer)).WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(musicbot.RESUME_PLAYER_EMOJI_ID)})
	} else {
		playPauseButton = discord.NewSecondaryButton("", string(PausePlayer)).WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(musicbot.PAUSE_PLAYER_EMOJI_ID)})
	}

	switch state.Loop() {
	case musicbot.LoopNone:
		repeatButton = discord.NewSecondaryButton("", string(LoopQueue)).WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(musicbot.LOOP_OFF_EMOJI_ID)})
	case musicbot.LoopQueue:
		repeatButton = discord.NewSecondaryButton("", string(LoopTrack)).WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(musicbot.LOOP_QUEUE_EMOJI_ID)})
	case musicbot.LoopTrack:
		repeatButton = discord.NewSecondaryButton("", string(LoopOff)).WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(musicbot.LOOP_TRACK_EMOJI_ID)})
	}

	if state.Shuffle() {
		shuffleButton = discord.NewSecondaryButton("", string(ShuffleOff)).WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(musicbot.SHUFFLE_ON_EMOJI_ID)})
	} else {
		shuffleButton = discord.NewSecondaryButton("", string(ShuffleOn)).WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(musicbot.SHUFFLE_OFF_EMOJI_ID)})
	}

	return []discord.ButtonComponent{
		playPreviousButton, playPauseButton, playNextButton,
		stopButton, repeatButton, shuffleButton,
	}
}

func createEmbed(track lavalink.Track) discord.EmbedBuilder {
	var (
		playtime string
		userData commands.UserData
	)

	_ = track.UserData.Unmarshal(&userData)

	if track.Info.IsStream {
		playtime = "LIVE"
	} else {
		playtime = musicbot.FormatTime(track.Info.Length)
	}

	return *discord.NewEmbedBuilder().
		SetDescriptionf("[%s](%s)\n%s `%s`\n\n<@%s>",
			track.Info.Title, *track.Info.URI, track.Info.Author,
			playtime, userData.Requester).
		SetThumbnail(*track.Info.ArtworkURL)
}
