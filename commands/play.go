package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/json"
	"github.com/disgoorg/lavasearch-plugin"
	"github.com/disgoorg/lavasrc-plugin"
	"github.com/disgoorg/snowflake/v2"
)

var (
	urlPattern = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
)

type UserData struct {
	Requester    snowflake.ID `json:"requester"`
	PlaylistName string       `json:"playlistName"`
	PlaylistURL  string       `json:"playlistUrl"`
}

func (c *Commands) SearchAutocomplete(e *handler.AutocompleteEvent) error {
	query := e.Data.String("query")
	if query == "" {
		return e.AutocompleteResult(nil)
	}

	choices := make([]discord.AutocompleteChoice, 0)

	source := lavalink.SearchType(e.Data.String("source"))
	t, typeOK := e.Data.OptString("type")

	if typeOK || source == "deezer" || source == "spotify" {

		if source != "deezer" {
			source = "spsearch"
		} else {
			source = "dzsearch"
		}
		query = source.Apply(query)

		var (
			searchType []lavasearch.SearchType
			numChoices int
		)
		if t == "" {
			numChoices = 5
			searchType = []lavasearch.SearchType{
				lavasearch.SearchTypeTrack,
				lavasearch.SearchTypeArtist,
				lavasearch.SearchTypeAlbum,
				lavasearch.SearchTypePlaylist,
			}
		} else {
			numChoices = 20
			searchType = []lavasearch.SearchType{
				lavasearch.SearchType(t),
			}
		}

		result, err := lavasearch.LoadSearch(c.Lavalink.BestNode().Rest(), query, searchType)
		if err != nil {
			if errors.Is(err, lavasearch.ErrEmptySearchResult) {
				return e.AutocompleteResult(nil)
			}
			return e.AutocompleteResult([]discord.AutocompleteChoice{
				discord.AutocompleteChoiceString{
					Name:  "Failed to load search results",
					Value: "error",
				},
			})
		}

		for _, track := range result.Tracks[:min(len(result.Tracks), numChoices)] {

			var trackInfo lavasrc.PlaylistInfo
			_ = track.PluginInfo.Unmarshal(&trackInfo)

			choices = append(choices, discord.AutocompleteChoiceString{
				Name:  fmt.Sprintf("🎵 %s - %s", track.Info.Title, track.Info.Author),
				Value: *track.Info.URI,
			})
		}

		for _, artist := range result.Artists[:min(len(result.Artists), numChoices)] {

			var artistInfo lavasrc.PlaylistInfo
			_ = artist.PluginInfo.Unmarshal(&artistInfo)

			choices = append(choices, discord.AutocompleteChoiceString{
				Name:  fmt.Sprintf("🎤 %s", artistInfo.Author),
				Value: artistInfo.URL,
			})
		}

		for _, playlist := range result.Playlists[:min(len(result.Playlists), numChoices)] {

			var playlistInfo lavasrc.PlaylistInfo
			_ = playlist.PluginInfo.Unmarshal(&playlistInfo)

			choices = append(choices, discord.AutocompleteChoiceString{
				Name:  fmt.Sprintf("🎧 %s - %s ⭐", playlist.Info.Name, playlistInfo.Author),
				Value: playlistInfo.URL,
			})
		}

		for _, album := range result.Albums[:min(len(result.Albums), numChoices)] {

			var albumInfo lavasrc.PlaylistInfo
			_ = album.PluginInfo.Unmarshal(&albumInfo)

			choices = append(choices, discord.AutocompleteChoiceString{
				Name:  fmt.Sprintf("💿 %s - %s 🎤", album.Info.Name, albumInfo.Author),
				Value: albumInfo.URL,
			})
		}
		return e.AutocompleteResult(choices)
	}

	query = lavalink.SearchTypeYouTube.Apply(query)

	ctx, cancel := context.WithTimeout(e.Ctx, 10*time.Second)
	defer cancel()
	result, err := c.Lavalink.BestNode().LoadTracks(ctx, query)
	if err == nil {
		if tracks, ok := result.Data.(lavalink.Search); ok {
			for _, track := range tracks[:min(len(tracks), 20)] {
				choices = append(choices, discord.AutocompleteChoiceString{
					Name:  fmt.Sprintf("🎬 %s [%s]", Trim(track.Info.Title, 60), Trim(track.Info.Author, 20)),
					Value: *track.Info.URI,
				})
			}

			return e.AutocompleteResult(choices)
		}
	}

	return e.AutocompleteResult(nil)
}

func _Play(query string, e *handler.CommandEvent, c *Commands) error {

	if !urlPattern.MatchString(query) {
		query = lavalink.SearchTypeYouTube.Apply(query)
	}

	ctx, cancel := context.WithTimeout(e.Ctx, 10*time.Second)
	defer cancel()
	result, err := c.Lavalink.BestNode().LoadTracks(ctx, query)
	if err != nil {
		slog.Any("error", err)
		return err
	}

	var (
		tracks       []lavalink.Track
		userData     = UserData{Requester: e.User().ID}
		embedBuilder discord.EmbedBuilder
	)

	switch loadData := result.Data.(type) {
	case lavalink.Track, lavalink.Search:
		var (
			track    lavalink.Track
			playtime string
		)
		if t, ok := loadData.(lavalink.Track); ok {
			track, tracks = t, append(tracks, t)
		} else if t, ok := loadData.(lavalink.Search); ok {
			track, tracks = t[0], append(tracks, t[0])
		}

		if track.Info.IsStream {
			playtime = "LIVE"
		} else {
			playtime = FormatTime(track.Info.Length)
		}

		embedBuilder = *discord.NewEmbedBuilder().
			SetTitle("Track added").
			SetDescription(fmt.Sprintf("[%s](%s)\n%s `%s`\n\n<@%s>",
				track.Info.Title, *track.Info.URI, track.Info.Author,
				playtime, userData.Requester)).
			SetThumbnail(*track.Info.ArtworkURL)

	case lavalink.Playlist:
		var (
			description  string
			lavasrcInfo  lavasrc.PlaylistInfo
			thumbnailUrl = ""
			playlistType = "playlist"
			numTracks    = len(loadData.Tracks)
		)

		tracks = append(tracks, loadData.Tracks...)
		userData.PlaylistName = loadData.Info.Name
		userData.PlaylistURL = query

		var _ = loadData.PluginInfo.Unmarshal(&lavasrcInfo)

		if lavasrcInfo.Type == "" {
			description = fmt.Sprintf("[%s](%s) - %d tracks\n\n<@%s>",
				loadData.Info.Name, userData.PlaylistURL, numTracks, userData.Requester)
		} else {
			playlistType = string(lavasrcInfo.Type)
			thumbnailUrl = lavasrcInfo.ArtworkURL
			switch lavasrcInfo.Type {
			case lavasrc.PlaylistTypeArtist:
				description = fmt.Sprintf("[%s](%s) - `%d tracks`\n\n<@%s>",
					lavasrcInfo.Author, lavasrcInfo.URL, numTracks, userData.Requester)
			case lavasrc.PlaylistTypePlaylist, lavasrc.PlaylistTypeAlbum:
				description = fmt.Sprintf("[%s](%s) `%d track(s)`\n%s\n\n<@%s>",
					loadData.Info.Name, lavasrcInfo.URL, numTracks, lavasrcInfo.Author, userData.Requester)
			}
		}

		embedBuilder = *discord.NewEmbedBuilder().
			SetTitle(strings.ToUpper(string(playlistType[0])) + playlistType[1:] + " added").
			SetDescription(description).
			SetThumbnail(thumbnailUrl)

	case lavalink.Empty:
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr("No matches found"),
		})
		return err
	case lavalink.Exception:
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(fmt.Sprintf("Failed to load tracks: %s", loadData.Error())),
		})
		return err
	}

	if _, err = e.UpdateInteractionResponse(discord.MessageUpdate{
		Embeds: &[]discord.Embed{embedBuilder.Build()},
	}); err != nil {
		return err
	}

	voiceState, _ := c.Client.Caches().VoiceState(*e.GuildID(), e.User().ID)
	if err = c.Client.UpdateVoiceState(context.Background(), *e.GuildID(), voiceState.ChannelID, false, true); err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Failed to join voice channel: %s", err),
		})
		return err
	}

	userDataRaw, _ := json.Marshal(userData)
	for i := range tracks {
		tracks[i].UserData = userDataRaw
	}

	player := c.Lavalink.Player(*e.GuildID())
	if player.Track() == nil {
		var track lavalink.Track
		if len(tracks) == 1 {
			track = tracks[0]
			tracks = nil
		} else {
			track, tracks = tracks[0], tracks[1:]
		}
		playCtx, playCancel := context.WithTimeout(e.Ctx, 10*time.Second)
		defer playCancel()
		if err = player.Update(playCtx, lavalink.WithTrack(track)); err != nil {
			_, err = e.CreateFollowupMessage(discord.MessageCreate{
				Content: fmt.Sprintf("Failed to play track: %s", err),
			})
			return err
		}
	}

	if len(tracks) > 0 {
		c.PlayerManager.Add(*e.GuildID(), e.Channel().ID(), tracks...)
	}

	AutoRemove(e)
	return nil
}

func (cmd *Commands) Play(data discord.SlashCommandInteractionData, event *handler.CommandEvent) error {

	_, ok := cmd.Client.Caches().VoiceState(*event.GuildID(), event.User().ID)
	if !ok {
		return event.CreateMessage(discord.MessageCreate{
			Content: "You need to be in a voice channel to use this command.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}

	return _Play(data.String("query"), event, cmd)
}
