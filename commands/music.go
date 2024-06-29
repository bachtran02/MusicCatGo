package commands

import (
	"github.com/disgoorg/disgo/discord"
)

var searchTypeChoices = []discord.ApplicationCommandOptionChoiceString{
	{
		Name:  "Track",
		Value: "track",
	},
	{
		Name:  "Album",
		Value: "album",
	},
	{
		Name:  "Artist",
		Value: "artist",
	},
	{
		Name:  "Playlist",
		Value: "playlist",
	},
}

var searchSourceChoices = []discord.ApplicationCommandOptionChoiceString{
	{
		Name:  "YouTube",
		Value: "youtube",
	},
	{
		Name:  "Deezer",
		Value: "deezer",
	},
	{
		Name:  "Spotify",
		Value: "spotify",
	},
}

var music = discord.SlashCommandCreate{
	Name:        "music",
	Description: "music commands",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{
			Name:        "play",
			Description: "Plays a song from query",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "query",
					Description: "Search query for track",
					Required:    true,
				},
			}},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "search",
			Description: "Add & play track/playlist from search results",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:         "query",
					Description:  "Search query for track",
					Required:     true,
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "source",
					Description: "The source to search from",
					Required:    false,
					Choices:     searchSourceChoices,
				},
				discord.ApplicationCommandOptionString{
					Name:        "type",
					Description: "The type of the search",
					Required:    false,
					Choices:     searchTypeChoices,
				},
			}},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "queue",
			Description: "Display queue",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "resume",
			Description: "Resume player",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "pause",
			Description: "Pause player",
		},
	}}
