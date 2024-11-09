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

func (c *Commands) Loop(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	var (
		body string
		mode string = data.String("mode")
	)

	if state, ok := c.PlayerManager.GetState(*e.GuildID()); ok {
		if mode == string(musicbot.LoopNone) {
			state.SetLoop(musicbot.LoopNone)
			body = "⏭️ Disable loop"
		} else if mode == string(musicbot.LoopTrack) {
			state.SetLoop(musicbot.LoopTrack)
			body = "🔂 Enabled Track loop"
		} else {
			state.SetLoop(musicbot.LoopQueue)
			body = "🔁 Enabled Queue loop"
		}
	} else {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Player is not playing",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: body}},
	})

	musicbot.AutoRemove(e)
	return nil
}
