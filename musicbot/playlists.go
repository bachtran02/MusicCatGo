package musicbot

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
)

type Playlist struct {
	ID     string       `db:"id"`
	Name   string       `db:"name"`
	UserID snowflake.ID `db:"user_id"`
}

type Track struct {
	ID      int    `db:"id"`
	Title   string `db:"title"`
	Author  string `db:"author"`
	Encoded string `db:"encoded"`
}

type User struct {
	ID       snowflake.ID `db:"id"`
	Username string       `db:"username"`
}

type PlaylistTrack struct {
	TrackID    int `db:"track_id"`
	PlaylistID int `db:"playlist_id"`
}

func (d *DB) CreatePlaylist(ctx context.Context, userID snowflake.ID, name string) error {

	_, err := d.Pool.Exec(ctx, "INSERT INTO playlists (id, name, user_id) VALUES ($1, $2, $3)", name, userID)
	return err
}

func (d *DB) DeletePlaylist(ctx context.Context, userID snowflake.ID, name string) error {

	_, err := d.Pool.Exec(ctx, "DELETE FROM playlists WHERE ")

}
