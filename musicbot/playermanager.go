package musicbot

import (
	"sync"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"golang.org/x/exp/rand"
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
			loop:    LoopNone,
			shuffle: false,
		}
		q.states[guildID] = state
	}
	state.channelID = channelID
	state.tracks = append(state.tracks, tracks...)
}

func (q *PlayerManager) AddNext(guildID snowflake.ID, channelID snowflake.ID, tracks ...lavalink.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()

	state, ok := q.states[guildID]
	if !ok {
		q.Add(guildID, channelID, tracks...)
		return
	}

	state.channelID = channelID
	state.tracks = append(tracks, state.tracks...)
}

func (q *PlayerManager) Next(guildID snowflake.ID) (lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	player, ok := q.states[guildID]
	if !ok || (len(player.tracks) == 0 && player.loop == LoopNone) {
		player.current = lavalink.Track{}
		return lavalink.Track{}, false
	}

	track := player.current
	if player.loop != LoopTrack && len(player.tracks) > 0 {
		if player.shuffle == ShuffleOn {
			i := rand.Intn(len(player.tracks))
			track = player.tracks[i]
			player.tracks = append(player.tracks[:i], player.tracks[i+1:]...)

		} else {
			track = player.tracks[0]
			player.tracks = player.tracks[1:]
		}
		if player.loop == LoopQueue {
			player.tracks = append(player.tracks, player.current)
		}
		player.current = track
		player.prevtracks = append(player.prevtracks, track)
	}
	return track, true
}

func (q *PlayerManager) Previous(guildID snowflake.ID) (lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	player, ok := q.states[guildID]
	if !ok || len(player.prevtracks) < 2 {
		return lavalink.Track{}, false
	}

	track := player.prevtracks[len(player.prevtracks)-2]
	player.prevtracks = player.prevtracks[:len(player.prevtracks)-1]
	player.tracks = append([]lavalink.Track{player.current}, player.tracks...)
	player.current = track

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
