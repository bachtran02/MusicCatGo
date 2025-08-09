package commands

import (
	"MusicCatGo/musicbot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (c *Commands) Loop(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	var (
		body string
		mode string = data.String("mode")
	)

	if _, ok := c.PlayerManager.GetState(*e.GuildID()); ok {
		if mode == string(musicbot.LoopNone) {
			c.PlayerManager.SetLoop(*e.GuildID(), musicbot.LoopNone)
			body = "⏭️ Disable loop"
		} else if mode == string(musicbot.LoopTrack) {
			c.PlayerManager.SetLoop(*e.GuildID(), musicbot.LoopTrack)
			body = "🔂 Enabled Track loop"
		} else {
			c.PlayerManager.SetLoop(*e.GuildID(), musicbot.LoopQueue)
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
