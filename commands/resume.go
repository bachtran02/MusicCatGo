package commands

import (
	"MusicCatGo/musicbot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (c *Commands) Resume(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	if err := c.PlayerManager.Resume(&c.Lavalink, e.Ctx, *e.GuildID()); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to resume player",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: "▶️ Resumed player"}},
	})

	musicbot.AutoRemove(e)
	return nil
}
