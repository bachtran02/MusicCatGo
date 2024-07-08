package musicbot

import (
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type RepeatMode string

const (
	RepeatModeNone  RepeatMode = "none"
	RepeatModeTrack RepeatMode = "track"
	RepeatModeQueue RepeatMode = "queue"
)

type ShuffleMode bool

const (
	ShuffleOn  ShuffleMode = true
	ShuffleOff ShuffleMode = false
)

type PlayerState struct {
	tracks []lavalink.Track
	// prevtracks []lavalink.Track
	paused    bool
	repeat    RepeatMode
	shuffle   ShuffleMode
	channelID snowflake.ID
	messageID snowflake.ID
}

func (s *PlayerState) SetPause(paused bool) {
	s.paused = paused
}

func (s *PlayerState) Paused() bool {
	return s.paused
}

func (s *PlayerState) SetRepeat(repeat RepeatMode) {
	s.repeat = repeat
}

func (s *PlayerState) Repeat() RepeatMode {
	return s.repeat
}

func (s *PlayerState) SetShuffe(shuffle ShuffleMode) {
	s.shuffle = shuffle
}

func (s *PlayerState) Shuffle() ShuffleMode {
	return s.shuffle
}

func (s *PlayerState) SetMessageID(messageID snowflake.ID) {
	s.messageID = messageID
}

func (s *PlayerState) MessageID() snowflake.ID {
	return s.messageID
}

func (s *PlayerState) ChannelID() snowflake.ID {
	return s.channelID
}
