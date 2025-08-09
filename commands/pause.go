package commands

import (
	"MusicCatGo/musicbot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (c *Commands) Pause(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	if err := c.PlayerManager.Pause(&c.Lavalink, e.Ctx, *e.GuildID()); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to pause player",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: "⏸️ Paused player"}},
	})
	musicbot.AutoRemove(e)
	return nil
}
