package musicbot

import (
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type Playlist struct {
	ID     int          `db:"id"`
	Name   string       `db:"name"`
	UserID snowflake.ID `db:"user_id"`
}

type PlaylistTrack struct {
	TrackID    int `db:track_id`
	PlaylistID int `db:"playlist_id"`
}

type Track struct {
	ID int `db:"id"`

	Position int            `db:"position"`
	Track    lavalink.Track `db:"track"`
}

type User struct {
	ID snowflake.ID
}

func (d *DB) InitDb() {

}

func (d *DB) CreatePlaylist(userID snowflake.ID, name string) (Playlist, error) {
	var playlist Playlist
	err := d.dbx.Get(&playlist, "INSERT INTO playlists (name, user_id) VALUES ($1, $2) RETURNING *", name, userID)
	return playlist, err

}
