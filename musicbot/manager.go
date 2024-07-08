package musicbot

import (
	"sync"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		states: map[snowflake.ID]*PlayerState{},
	}
}

type PlayerManager struct {
	states map[snowflake.ID]*PlayerState
	mu     sync.Mutex
}

func (q *PlayerManager) GetState(guildID snowflake.ID) (*PlayerState, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	state, ok := q.states[guildID]
	if !ok {
		return nil, false
	}
	return state, true
}

func (q *PlayerManager) Delete(guildID snowflake.ID) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.states, guildID)
}

func (q *PlayerManager) Add(guildID snowflake.ID, channelID snowflake.ID, tracks ...lavalink.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()

	state, ok := q.states[guildID]
	if !ok {
		state = &PlayerState{
			repeat:  RepeatModeNone,
			shuffle: false,
		}
		q.states[guildID] = state
	}
	state.channelID = channelID
	state.tracks = append(state.tracks, tracks...)
}

func (q *PlayerManager) Next(guildID snowflake.ID) (lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	player, ok := q.states[guildID]
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

	player, ok := q.states[guildID]
	if !ok || len(player.tracks) == 0 {
		return []lavalink.Track{}, false
	}
	return player.tracks, true
}
