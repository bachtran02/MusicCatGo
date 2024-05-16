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

type Player struct {
	tracks []lavalink.Track
	// prevtracks    []lavalink.Track
	repeat RepeatMode
	// shuffle       bool
	channelID snowflake.ID
	// textchannelID snowflake.ID
	// playerID      snowflake.ID
}
