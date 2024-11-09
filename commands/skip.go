package commands

import (
	"MusicCatGo/musicbot"
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func Skip(c *disgolink.Client, playerManager *musicbot.PlayerManager, ctx context.Context, guildId snowflake.ID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var (
		player        = (*c).ExistingPlayer(guildId)
		nextTrack, ok = playerManager.Next(guildId)
		updateOpt     lavalink.PlayerUpdateOpt
	)

	if !ok {
		updateOpt = lavalink.WithNullTrack()
	} else {
		updateOpt = lavalink.WithTrack(nextTrack)
	}

	if err := player.Update(ctx, updateOpt); err != nil {
		return err
	}
	return nil
}

func (c *Commands) Skip(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	curTrack := c.Lavalink.ExistingPlayer(*e.GuildID()).Track()

	if err := Skip(&c.Lavalink, &c.PlayerManager, e.Ctx, *e.GuildID()); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to skip track",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: fmt.Sprintf(
			"⏭️ Track skipped: [%s](%s)", curTrack.Info.Title, *curTrack.Info.URI)}},
	})
	musicbot.AutoRemove(e)
	return nil
}
