package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/json"
	"github.com/disgoorg/lavasearch-plugin"
	"github.com/disgoorg/lavasrc-plugin"
)

var (
	urlPattern = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
)

type TrackSource string

const (
	Spotify TrackSource = "spotify"
	Deezer  TrackSource = "deezer"
	YouTube TrackSource = "youtube"
)

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
		tracks         []lavalink.Track
		messageContent string
	)

	switch loadData := result.Data.(type) {
	case lavalink.Track:
		tracks = append(tracks, loadData)
		messageContent = "Added track to queue"
	case lavalink.Search:
		tracks = append(tracks, loadData[0])
		messageContent = "Added track to queue"
	case lavalink.Playlist:
		tracks = append(tracks, loadData.Tracks...)
		messageContent = "Added tracks to queue"
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
		Content: &messageContent,
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

	player := c.Lavalink.Player(*e.GuildID())
	if player.Track() == nil {
		var track lavalink.Track
		if len(tracks) == 1 {
			track = tracks[0]
			tracks = nil
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
	return nil
}
