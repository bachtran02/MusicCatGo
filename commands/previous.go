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
		player                             = (*c).ExistingPlayer(guildId)
		updateOpt lavalink.PlayerUpdateOpt = lavalink.WithPosition(0)
	)

	/* if past 10 seconds of current track then restart,
	else move back to previous track */
	if player.Position() < lavalink.Second*10 {
		if prevTrack, ok := playerManager.Previous(guildId); ok {
			updateOpt = lavalink.WithTrack(prevTrack)
		}
	}

	if err := player.Update(ctx, updateOpt); err != nil {
		return err
	}
	return nil
}
