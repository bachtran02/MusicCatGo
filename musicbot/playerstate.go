package musicbot

import (
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type LoopMode string

const (
	LoopNone  LoopMode = "none"
	LoopTrack LoopMode = "track"
	LoopQueue LoopMode = "queue"
)

type ShuffleMode bool

const (
	ShuffleOn  ShuffleMode = true
	ShuffleOff ShuffleMode = false
)

type PlayerState struct {
	current    lavalink.Track
	tracks     []lavalink.Track
	prevtracks []lavalink.Track
	paused     bool
	loop       LoopMode
	shuffle    ShuffleMode
	channelID  snowflake.ID
	messageID  snowflake.ID
}

func (s *PlayerState) SetPause(paused bool) {
	s.paused = paused
}

func (s *PlayerState) Paused() bool {
	return s.paused
}

func (s *PlayerState) SetLoop(loop LoopMode) {
	s.loop = loop
}

func (s *PlayerState) Loop() LoopMode {
	return s.loop
}

func (s *PlayerState) SetShuffle(shuffle ShuffleMode) {
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
