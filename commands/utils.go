package commands

import (
	"fmt"
	"time"

	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavasrc-plugin"
)

const DELETE_AFTER = 30

const (
	EMOJI_RESUME_PLAYER string = "<:mc_resume:1187705966263812218>"
	EMOJI_PAUSE_PLAYER  string = "<:mc_pause:1187705962358902806>"
	EMOJI_STOP_PLAYER   string = "<:mc_stop:1187705975638081557>"
	EMOJI_PLAY_PREVIOUS string = "<:mc_previous:1187705971070488627>"
	EMOJI_PLAY_NEXT     string = "<:mc_next:1187705968331591710>"
	EMOJI_RADIO_BUTTON  string = "<:mc_radio_button:1187818871072247858>"
	EMOJI_LOOP_OFF      string = "<:mc_loop_off:1189020553353371678>"
	EMOJI_LOOP_TRACK    string = "<:mc_loop_track:1189020551340114032>"
	EMOJI_LOOP_QUEUE    string = "<:mc_loop_queue:1189020548525735956>"
	EMOJI_SHUFFLE_OFF   string = "<:mc_shuffle_off:1189022239354531890>"
	EMOJI_SHUFFLE_ON    string = "<:mc_shuffle_on:1189022235621605498>"
)

func AutoRemove(e *handler.CommandEvent) {
	time.AfterFunc(DELETE_AFTER*time.Second, func() {
		e.DeleteInteractionResponse()
	})
}

func Trim(s string, length int) string {
	r := []rune(s)
	if len(r) > length {
		return string(r[:length-1]) + "…"
	}
	return s
}

func FormatTrack(track lavalink.Track) string {

	var lavasrcInfo lavasrc.TrackInfo
	_ = track.PluginInfo.Unmarshal(&lavasrcInfo)

	return ""
}

func FormatTime(d lavalink.Duration) string {

	if d.Hours() < 1 {
		return fmt.Sprintf("%02d:%02d", d.MinutesPart(), d.SecondsPart())
	} else if d.Days() < 1 {
		return fmt.Sprintf("%02d:%02d:%02d", d.HoursPart(), d.MinutesPart(), d.SecondsPart())
	} else {
		return fmt.Sprintf("%02d:%02d:%02d:%02d", d.Days(), d.HoursPart(), d.MinutesPart(), d.SecondsPart())
	}
}

func PlayerBar(player disgolink.Player) string {

	var (
		PlayPause string
		Playtime  string
		Bar       string
	)

	if player.Paused() {
		PlayPause = EMOJI_RESUME_PLAYER
	} else {
		PlayPause = EMOJI_PAUSE_PLAYER
	}

	if player.Track().Info.IsStream {
		Playtime = "LIVE"
		Bar = ProgressBar(0.99)
	} else {
		Playtime = fmt.Sprintf("`%s | %s`", FormatTime(player.Position()), FormatTime(player.Track().Info.Length))
		Bar = ProgressBar(float32(player.Position()) / float32(player.Track().Info.Length))
	}

	return fmt.Sprintf("%s %s `%s`", PlayPause, Bar, Playtime)

}

func ProgressBar(percent float32) string {

	bar := make([]rune, 12)

	for i := range bar {
		if i == int(percent*12) {
			bar[i] = '🔘'
		} else {
			bar[i] = '▬'
		}
	}
	return string(bar)
}
