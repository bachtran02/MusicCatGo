package commands

import (
	"MusicCatGo/musicbot"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (cmd *Commands) Now(data discord.SlashCommandInteractionData, event *handler.CommandEvent) error {

	player := cmd.Lavalink.Player(*event.GuildID())

	track := player.Track()
	if track == nil {
		if sendErr := event.CreateMessage(discord.MessageCreate{
			Content: "Player is not playing",
			Flags:   discord.MessageFlagEphemeral,
		}); sendErr != nil {
			musicbot.LogSendError(sendErr, event.GuildID().String(), event.User().ID.String(), true)
		}
		return nil
	}

	var userData UserData
	_ = track.UserData.Unmarshal(&userData)

	content := fmt.Sprintf("[%s](%s)\n%s\n%s\n\nRequested: <@!%s>\n",
		track.Info.Title, *track.Info.URI, track.Info.Author, musicbot.PlayerBar(player), userData.Requester)

	if tracks, ok := cmd.PlayerManager.Queue(*event.GuildID()); ok {
		content += "\n**Up next:**"
		for _, track := range tracks[:1] {
			var Playtime string
			if track.Info.IsStream {
				Playtime = "`LIVE`"
			} else {
				Playtime = musicbot.FormatTime(track.Info.Length)
			}
			content += fmt.Sprintf("\n[%s](%s) `%s`",
				track.Info.Title, *track.Info.URI, Playtime)

			if track.Info.SourceName == "deezer" || track.Info.SourceName == "spotify" {
				content += " " + track.Info.Author
			}
		}
	}

	embed := discord.NewEmbedBuilder().
		SetTitle("Current").
		SetDescription(content).
		SetThumbnail(*track.Info.ArtworkURL)

	if sendErr := event.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed.Build()},
	}); sendErr != nil {
		musicbot.LogSendError(sendErr, event.GuildID().String(), event.User().ID.String(), false)
	}
	return nil
}
