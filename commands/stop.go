package commands

import (
	"MusicCatGo/utils"
	"context"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func (c *Commands) Stop(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	ctx, cancel := context.WithTimeout(e.Ctx, 10*time.Second)
	defer cancel()
	player := c.Lavalink.ExistingPlayer(*e.GuildID())
	if err := player.Update(ctx, lavalink.WithNullTrack()); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to stop player",
		})
	}

	c.PlayerManager.Delete(*e.GuildID())

	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: "⏹️ Stopped playing"}},
	})
	utils.AutoRemove(e)
	return nil
}
