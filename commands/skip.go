package commands

import (
	"MusicCatGo/musicbot"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (c *Commands) Skip(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	curTrack := c.Lavalink.ExistingPlayer(*e.GuildID()).Track()

	if err := c.PlayerManager.Skip(&c.Lavalink, e.Ctx, *e.GuildID()); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to skip track",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: fmt.Sprintf(
			"⏭️ Track skipped: [%s](%s)", curTrack.Info.Title, *curTrack.Info.URI)}},
	})
	musicbot.AutoRemove(e)
	return nil
}
