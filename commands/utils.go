package commands

import (
	"time"

	"github.com/disgoorg/disgo/handler"
)

const DELETE_AFTER = 30

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
