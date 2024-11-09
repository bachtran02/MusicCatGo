package commands

import (
	"MusicCatGo/musicbot"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (cmd *Commands) Queue(data discord.SlashCommandInteractionData, event *handler.CommandEvent) error {

	player := cmd.Lavalink.Player(*event.GuildID())

	track := player.Track()
	if track == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "Player is not playing",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	var userData UserData
	_ = track.UserData.Unmarshal(&userData)

	content := fmt.Sprintf("[%s](%s)\n%s\n%s\n\nRequested: <@!%s>\n",
		track.Info.Title, *track.Info.URI, track.Info.Author, musicbot.PlayerBar(player), userData.Requester)

	if tracks, ok := cmd.PlayerManager.Queue(*event.GuildID()); ok {
		content += fmt.Sprintf("\n**Up next:** `%d tracks`", len(tracks))
		limit := min(10, len(tracks))
		for i, track := range tracks[:limit] {
			var Playtime string
			if track.Info.IsStream {
				Playtime = "`LIVE`"
			} else {
				Playtime = musicbot.FormatTime(track.Info.Length)
			}
			content += fmt.Sprintf("\n%d. [%s](%s) `%s`",
				i+1, track.Info.Title, *track.Info.URI, Playtime)

			if track.Info.SourceName == "deezer" || track.Info.SourceName == "spotify" {
				content += " " + track.Info.Author
			}
		}
	}

	embed := discord.NewEmbedBuilder().
		SetTitle("Queue").
		SetDescription(content).
		SetThumbnail(*track.Info.ArtworkURL)

	return event.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed.Build()},
	})
}
