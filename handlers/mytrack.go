package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type MyTrackHandler struct {
	ChannelID snowflake.ID
	GuildID   snowflake.ID
	track     lavalink.Track
	isPlaying bool
	mutex     sync.Mutex
}

func (h *MyTrackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	h.mutex.Lock()
	trackData := map[string]interface{}{
		"title":       h.track.Info.Title,
		"artist":      h.track.Info.Author,
		"url":         h.track.Info.URI,
		"artwork_url": h.track.Info.ArtworkURL,
		"is_playing":  h.isPlaying,
	}
	h.mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(trackData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *MyTrackHandler) OnTrackStart(p disgolink.Player, event lavalink.TrackStartEvent) {

	if p.GuildID() == h.GuildID && *p.ChannelID() == h.ChannelID {
		h.mutex.Lock()
		h.track = event.Track
		h.isPlaying = true
		h.mutex.Unlock()
	}
}

func (h *MyTrackHandler) OnTrackEnd(p disgolink.Player, event lavalink.TrackEndEvent) {

	h.mutex.Lock()
	if p.GuildID() == h.GuildID && h.isPlaying {
		h.track = lavalink.Track{}
		h.isPlaying = false
	}
	h.mutex.Unlock()
}
