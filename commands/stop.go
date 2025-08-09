package commands

import (
	"MusicCatGo/musicbot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (c *Commands) Stop(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	if err := c.PlayerManager.Stop(&c.Lavalink, e.Ctx, *e.GuildID()); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to stop player",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: "⏹️ Stopped playing"}},
	})
	musicbot.AutoRemove(e)
	return nil
}
