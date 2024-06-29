package handlers

import (
	"MusicCatGo/utils"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func createPlayerEmbed(track lavalink.Track, requester string) discord.MessageCreate {
	var playtime string

	if track.Info.IsStream {
		playtime = "LIVE"
	} else {
		playtime = utils.FormatTime(track.Info.Length)
	}

	embedBuilder := *discord.NewEmbedBuilder().
		SetTitle("Track added").
		SetDescriptionf("[%s](%s)\n%s `%s`\n\n<@%s>",
			track.Info.Title, *track.Info.URI, track.Info.Author,
			playtime, requester).
		SetThumbnail(*track.Info.ArtworkURL)

	return discord.NewMessageCreateBuilder().SetEmbeds(embedBuilder.Build()).AddActionRow(
		discord.NewSecondaryButton("", "play_previous").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.PLAYER_PREVIOUS_EMOJI_ID)}),
		discord.NewSecondaryButton("", "pause_player").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.PAUSE_PLAYER_EMOJI_ID)}),
		discord.NewSecondaryButton("", "play_next").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.PLAYER_NEXT_EMOJI_ID)}),
	).AddActionRow(
		discord.NewSecondaryButton("", "stop_player").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.STOP_PLAYER_EMOJI_ID)}),
		discord.NewSecondaryButton("", "loop_off").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.LOOP_OFF_EMOJI_ID)}),
		discord.NewSecondaryButton("", "shuffle_off").WithEmoji(discord.ComponentEmoji{ID: snowflake.ID(utils.SHUFFLE_OFF_EMOJI_ID)}),
	).Build()
}
