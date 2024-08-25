package commands

import (
	"MusicCatGo/musicbot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

func Loop(playerManager *musicbot.PlayerManager, guildId snowflake.ID, loopMode musicbot.LoopMode) error {

	if state, ok := playerManager.GetState(guildId); ok {
		state.SetLoop(loopMode)
	}
	return nil
}

func (c *Commands) Loop(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return nil
}
