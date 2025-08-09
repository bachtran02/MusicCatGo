package commands

import (
	"MusicCatGo/musicbot"
	"context"
	"errors"
	"fmt"
	"math/rand"
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

type SearchType int

const (
	LavalinkSearch SearchType = 0
	PlaylistSearch SearchType = 1
)

type OptBool int

const (
	OptTrue  OptBool = 1
	OptFalse OptBool = 0
	OptUnset OptBool = -1
)

type UserData struct {
	Requester    snowflake.ID `json:"requester"`
	PlaylistName string       `json:"playlistName"`
	PlaylistURL  string       `json:"playlistUrl"`
}

type PlayOpts struct {
	Query    interface{}
	Type     SearchType
	PlayNext OptBool
	Loop     OptBool
	Shuffle  OptBool
}

func optBoolValue(b bool, ok bool) OptBool {
	if !ok {
		return OptUnset
	}
	if b {
		return OptTrue
	}
	return OptFalse
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
					Name:  fmt.Sprintf("🎬 %s [%s]", musicbot.Trim(track.Info.Title, 60), musicbot.Trim(track.Info.Author, 20)),
					Value: *track.Info.URI,
				})
			}

			return e.AutocompleteResult(choices)
		}
	}

	return e.AutocompleteResult(nil)
}

func SearchLavalink(query string, c *Commands, ctx context.Context) (*lavalink.LoadResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if !urlPattern.MatchString(query) {
		query = lavalink.SearchTypeYouTube.Apply(query)
	}

	result, err := c.Lavalink.BestNode().LoadTracks(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to load tracks from Lavalink for query '%s': %w", query, err)
	}

	return result, nil
}

func SearchPlaylist(playlistId int, c *Commands, userId snowflake.ID, ctx context.Context) (*lavalink.LoadResult, error) {

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	dbPlaylist, dbTracks, err := c.Db.GetPlaylist(ctx, playlistId)
	if err != nil {
		return nil, err
	}

	playlist := lavalink.Playlist{
		Info: lavalink.PlaylistInfo{
			Name:          dbPlaylist.Name,
			SelectedTrack: -1,
		},
		Tracks: make([]lavalink.Track, 0),
	}

	for _, track := range dbTracks {
		playlist.Tracks = append(playlist.Tracks, track.Track)
	}

	return &lavalink.LoadResult{
		LoadType: lavalink.LoadTypePlaylist,
		Data:     playlist,
	}, nil
}

func SearchQuery(query interface{}, searchType SearchType, c *Commands, userId snowflake.ID, ctx context.Context) (*lavalink.LoadResult, error) {
	switch searchType {
	case LavalinkSearch:
		q, ok := query.(string)
		if !ok {
			return nil, fmt.Errorf("query should be a string for Lavalink search, got %T", query)
		}
		return SearchLavalink(q, c, ctx)

	case PlaylistSearch:
		q, ok := query.(int)
		if !ok {
			return nil, fmt.Errorf("query should be an int for Playlist search, got %T", query)
		}
		return SearchPlaylist(q, c, userId, ctx)
	default:
		return nil, fmt.Errorf("unknown search type: %v", searchType)
	}
}

func buildTrackEmbed(track lavalink.Track, requester snowflake.ID) discord.Embed {
	var playtime string
	if track.Info.IsStream {
		playtime = "LIVE"
	} else {
		playtime = musicbot.FormatTime(track.Info.Length)
	}

	return discord.NewEmbedBuilder().
		SetTitle("Track added").
		SetDescription(fmt.Sprintf("[%s](%s)\n%s `%s`\n\n<@%s>",
			track.Info.Title, *track.Info.URI, track.Info.Author,
			playtime, requester)).
		SetThumbnail(*track.Info.ArtworkURL).
		Build()
}

func buildPlaylistEmbed(playlist lavalink.Playlist, requester snowflake.ID) discord.Embed {
	var (
		description  string
		lavasrcInfo  lavasrc.PlaylistInfo
		thumbnailUrl = ""
		playlistType = "playlist"
		numTracks    = len(playlist.Tracks)
	)

	var _ = playlist.PluginInfo.Unmarshal(&lavasrcInfo)

	if lavasrcInfo.Type == "" {
		description = fmt.Sprintf("%s - %d tracks\n\n<@%s>",
			playlist.Info.Name, numTracks, requester)
	} else {
		playlistType = string(lavasrcInfo.Type)
		thumbnailUrl = lavasrcInfo.ArtworkURL
		switch lavasrcInfo.Type {
		case lavasrc.PlaylistTypeArtist:
			description = fmt.Sprintf("[%s](%s) - `%d tracks`\n\n<@%s>",
				lavasrcInfo.Author, lavasrcInfo.URL, numTracks, requester)
		case lavasrc.PlaylistTypePlaylist, lavasrc.PlaylistTypeAlbum:
			description = fmt.Sprintf("[%s](%s) `%d track(s)`\n%s\n\n<@%s>",
				playlist.Info.Name, lavasrcInfo.URL, numTracks, lavasrcInfo.Author, requester)
		}
	}

	return discord.NewEmbedBuilder().
		SetTitle(strings.ToUpper(string(playlistType[0])) + playlistType[1:] + " added").
		SetDescription(description).
		SetThumbnail(thumbnailUrl).
		Build()
}

func _Play(playOpts PlayOpts, e *handler.CommandEvent, c *Commands) error {

	// delete response message
	defer musicbot.AutoRemove(e)

	result, err := SearchQuery(playOpts.Query, playOpts.Type, c, e.User().ID, e.Ctx)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: err.Error(),
		})
		return err
	}

	var (
		tracks   []lavalink.Track
		embed    discord.Embed
		userData = UserData{
			Requester: e.User().ID,
		}
		loop = musicbot.LoopNone
	)

	switch loadData := result.Data.(type) {
	case lavalink.Track:
		tracks = append(tracks, loadData)
		embed = buildTrackEmbed(loadData, userData.Requester)
		if playOpts.Loop == OptTrue {
			loop = musicbot.LoopTrack
		}

	case lavalink.Search:
		track := loadData[0]
		tracks = append(tracks, track)
		embed = buildTrackEmbed(track, userData.Requester)
		if playOpts.Loop == OptTrue {
			loop = musicbot.LoopTrack
		}

	case lavalink.Playlist:
		if playOpts.Shuffle != OptFalse {
			rand.Shuffle(len(loadData.Tracks), func(i, j int) {
				loadData.Tracks[i], loadData.Tracks[j] = loadData.Tracks[j], loadData.Tracks[i]
			})
		}
		if playOpts.Loop == OptTrue {
			loop = musicbot.LoopQueue
		}
		tracks = append(tracks, loadData.Tracks...)
		userData.PlaylistName = loadData.Info.Name
		embed = buildPlaylistEmbed(loadData, userData.Requester)

	case lavalink.Empty:
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr("No matches found for search query."),
		})
		return err
	case lavalink.Exception:
		_, err = e.UpdateInteractionResponse(discord.MessageUpdate{
			Content: json.Ptr(fmt.Sprintf("Failed to load tracks: `%s`", loadData.Error())),
		})
		return err
	}

	if _, err = e.UpdateInteractionResponse(discord.MessageUpdate{
		Embeds: &[]discord.Embed{embed},
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

	if playOpts.PlayNext == OptTrue {
		c.PlayerManager.AddNext(*e.GuildID(), e.Channel().ID(), tracks...)
	} else {
		c.PlayerManager.Add(*e.GuildID(), e.Channel().ID(), tracks...)
	}

	state, ok := c.PlayerManager.GetState(*e.GuildID())
	if ok {
		if playOpts.Loop != OptUnset {
			state.SetLoop(loop)
		}
		if playOpts.Shuffle == OptTrue {
			state.SetShuffle(musicbot.ShuffleOn)
		}
		if playOpts.Shuffle == OptFalse {
			state.SetShuffle(musicbot.ShuffleOff)
		}
	}

	player := c.Lavalink.Player(*e.GuildID())
	if player.Track() == nil {
		track, _ := c.PlayerManager.Next(*e.GuildID())
		playCtx, playCancel := context.WithTimeout(e.Ctx, 10*time.Second)
		defer playCancel()
		if err = player.Update(playCtx, lavalink.WithTrack(track)); err != nil {
			_, err = e.CreateFollowupMessage(discord.MessageCreate{
				Content: fmt.Sprintf("Failed to play track: %s", err),
			})
			return err
		}
	}
	return nil
}

func (cmd *Commands) PlayPlaylist(data discord.SlashCommandInteractionData, event *handler.CommandEvent) error {

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

	var (
		next    OptBool
		loop    OptBool
		shuffle OptBool
	)

	n, ok := data.OptBool("next")
	next = optBoolValue(n, ok)

	l, ok := data.OptBool("loop")
	loop = optBoolValue(l, ok)

	s, ok := data.OptBool("shuffle")
	shuffle = optBoolValue(s, ok)

	return _Play(
		PlayOpts{
			Query:    data.Int("playlist"),
			Type:     PlaylistSearch,
			PlayNext: next,
			Loop:     loop,
			Shuffle:  shuffle,
		},
		event,
		cmd)
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

	var (
		next    OptBool
		loop    OptBool
		shuffle OptBool
	)

	n, ok := data.OptBool("next")
	next = optBoolValue(n, ok)

	l, ok := data.OptBool("loop")
	loop = optBoolValue(l, ok)

	s, ok := data.OptBool("shuffle")
	shuffle = optBoolValue(s, ok)

	return _Play(
		PlayOpts{
			Query:    data.String("query"),
			Type:     LavalinkSearch,
			PlayNext: next,
			Loop:     loop,
			Shuffle:  shuffle,
		},
		event,
		cmd)
}
