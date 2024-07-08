package handlers

import (
	"MusicCatGo/musicbot"
	"MusicCatGo/utils"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func (h *Handlers) OnPlayerInteraction(event *events.ComponentInteractionCreate) {

}

func createPlayerEmbed(track lavalink.Track, state *musicbot.PlayerState, requester string) discord.MessageCreate {
	var playtime string

	if track.Info.IsStream {
		playtime = "LIVE"
	} else {
		playtime = utils.FormatTime(track.Info.Length)
	}

	embedBuilder := *discord.NewEmbedBuilder().
		SetDescriptionf("[%s](%s)\n%s `%s`\n\n<@%s>",
			track.Info.Title, *track.Info.URI, track.Info.Author,
			playtime, requester).
		SetThumbnail(*track.Info.ArtworkURL)

	messageBuilder := discord.NewMessageCreateBuilder().SetEmbeds(embedBuilder.Build())
	return addPlayerButtons(state, messageBuilder).Build()
}

func addPlayerButtons(state *musicbot.PlayerState, messageBuilder *discord.MessageCreateBuilder) *discord.MessageCreateBuilder {
	var (
		playPauseButton discord.ButtonComponent
		repeatButton    discord.ButtonComponent
		shuffleButton   discord.ButtonComponent

		playPreviousButton = discord.NewSecondaryButton("", "play_previous").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.PLAYER_PREVIOUS_EMOJI_ID)})
		playNextButton     = discord.NewSecondaryButton("", "play_next").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.PLAYER_NEXT_EMOJI_ID)})
		stopButton         = discord.NewSecondaryButton("", "stop_player").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.STOP_PLAYER_EMOJI_ID)})
	)

	if state.Paused() {
		playPauseButton = discord.NewSecondaryButton("", "resume_player").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.RESUME_PLAYER_EMOJI_ID)})
	} else {
		playPauseButton = discord.NewSecondaryButton("", "pause_player").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.PAUSE_PLAYER_EMOJI_ID)})
	}

	switch state.Repeat() {
	case musicbot.RepeatModeNone:
		repeatButton = discord.NewSecondaryButton("", "loop_queue").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.LOOP_OFF_EMOJI_ID)})
	case musicbot.RepeatModeQueue:
		repeatButton = discord.NewSecondaryButton("", "loop_track").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.LOOP_QUEUE_EMOJI_ID)})
	case musicbot.RepeatModeTrack:
		repeatButton = discord.NewSecondaryButton("", "loop_off").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.LOOP_TRACK_EMOJI_ID)})
	}

	if state.Shuffle() {
		shuffleButton = discord.NewSecondaryButton("", "shuffle_off").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.SHUFFLE_ON_EMOJI_ID)})
	} else {
		shuffleButton = discord.NewSecondaryButton("", "shuffle_on").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.SHUFFLE_OFF_EMOJI_ID)})
	}

	return messageBuilder.
		AddActionRow(playPreviousButton, playPauseButton, playNextButton).
		AddActionRow(stopButton, repeatButton, shuffleButton)
}
