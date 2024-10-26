package commands

import (
	"MusicCatGo/utils"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavasrc-plugin"
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
					Name:        "playlist_name",
					Description: "Playlist name",
					Required:    true,
				},
			}},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "delete",
			Description: "Delete playlist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:         "playlist",
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
				discord.ApplicationCommandOptionInt{
					Name:         "playlist",
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
		discord.ApplicationCommandOptionSubCommand{
			Name:        "remove",
			Description: "Remove track from playlist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:         "playlist",
					Description:  "Playlist to remove track from",
					Required:     true,
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionInt{
					Name:         "track",
					Description:  "Track to remove",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
	}}

func (c *Commands) PlaylistAutocomplete(e *handler.AutocompleteEvent) error {
	var (
		limit = 10
		query = e.Data.String("playlist")
	)

	playlists, err := c.Db.SearchPlaylist(e.Ctx, e.User().ID, query, limit)
	if err != nil {
		return e.AutocompleteResult(nil)
	}

	choices := make([]discord.AutocompleteChoice, 0)
	for _, playlist := range playlists {
		choices = append(choices, discord.AutocompleteChoiceInt{
			Name:  playlist.Name,
			Value: playlist.ID,
		})
	}
	return e.AutocompleteResult(choices)
}

func (c *Commands) PlaylistTrackAutocomplete(e *handler.AutocompleteEvent) error {
	var (
		limit = 10
		query = e.Data.String("playlist")
	)

	playlists, err := c.Db.SearchPlaylist(e.Ctx, e.User().ID, query, limit)
	if err != nil || len(playlists) == 0 {
		return e.AutocompleteResult(nil)
	}

	// Get ID of top matched playlist
	playlistId := playlists[0].ID

	// Fetch playlist info
	_, playlistTracks, err := c.Db.GetPlaylist(e.Ctx, playlistId)
	if err != nil || len(playlistTracks) == 0 {
		return e.AutocompleteResult(nil)
	}

	choices := make([]discord.AutocompleteChoice, 0)
	for _, playlistTrack := range playlistTracks {
		choices = append(choices, discord.AutocompleteChoiceInt{
			Name:  playlistTrack.TrackTitle,
			Value: playlistTrack.ID,
		})
	}
	return e.AutocompleteResult(choices)
}

func (c *Commands) AddToPlaylistAutocomplete(e *handler.AutocompleteEvent) error {

	focusedOption := e.Data.Focused()
	switch focusedOption.Name {
	case "playlist":
		return c.PlaylistAutocomplete(e)
	case "query":
		return c.SearchAutocomplete(e)
	}
	return nil
}

func (c *Commands) RemoveFromPlaylistAutocomplete(e *handler.AutocompleteEvent) error {

	focusedOption := e.Data.Focused()
	switch focusedOption.Name {
	case "playlist":
		return c.PlaylistAutocomplete(e)
	case "track":
		return c.PlaylistTrackAutocomplete(e)
	}
	return nil
}

func (c *Commands) CreatePlaylist(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	playlistName := data.String("playlist_name")

	err := c.Db.CreatePlaylist(e.Ctx, e.User().ID, e.User().Username, playlistName)
	if err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Failed to create playlist: `%s`", err),
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: fmt.Sprintf("📋 Playlist `%s` created", playlistName)}},
	})
	utils.AutoRemove(e)
	return nil
}

func (c *Commands) DeletePlaylist(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	playlistId := data.Int("playlist")

	err := c.Db.DeletePlaylist(e.Ctx, playlistId)
	if err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to remove playlist",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{Description: "📋 Playlist deleted"}},
	})
	utils.AutoRemove(e)
	return nil
}

func (c *Commands) ListPlaylists(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {

	// TODO: don't hardcode limit
	playlists, err := c.Db.SearchPlaylist(e.Ctx, e.User().ID, "", 10)
	if err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: err.Error(),
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	if len(playlists) == 0 {
		return e.CreateMessage(discord.MessageCreate{
			Content: "You don't have any playlist.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	content := fmt.Sprintf("<@%s>'s playlists\n", e.User().ID)
	for _, playlist := range playlists {
		content += fmt.Sprintf(
			"- `%s` <t:%d:R>\n", playlist.Name, playlist.CreatedAt.Unix())
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
		playlistID = data.Int("playlist")
		query      = data.String("query")
	)

	if !urlPattern.MatchString(query) {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Please enter a valid URL or use search autocomplete to add to playlist.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	result, err := c.Lavalink.BestNode().LoadTracks(e.Ctx, query)
	if err != nil {
		slog.Error("failed to load tracks", slog.Any("err", err))
		return err
	}

	playlist, _, err := c.Db.GetPlaylist(e.Ctx, playlistID)
	if err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: err.Error(),
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	switch loadData := result.Data.(type) {
	case lavalink.Track:
		err := c.Db.AddTracksToPlaylist(e.Ctx, playlistID, e.User().ID, []lavalink.Track{loadData})
		if err != nil {
			return err
		}
		e.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{{
				Description: fmt.Sprintf("[%s](%s) added to playlist `%s`",
					loadData.Info.Title, *loadData.Info.URI, playlist.Name)}},
		})

	case lavalink.Playlist:
		err := c.Db.AddTracksToPlaylist(e.Ctx, playlistID, e.User().ID, loadData.Tracks)
		if err != nil {
			return err
		}

		var (
			playlistInfo lavasrc.PlaylistInfo
			description  string
		)

		err = loadData.PluginInfo.Unmarshal(&playlistInfo)
		if err != nil {
			description = fmt.Sprintf("Playlist %s `%d tracks` added to playlist `%s`",
				loadData.Info.Name, len(loadData.Tracks), playlist.Name)
		} else {
			description = fmt.Sprintf("Playlist [%s](%s) `%d tracks` added to playlist `%s`",
				loadData.Info.Name, playlistInfo.URL, len(loadData.Tracks), playlist.Name)
		}

		e.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{{Description: description}},
		})
	}

	utils.AutoRemove(e)
	return nil
}

func (c *Commands) RemoveFromPlaylist(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	var (
		trackId = data.Int("track")
	)

	err := c.Db.RemoveTrackFromPlaylist(e.Ctx, trackId, e.User().ID)
	if err != nil {
		return err
	}
	e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{{
			Description: "Track removed from playlist"}},
	})
	utils.AutoRemove(e)
	return nil
}
