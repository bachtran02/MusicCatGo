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

func (q *PlayerManager) getOrCreateState(guildID snowflake.ID) *PlayerState {
	state, ok := q.states[guildID]
	if !ok {
		state = &PlayerState{
			loop:    LoopNone,
			shuffle: false,
		}
		q.states[guildID] = state
	}
	return state
}

func (q *PlayerManager) Add(guildID snowflake.ID, channelID snowflake.ID, tracks ...lavalink.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()

	state := q.getOrCreateState(guildID)
	state.channelID = channelID
	state.tracks = append(state.tracks, tracks...)
}

func (q *PlayerManager) AddNext(guildID snowflake.ID, channelID snowflake.ID, tracks ...lavalink.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()

	state := q.getOrCreateState(guildID)
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

	var (
		current_track = player.current // current track
		next_track    lavalink.Track   // next track
	)

	if player.loop == LoopTrack && player.current.Encoded != "" {
		/* repeating current track */
		next_track = player.current
	} else if len(player.tracks) > 0 {
		/* selecting next track */
		if player.shuffle == ShuffleOn {
			/* select a random track from the queue */
			i := rand.Intn(len(player.tracks))
			next_track = player.tracks[i]
			player.tracks = append(player.tracks[:i], player.tracks[i+1:]...)
		} else {
			/* select firt track from the queue */
			next_track = player.tracks[0]
			player.tracks = player.tracks[1:]
		}
		if player.loop == LoopQueue && next_track.Encoded != "" {
			/* if in queue loop -> add next track back to queue */
			player.tracks = append(player.tracks, next_track)
		}
		player.current = next_track                                  /* replace current with next track */
		player.prevtracks = append(player.prevtracks, current_track) /* add old track to previous tracks */
	}
	return next_track, true
}

func (q *PlayerManager) Previous(guildID snowflake.ID) (lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	player, ok := q.states[guildID]
	if !ok || len(player.prevtracks) < 2 {
		return lavalink.Track{}, false
	}

	prev_track := player.prevtracks[len(player.prevtracks)-1]
	player.prevtracks = player.prevtracks[:len(player.prevtracks)-1]
	player.tracks = append([]lavalink.Track{player.current}, player.tracks...)
	player.current = prev_track

	return prev_track, true
}

func (q *PlayerManager) RemoveTrack(guildID snowflake.ID, trackIndex int) (lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	state, ok := q.states[guildID]
	if !ok {
		return lavalink.Track{}, false
	}

	removedTrack := state.tracks[trackIndex]
	state.tracks = append(state.tracks[:trackIndex], state.tracks[trackIndex+1:]...)
	return removedTrack, true
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

func (q *PlayerManager) IsPlaying(guildID snowflake.ID) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	state, ok := q.states[guildID]
	if !ok {
		return false
	}
	return state.current.Encoded != ""
}
