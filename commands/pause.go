package commands

import (
	"MusicCatGo/musicbot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (c *Commands) Pause(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	if err := c.PlayerManager.Pause(&c.Lavalink, e.Ctx, *e.GuildID()); err != nil {
		if sendErr := e.CreateMessage(discord.MessageCreate{
			Content: "Failed to pause player",
			Flags:   discord.MessageFlagEphemeral,
		}); sendErr != nil {
			musicbot.LogSendError(sendErr, e.GuildID().String(), e.User().ID.String(), true)
		}
		return err
	}

	if sendErr := e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: "⏸️ Paused player"}},
	}); sendErr != nil {
		musicbot.LogSendError(sendErr, e.GuildID().String(), e.User().ID.String(), false)
	}
	musicbot.AutoRemove(e)
	return nil
}
