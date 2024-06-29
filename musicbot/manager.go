package musicbot

import (
	"sync"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		Players: map[snowflake.ID]*PlayerState{},
	}
}

type PlayerManager struct {
	Players map[snowflake.ID]*PlayerState
	mu      sync.Mutex
}

func (q *PlayerManager) PlayerState(guildID snowflake.ID) (*PlayerState, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	player, ok := q.Players[guildID]
	if !ok {
		return nil, false
	}
	return player, true
}

func (q *PlayerManager) Delete(guildID snowflake.ID) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.Players, guildID)
}

func (q *PlayerManager) Add(guildID snowflake.ID, channelID snowflake.ID, tracks ...lavalink.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()

	player, ok := q.Players[guildID]
	if !ok {
		// initiate new player with configs here
		player = &PlayerState{
			channelID: channelID,
		}
		q.Players[guildID] = player
	}
	player.tracks = append(player.tracks, tracks...)
}

func (q *PlayerManager) Next(guildID snowflake.ID) (lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	player, ok := q.Players[guildID]
	if !ok || len(player.tracks) == 0 {
		return lavalink.Track{}, false
	}

	track := player.tracks[0]
	if player.repeat != RepeatModeTrack {
		if player.repeat == RepeatModeQueue {
			player.tracks = append(player.tracks, track)
		}
		player.tracks = player.tracks[1:]
	}
	return track, true
}

func (q *PlayerManager) Queue(guildID snowflake.ID) ([]lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	player, ok := q.Players[guildID]
	if !ok || len(player.tracks) == 0 {
		return []lavalink.Track{}, false
	}
	return player.tracks, true
}
