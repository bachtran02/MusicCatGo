package commands

import (
	"MusicCatGo/musicbot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (c *Commands) Shuffle(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	var body string

	if !c.PlayerManager.IsPlaying(*e.GuildID()) {
		if sendErr := e.CreateMessage(discord.MessageCreate{
			Content: "Player is not playing.",
			Flags:   discord.MessageFlagEphemeral,
		}); sendErr != nil {
			musicbot.LogSendError(sendErr, e.GuildID().String(), e.User().ID.String(), true)
		}
		return nil
	}

	if state, ok := c.PlayerManager.GetState(*e.GuildID()); ok {
		if state.Shuffle() {
			c.PlayerManager.SetShuffle(*e.GuildID(), musicbot.ShuffleOff)
			body = "🔀 Shuffle off"
		} else {
			c.PlayerManager.SetShuffle(*e.GuildID(), musicbot.ShuffleOn)
			body = "🔀 Shuffle on"
		}
	}

	if sendErr := e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: body}},
	}); sendErr != nil {
		musicbot.LogSendError(sendErr, e.GuildID().String(), e.User().ID.String(), false)
	}
	musicbot.AutoRemove(e)
	return nil
}
