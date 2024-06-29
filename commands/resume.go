package commands

import (
	"MusicCatGo/utils"
	"context"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func Resume(c disgolink.Client, ctx context.Context, guildId snowflake.ID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	player := c.ExistingPlayer(guildId)

	if err := player.Update(ctx, lavalink.WithPaused(false)); err != nil {
		return err
	}
	return nil
}

func (c *Commands) Resume(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	if err := Resume(c.Lavalink, e.Ctx, *e.GuildID()); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to resume player",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: "▶️ Resumed player"}},
	})

	utils.AutoRemove(e)
	return nil
}
