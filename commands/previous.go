package commands

import (
	"MusicCatGo/musicbot"
	"context"
	"time"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func Previous(c *disgolink.Client, playerManager *musicbot.PlayerManager, ctx context.Context, guildId snowflake.ID) error {

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var (
		player        = (*c).ExistingPlayer(guildId)
		prevTrack, ok = playerManager.Previous(guildId)
		updateOpt     lavalink.PlayerUpdateOpt
	)

	if ok && player.Position() < lavalink.Second*10 {
		updateOpt = lavalink.WithTrack(prevTrack)
	} else {
		updateOpt = lavalink.WithPosition(0)
	}

	if err := player.Update(ctx, updateOpt); err != nil {
		return err
	}
	return nil
}
