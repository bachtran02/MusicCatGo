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

	if !c.PlayerManager.IsPlaying(*e.GuildID()) {
		if sendErr := e.CreateMessage(discord.MessageCreate{
			Content: "Player is not playing.",
			Flags:   discord.MessageFlagEphemeral,
		}); sendErr != nil {
			musicbot.LogSendError(sendErr, e.GuildID().String(), e.User().ID.String(), true)
		}
		return nil
	}

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

	if sendErr := e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: body}},
	}); sendErr != nil {
		musicbot.LogSendError(sendErr, e.GuildID().String(), e.User().ID.String(), false)
	}

	musicbot.AutoRemove(e)
	return nil
}
