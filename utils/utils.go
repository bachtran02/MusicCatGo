package utils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavasrc-plugin"
)

const deleteAfter = 30

const (
	RESUME_PLAYER_EMOJI_ID   int = 1187705966263812218
	PAUSE_PLAYER_EMOJI_ID    int = 1187705962358902806
	STOP_PLAYER_EMOJI_ID     int = 1187705975638081557
	PLAYER_PREVIOUS_EMOJI_ID int = 1187705971070488627
	PLAYER_NEXT_EMOJI_ID     int = 1187705968331591710

	LOOP_OFF_EMOJI_ID     int = 1189020553353371678
	LOOP_TRACK_EMOJI_ID   int = 1189020551340114032
	LOOP_QUEUE_EMOJI_ID   int = 1189020548525735956
	SHUFFLE_OFF_EMOJI_ID  int = 1189022239354531890
	SHUFFLE_ON_EMOJI_ID   int = 1189022235621605498
	RADIO_BUTTON_EMOJI_ID int = 1187818871072247858
)

var (
	resumePlayerEmoji string = "<:mc_resume:" + strconv.Itoa(RESUME_PLAYER_EMOJI_ID) + ">"
	pausePlayerEmoji  string = "<:mc_pause:" + strconv.Itoa(PAUSE_PLAYER_EMOJI_ID) + ">"
	// stopPlayerEmoji   string = "<:mc_stop:" + strconv.Itoa(stopPlayerEmojiID) + ">"
	// playPreviousEmoji string = "<:mc_previous:" + strconv.Itoa(playPreviousEmojiID) + ">"
	// playNextEmoji     string = "<:mc_next:" + strconv.Itoa(playNextEmojiID) + ">"
	// radioButtonEmoji  string = "<:mc_radio_button:" + strconv.Itoa(emojiRadioButtonID) + ">"
	// loopOffEmoji      string = "<:mc_loop_off:" + strconv.Itoa(emojiLoopOffID) + ">"
	// loopTrackEmoji    string = "<:mc_loop_track:" + strconv.Itoa(emojiLoopTrackID) + ">"
	// loopQueueEmoji    string = "<:mc_loop_queue:" + strconv.Itoa(emojiLoopQueueID) + ">"
	// shuffleOffEmoji   string = "<:mc_shuffle_off:" + strconv.Itoa(emojiShuffleOffID) + ">"
	// shuffleOnEmoji    string = "<:mc_shuffle_on:" + strconv.Itoa(emojiShuffleOnID) + ">"
)

func AutoRemove(e *handler.CommandEvent) {
	time.AfterFunc(deleteAfter*time.Second, func() {
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
		PlayPause = resumePlayerEmoji
	} else {
		PlayPause = pausePlayerEmoji
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
