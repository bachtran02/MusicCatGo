package commands

import (
	"MusicCatGo/musicbot"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (c *Commands) RemoveQueueTrackAutocomplete(e *handler.AutocompleteEvent) error {

	queue, ok := c.PlayerManager.Queue(*e.GuildID())
	if !ok || len(queue) == 0 {
		return e.AutocompleteResult(nil)
	}

	choices := make([]discord.AutocompleteChoice, 0)
	limit := min(20, len(queue)) // can only remove one of next 20 tracks
	for i, track := range queue[:limit] {
		choices = append(choices, discord.AutocompleteChoiceInt{
			Name:  fmt.Sprintf("%s - %s", track.Info.Title, track.Info.Author),
			Value: i,
		})
	}
	return e.AutocompleteResult(choices)
}

func (c *Commands) RemoveQueueTrack(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	removeIndex := data.Int("track")

	queue, ok := c.PlayerManager.Queue(*e.GuildID())
	if !ok || len(queue) == 0 || removeIndex < 0 || removeIndex >= len(queue) {
		if sendErr := e.CreateMessage(discord.MessageCreate{
			Content: "Invalid index or no track to remove.",
			Flags:   discord.MessageFlagEphemeral,
		}); sendErr != nil {
			musicbot.LogSendError(sendErr, e.GuildID().String(), e.User().ID.String(), true)
		}
		return nil
	}

	track, ok := c.PlayerManager.RemoveTrack(*e.GuildID(), removeIndex)
	if !ok {
		if sendErr := e.CreateMessage(discord.MessageCreate{
			Content: "Failed to remove track.",
			Flags:   discord.MessageFlagEphemeral,
		}); sendErr != nil {
			musicbot.LogSendError(sendErr, e.GuildID().String(), e.User().ID.String(), true)
		}
		return nil
	}

	if sendErr := e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: fmt.Sprintf(
			"Track [%s](%s) removed from queue.", track.Info.Title, *track.Info.URI)}},
	}); sendErr != nil {
		musicbot.LogSendError(sendErr, e.GuildID().String(), e.User().ID.String(), false)
	}
	musicbot.AutoRemove(e)
	return nil
}
