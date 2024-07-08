package commands

import (
	"MusicCatGo/utils"
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func (c *Commands) Skip(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	ctx, cancel := context.WithTimeout(e.Ctx, 10*time.Second)
	defer cancel()

	var (
		player        = c.Lavalink.ExistingPlayer(*e.GuildID())
		currentTrack  = player.Track()
		nextTrack, ok = c.PlayerManager.Next(*e.GuildID())
		updateOpt     lavalink.PlayerUpdateOpt
	)

	if !ok {
		updateOpt = lavalink.WithNullTrack()
	} else {
		updateOpt = lavalink.WithTrack(nextTrack)
	}

	if err := player.Update(ctx, updateOpt); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to skip track",
		})
	}

	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: fmt.Sprintf(
			"⏭️ Track skipped: [%s](%s)", currentTrack.Info.Title, *currentTrack.Info.URI)}},
	})
	utils.AutoRemove(e)
	return nil
}
