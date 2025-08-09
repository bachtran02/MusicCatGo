package musicbot

import (
	"testing"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupPlayerManager is a helper to create a PlayerManager and some default tracks for tests
func setupPlayerManager() (*PlayerManager, snowflake.ID, []lavalink.Track) {
	pm := NewPlayerManager()
	guildID := snowflake.ID(112233)
	tracks := []lavalink.Track{
		{Encoded: "track1"},
		{Encoded: "track2"},
		{Encoded: "track3"},
	}
	return pm, guildID, tracks
}

func TestPlayerManager_GetState(t *testing.T) {
	pm, guildID, tracks := setupPlayerManager()
	channelID := snowflake.ID(445566)

	// Test getting a non-existent state
	_, ok := pm.GetState(guildID)
	assert.False(t, ok, "Should not get a state for a guild that hasn't been added")

	// Add a state
	pm.Add(guildID, channelID, tracks...)

	// Test getting an existing state
	state, ok := pm.GetState(guildID)
	assert.True(t, ok, "Should get a state for a guild that has been added")
	require.NotNil(t, state)
	assert.Equal(t, channelID, state.channelID)
	assert.Equal(t, tracks, state.tracks)
}

func TestPlayerManager_Delete(t *testing.T) {
	pm, guildID, tracks := setupPlayerManager()
	pm.Add(guildID, snowflake.ID(445566), tracks...)

	// Ensure state exists before deleting
	_, ok := pm.GetState(guildID)
	require.True(t, ok, "State should exist before delete")

	// Delete the state
	pm.Delete(guildID)

	// Ensure state is gone
	_, ok = pm.GetState(guildID)
	assert.False(t, ok, "State should not exist after being deleted")
}

func TestPlayerManager_Add(t *testing.T) {
	pm, guildID, tracks := setupPlayerManager()
	channelID := snowflake.ID(445566)

	// Add tracks to a new guild
	pm.Add(guildID, channelID, tracks[0], tracks[1])
	state, ok := pm.GetState(guildID)
	require.True(t, ok)
	assert.Len(t, state.tracks, 2, "Should have 2 tracks in the queue")

	// Add more tracks to the same guild
	pm.Add(guildID, channelID, tracks[2])
	assert.Len(t, state.tracks, 3, "Should have 3 tracks in the queue after adding more")
	assert.Equal(t, tracks, state.tracks)
}

func TestPlayerManager_AddNext(t *testing.T) {
	pm, guildID, tracks := setupPlayerManager()
	channelID := snowflake.ID(445566)

	// AddNext on a new queue should behave like Add
	pm.AddNext(guildID, channelID, tracks[1])
	state, ok := pm.GetState(guildID)
	require.True(t, ok)
	assert.Equal(t, tracks[1].Encoded, state.tracks[0].Encoded)

	// Add a track to the front of the existing queue
	pm.AddNext(guildID, channelID, tracks[0])
	assert.Len(t, state.tracks, 2, "Queue should have 2 tracks")
	assert.Equal(t, tracks[0].Encoded, state.tracks[0].Encoded, "New track should be at the front")
	assert.Equal(t, tracks[1].Encoded, state.tracks[1].Encoded)
}

func TestPlayerManager_Next(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		pm, guildID, tracks := setupPlayerManager()
		pm.Add(guildID, 0, tracks...)

		// First call to Next
		track, ok := pm.Next(guildID)
		assert.True(t, ok)
		assert.Equal(t, tracks[0].Encoded, track.Encoded)
		state, _ := pm.GetState(guildID)
		assert.Equal(t, tracks[0].Encoded, state.current.Encoded)
		assert.Len(t, state.tracks, 2)
		assert.Len(t, state.prevtracks, 1)

		// Second call to Next
		track, ok = pm.Next(guildID)
		assert.True(t, ok)
		assert.Equal(t, tracks[1].Encoded, track.Encoded)
		assert.Len(t, state.tracks, 1)
		assert.Len(t, state.prevtracks, 2)

		// Play through the rest
		pm.Next(guildID)

		// Call on empty queue
		track, ok = pm.Next(guildID)
		assert.False(t, ok)
		assert.Empty(t, track.Encoded)
	})

	t.Run("LoopTrack", func(t *testing.T) {
		pm, guildID, tracks := setupPlayerManager()
		pm.Add(guildID, 0, tracks[0])
		state, _ := pm.GetState(guildID)
		state.SetLoop(LoopTrack)

		// First call to Next should set the current track
		track, ok := pm.Next(guildID)
		require.True(t, ok)
		require.Equal(t, tracks[0].Encoded, track.Encoded)

		// Subsequent calls should return the same track without changing the queue
		track, ok = pm.Next(guildID)
		assert.True(t, ok)
		assert.Equal(t, tracks[0].Encoded, track.Encoded)
		assert.Empty(t, state.tracks, "Queue should remain empty")
		assert.Equal(t, tracks[0].Encoded, state.current.Encoded)
	})

	t.Run("LoopQueue", func(t *testing.T) {
		pm, guildID, tracks := setupPlayerManager()
		pm.Add(guildID, 0, tracks...)
		state, _ := pm.GetState(guildID)
		state.SetLoop(LoopQueue)

		// Play first track
		firstTrack, ok := pm.Next(guildID)
		require.True(t, ok)
		require.Equal(t, tracks[0].Encoded, firstTrack.Encoded)
		assert.Len(t, state.tracks, 3, "Queue should still have 3 tracks")

		// Play second track
		pm.Next(guildID)
		// Play third track
		pm.Next(guildID)

		assert.Len(t, state.tracks, 3, "Queue should still have 3 tracks")
		assert.Equal(t, firstTrack.Encoded, state.tracks[0].Encoded, "The first track played should be back at the start of the queue")
	})

	t.Run("Shuffle", func(t *testing.T) {
		pm, guildID, tracks := setupPlayerManager()
		pm.Add(guildID, 0, tracks...)
		state, _ := pm.GetState(guildID)
		state.SetShuffle(ShuffleOn)

		// Play first track (randomly)
		track, ok := pm.Next(guildID)
		require.True(t, ok)
		assert.Len(t, state.tracks, 2, "Queue should have one less track")
		// The specific track is random, but it should be one of the initial tracks
		assert.Contains(t, tracks, track)

		// The played track should now be the current track
		assert.Equal(t, track.Encoded, state.current.Encoded)
	})
}

func TestPlayerManager_Previous(t *testing.T) {
	pm, guildID, tracks := setupPlayerManager()
	pm.Add(guildID, 0, tracks...)

	// Can't go previous if nothing has been played
	_, ok := pm.Previous(guildID)
	assert.False(t, ok, "Should not be able to go to previous track if none played")

	// Play two tracks
	pm.Next(guildID) // current: track1, prev: [track1]
	pm.Next(guildID) // current: track2, prev: [track1, track2]

	// Go back to the first track
	prevTrack, ok := pm.Previous(guildID)
	assert.True(t, ok)
	assert.Equal(t, tracks[0].Encoded, prevTrack.Encoded)

	state, _ := pm.GetState(guildID)
	assert.Equal(t, tracks[0].Encoded, state.current.Encoded, "Current track should be updated to previous")
	assert.Len(t, state.prevtracks, 1, "Previous tracks list should be smaller")
	assert.Equal(t, tracks[1].Encoded, state.tracks[0].Encoded, "The track we moved from should be at the front of the queue")

	// Can't go back further than the history allows
	_, ok = pm.Previous(guildID)
	assert.False(t, ok, "Should not be able to go back further than history")
}

func TestPlayerManager_RemoveTrack(t *testing.T) {
	pm, guildID, tracks := setupPlayerManager()
	pm.Add(guildID, 0, tracks...)

	// Remove the middle track
	removedTrack, ok := pm.RemoveTrack(guildID, 1)
	assert.True(t, ok)
	assert.Equal(t, tracks[1].Encoded, removedTrack.Encoded)

	state, _ := pm.GetState(guildID)
	assert.Len(t, state.tracks, 2, "Queue should have 2 tracks left")
	assert.Equal(t, tracks[0].Encoded, state.tracks[0].Encoded)
	assert.Equal(t, tracks[2].Encoded, state.tracks[1].Encoded)

	// Test removing from a non-existent guild
	_, ok = pm.RemoveTrack(snowflake.ID(999), 0)
	assert.False(t, ok)
}

func TestPlayerManager_Queue(t *testing.T) {
	pm, guildID, tracks := setupPlayerManager()

	// Test empty queue
	queue, ok := pm.Queue(guildID)
	assert.False(t, ok)
	assert.Empty(t, queue)

	// Test populated queue
	pm.Add(guildID, 0, tracks...)
	queue, ok = pm.Queue(guildID)
	assert.True(t, ok)
	assert.Equal(t, tracks, queue)
}
