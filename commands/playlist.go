package commands

import (
	"github.com/disgoorg/disgo/discord"
)

var playlist = discord.SlashCommandCreate{
	Name:        "playlist",
	Description: "playlist commands",
	Options:     []discord.ApplicationCommandOption{}}
