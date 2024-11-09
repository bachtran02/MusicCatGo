package commands

import (
	"MusicCatGo/musicbot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

func Shuffle(playerManager *musicbot.PlayerManager, guildId snowflake.ID, shuffleMode musicbot.ShuffleMode) error {

	if state, ok := playerManager.GetState(guildId); ok {
		state.SetShuffle(shuffleMode)
	}
	return nil
}

func (c *Commands) Shuffle(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	var body string

	if state, ok := c.PlayerManager.GetState(*e.GuildID()); ok {
		if state.Shuffle() {
			state.SetShuffle(musicbot.ShuffleOff)
			body = "🔀 Shuffle off"
		} else {
			state.SetShuffle(musicbot.ShuffleOn)
			body = "🔀 Shuffle on"
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
