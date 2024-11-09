package commands

import (
	"MusicCatGo/musicbot"
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func (c *Commands) Seek(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	ctx, cancel := context.WithTimeout(e.Ctx, 10*time.Second)
	defer cancel()
	player := c.Lavalink.ExistingPlayer(*e.GuildID())

	if player.Track().Info.IsStream {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Track is unseekable",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	position := data.String("position")
	h, m, s, err := musicbot.ParseTime(position)
	if err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Invalid position",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	newPosition := lavalink.Duration(s*int(lavalink.Second) + m*int(lavalink.Minute) + h*int(lavalink.Hour))
	if err := player.Update(ctx, lavalink.WithPosition(newPosition)); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to seek to position",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: fmt.Sprintf(
			"⏩ Player moved to `%s`", musicbot.FormatTime(newPosition))}},
	})
	musicbot.AutoRemove(e)
	return nil
}
