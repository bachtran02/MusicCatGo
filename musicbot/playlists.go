package musicbot

import (
	"context"
	"log/slog"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

type Playlist struct {
	ID        int          `db:"id"`
	Name      string       `db:"name"`
	UserID    snowflake.ID `db:"user_id"`
	CreatedAt time.Time    `db:"created_at"`
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

func (d *DB) CreatePlaylist(ctx context.Context, userID snowflake.ID, username string, name string) error {

	var exists bool
	err := d.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		_, err = d.Pool.Exec(ctx, "INSERT INTO users (id, username) VALUES ($1, $2)", userID, username)
		if err != nil {
			return err
		}
	}

	_, err = d.Pool.Exec(ctx, "INSERT INTO playlists (user_id, name) VALUES ($1, $2)", userID, name)
	return err
}

func (d *DB) RemovePlaylist(ctx context.Context, userID snowflake.ID, name string) error {

	_, err := d.Pool.Exec(ctx, "DELETE FROM playlists WHERE user_id = $1 AND name = $2", userID, name)
	return err
}

func (d *DB) SearchPlaylist(ctx context.Context, userID snowflake.ID, query string) ([]Playlist, error) {
	var playlists []Playlist

	rows, err := d.Pool.Query(ctx, "SELECT * FROM playlists WHERE user_id = $1 AND name ILIKE '%' || $2 || '%'", userID, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var playlist Playlist
		err := rows.Scan(&playlist.ID, &playlist.Name, &playlist.UserID, &playlist.CreatedAt)
		if err != nil {
			slog.Error("failed to parse playlist from database", slog.Any("err", err))
			continue
		}
		playlists = append(playlists, playlist)
	}
	return playlists, nil
}
