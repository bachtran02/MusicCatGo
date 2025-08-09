package commands

import (
	"MusicCatGo/musicbot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (c *Commands) Shuffle(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	var body string

	if state, ok := c.PlayerManager.GetState(*e.GuildID()); ok {
		if state.Shuffle() {
			c.PlayerManager.SetShuffle(*e.GuildID(), musicbot.ShuffleOff)
			body = "🔀 Shuffle off"
		} else {
			c.PlayerManager.SetShuffle(*e.GuildID(), musicbot.ShuffleOn)
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
