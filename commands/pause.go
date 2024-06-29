package commands

import (
	"context"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func Pause(c disgolink.Client, ctx context.Context, guildId snowflake.ID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	player := c.ExistingPlayer(guildId)

	if err := player.Update(ctx, lavalink.WithPaused(true)); err != nil {
		return err
	}
	return nil
}

func (c *Commands) Pause(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	if err := Pause(c.Lavalink, e.Ctx, *e.GuildID()); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to pause player",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: "⏸️ Paused player"}},
	})
	AutoRemove(e)
	return nil
}
