package commands

import (
	"MusicCatGo/utils"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

var playlist = discord.SlashCommandCreate{
	Name:        "playlist",
	Description: "playlist commands",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{
			Name:        "create",
			Description: "Create new playlist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "name",
					Description: "Playlist name",
					Required:    true,
				},
			}},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "remove",
			Description: "Remove playlist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:         "name",
					Description:  "Playlist name",
					Required:     true,
					Autocomplete: true,
				},
			}},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "list",
			Description: "List playlist",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "add",
			Description: "Add track(s) to playlist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:         "query",
					Description:  "Search query",
					Required:     true,
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionString{
					Name:         "playlist_name",
					Description:  "Playlist name",
					Required:     true,
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "source",
					Description: "Source to search from",
					Required:    false,
					Choices:     searchSourceChoices,
				},
				discord.ApplicationCommandOptionString{
					Name:        "type",
					Description: "Type of search",
					Required:    false,
					Choices:     searchTypeChoices,
				},
			}},
	}}

func (c *Commands) PlaylistAutocomplete(e *handler.AutocompleteEvent) error {
	var (
		limit = 10
		query = e.Data.String("playlist_name")
	)

	playlists, err := c.Db.SearchPlaylist(e.Ctx, e.User().ID, query, limit)
	if err != nil {
		return e.AutocompleteResult(nil)
	}

	choices := make([]discord.AutocompleteChoice, 0)
	for _, playlist := range playlists {
		choices = append(choices, discord.AutocompleteChoiceString{
			Name:  playlist.Name,
			Value: playlist.Name,
		})
	}
	return e.AutocompleteResult(choices)
}

func (c *Commands) AddToPlaylistAutocomplete(e *handler.AutocompleteEvent) error {

	focusedOption := e.Data.Focused()
	switch focusedOption.Name {
	case "playlist_name":
		return c.PlaylistAutocomplete(e)
	case "query":
		return c.SearchAutocomplete(e)
	}
	return nil
}

func (c *Commands) CreatePlaylist(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	playlistName := data.String("name")

	err := c.Db.CreatePlaylist(e.Ctx, e.User().ID, e.User().Username, playlistName)
	if err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Failed to create playlist `%s`", playlistName),
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: fmt.Sprintf("📋 Playlist `%s` created", playlistName)}},
	})
	utils.AutoRemove(e)
	return nil
}

func (c *Commands) RemovePlaylist(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	playlistName := data.String("name")

	err := c.Db.RemovePlaylist(e.Ctx, e.User().ID, playlistName)
	if err != nil {
		slog.Error("failed to remove playlist", slog.Any("err", err))
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Failed to remove playlist `%s`.", playlistName),
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: fmt.Sprintf("📋 Playlist `%s` deleted", playlistName)}},
	})
	utils.AutoRemove(e)
	return nil
}

func (c *Commands) ListPlaylists(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	// TODO: don't hardcode limit
	playlists, err := c.Db.SearchPlaylist(e.Ctx, e.User().ID, "", 10)
	if err != nil {
		slog.Error("failed to fetch playlists", slog.Any("err", err))
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to fetch playlists",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	if len(playlists) == 0 {
		return e.CreateMessage(discord.MessageCreate{
			Content: "You don't have any playlist.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	var content string
	for _, playlist := range playlists {
		content += fmt.Sprintf(
			"- %s `%d track(s)` <t:%d:R>", playlist.Name, 5, playlist.CreatedAt.Unix())
	}

	embed := discord.NewEmbedBuilder().
		SetTitle("Playlists").
		SetDescription(content)

	return e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed.Build()},
	})
}

func (c *Commands) AddToPlaylist(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	var (
		playlistID   int
		query        = data.String("query")
		playlistName = data.String("playlist_name")
	)

	if !urlPattern.MatchString(query) {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Please enter a valid URL or use search autocomplete to add to playlist.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	// TODO: if track exists in track database then no need to search
	result, err := c.Lavalink.BestNode().LoadTracks(e.Ctx, query)
	if err != nil {
		slog.Error("failed to load tracks", slog.Any("err", err))
		return err
	}

	playlists, err := c.Db.SearchPlaylist(e.Ctx, e.User().ID, playlistName, 1)
	if err != nil {
		return err
	}

	if len(playlists) == 0 {
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Playlist `%s` does not exist.", playlistName),
			Flags:   discord.MessageFlagEphemeral,
		})
	} else {
		playlistID = playlists[0].ID
	}

	switch loadData := result.Data.(type) {
	case lavalink.Track:
		err := c.Db.AddTracksToPlaylist(e.Ctx, playlistID, []lavalink.Track{loadData})
		if err != nil {
			return err
		}
		e.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{{
				Description: fmt.Sprintf("[%s](%s) added playlist `%s`",
					loadData.Info.Title, *loadData.Info.URI, playlistName)}},
		})

	case lavalink.Playlist:
		return nil
	}

	utils.AutoRemove(e)
	return nil
}
