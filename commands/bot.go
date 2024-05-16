package commands

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var bot = discord.SlashCommandCreate{
	Name:        "bot",
	Description: "bot commands",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{
			Name:        "ping",
			Description: "[test] Ping command",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "join",
			Description: "Joins voice chat channel",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "leave",
			Description: "Leaves voice chat channel",
		},
	}}

func (c *Commands) Connect(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	voiceState, ok := c.Client.Caches().VoiceState(*e.GuildID(), e.User().ID)

	if !ok {
		return e.CreateMessage(discord.MessageCreate{
			Content: "You need to be in a voice channel to use this command.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	if err := c.Client.UpdateVoiceState(context.Background(), *e.GuildID(), voiceState.ChannelID, false, true); err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Failed to join voice channel: %s", err),
		})
		return err
	}

	if err := e.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprint("Joined <#", voiceState.ChannelID, ">"),
	}); err != nil {
		return err
	}

	AutoRemove(e)
	return nil
}

func (c *Commands) Disconnect(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	if err := c.Client.UpdateVoiceState(context.Background(), *e.GuildID(), nil, false, true); err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Failed to leave voice channel: %s", err),
		})
		return err
	}

	if err := e.CreateMessage(discord.MessageCreate{
		Content: "Left voice channel!",
	}); err != nil {
		return err
	}
	AutoRemove(e)
	return nil
}

func (c *Commands) Ping(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return e.CreateMessage(discord.MessageCreate{
		Content: "Pong!",
	})
}
