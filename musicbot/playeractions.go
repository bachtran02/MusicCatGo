package musicbot

import (
	"context"
	"time"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func (pm *PlayerManager) Pause(c *disgolink.Client, ctx context.Context, guildId snowflake.ID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	player := (*c).ExistingPlayer(guildId)

	if err := player.Update(ctx, lavalink.WithPaused(true)); err != nil {
		return err
	}

	if state, ok := pm.GetState(guildId); ok {
		state.SetPause(true)
	}
	return nil
}

func (pm *PlayerManager) Resume(c *disgolink.Client, ctx context.Context, guildId snowflake.ID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	player := (*c).ExistingPlayer(guildId)

	if err := player.Update(ctx, lavalink.WithPaused(false)); err != nil {
		return err
	}

	if state, ok := pm.GetState(guildId); ok {
		state.SetPause(false)
	}
	return nil
}

func (pm *PlayerManager) Stop(c *disgolink.Client, ctx context.Context, guildId snowflake.ID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	player := (*c).ExistingPlayer(guildId)
	if err := player.Update(ctx, lavalink.WithNullTrack()); err != nil {
		return err
	}
	pm.Delete(guildId)
	return nil
}

func (pm *PlayerManager) Skip(c *disgolink.Client, ctx context.Context, guildId snowflake.ID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	player := (*c).ExistingPlayer(guildId)
	nextTrack, ok := pm.Next(guildId)
	var updateOpt lavalink.PlayerUpdateOpt

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

func (pm *PlayerManager) SetShuffle(guildId snowflake.ID, shuffleMode ShuffleMode) {
	if state, ok := pm.GetState(guildId); ok {
		state.SetShuffle(shuffleMode)
	}
}

func (pm *PlayerManager) SetLoop(guildId snowflake.ID, loopMode LoopMode) {
	if state, ok := pm.GetState(guildId); ok {
		state.SetLoop(loopMode)
	}
}

func (pm *PlayerManager) PlayPrevious(c *disgolink.Client, ctx context.Context, guildId snowflake.ID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	player := (*c).ExistingPlayer(guildId)
	var updateOpt lavalink.PlayerUpdateOpt = lavalink.WithPosition(0)

	if player.Position() < lavalink.Second*10 {
		if prevTrack, ok := pm.Previous(guildId); ok {
			updateOpt = lavalink.WithTrack(prevTrack)
		}
	}

	if err := player.Update(ctx, updateOpt); err != nil {
		return err
	}
	return nil
}
