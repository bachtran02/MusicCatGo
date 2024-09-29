package musicbot

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5"
)

type Playlist struct {
	ID        int          `db:"id"`
	Name      string       `db:"name"`
	UserID    snowflake.ID `db:"user_id"`
	CreatedAt time.Time    `db:"created_at"`
}

type Track struct {
	ID      string `db:"id"`
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

func (d *DB) CreatePlaylist(ctx context.Context, userID snowflake.ID, username string, playlistName string) error {

	err := d.InsertUser(ctx, userID, username)
	if err != nil {
		return err
	}

	// TODO: handle playlists with same name
	err = d.InsertPlaylist(ctx, userID, playlistName)
	return err
}

func (d *DB) RemovePlaylist(ctx context.Context, userID snowflake.ID, name string) error {

	_, err := d.Pool.Exec(ctx, "DELETE FROM playlists WHERE user_id = $1 AND name = $2", userID, name)
	return err
}

func (d *DB) SearchPlaylist(ctx context.Context, userID snowflake.ID, query string) ([]Playlist, error) {
	var (
		playlists []Playlist
		rows      pgx.Rows
		dbquery   string
		err       error
	)

	if query == "" {
		dbquery = fmt.Sprintf("SELECT * FROM playlists WHERE user_id = %s", userID)
	} else {
		dbquery = fmt.Sprintf("SELECT * FROM playlists WHERE user_id = %s AND name ILIKE '%' || %s || '%'", userID, query)
	}

	rows, err = d.Pool.Query(ctx, dbquery)
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

func (d *DB) AddTrackToPlaylist(ctx context.Context, userID snowflake.ID, playlistName string, track lavalink.Track) error {

	// insert track
	err := d.InsertTrack(ctx, track)
	if err != nil {
		return err
	}

	// Fetch the playlist ID
	var playlistID int
	err = d.Pool.QueryRow(ctx, "SELECT id FROM playlists WHERE user_id = $1 AND name = $2", userID, playlistName).Scan(&playlistID)
	if err != nil {
		return err
	}

	// insert to playlist_tracks
	err = d.InsertPlaylistTrack(ctx, *track.Info.URI, playlistID)
	return err
}

func (d *DB) InsertPlaylist(ctx context.Context, userID snowflake.ID, playlistName string) error {
	_, err := d.Pool.Exec(ctx, "INSERT INTO playlists (user_id, name) VALUES ($1, $2)", userID, playlistName)
	return err
}

func (d *DB) InsertTrack(ctx context.Context, track lavalink.Track) error {
	_, err := d.Pool.Exec(ctx, `
		INSERT INTO tracks (id, title, author, encoded)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING`, track.Info.URI, track.Info.Title, track.Info.Author, track.Encoded)
	return err
}

func (d *DB) InsertUser(ctx context.Context, userID snowflake.ID, username string) error {
	_, err := d.Pool.Exec(ctx, `
		INSERT INTO users (id, username) VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE 
		SET username = EXCLUDED.username`, userID, username)
	return err
}

func (d *DB) InsertPlaylistTrack(ctx context.Context, trackID string, playlistID int) error {
	_, err := d.Pool.Exec(ctx,
		`INSERT INTO playlist_tracks (track_id, playlist_id) VALUES ($1, $2)
		ON CONFLICT (track_id, playlist_id) DO NOTHING`, trackID, playlistID)
	return err
}

func (d *DB) QueryPlaylistTracks(ctx context.Context, userID snowflake.ID, playlist_name string) ([]lavalink.Track, error) {
	var tracks []lavalink.Track

	rows, err := d.Pool.Query(ctx,
		`SELECT * FROM playlist_tracks 	
		INNER JOIN playlists ON playlists.id = playlist_tracks.playlist_id
		WHERE playlists.name = $1 AND playlists.user_id = $2`, playlist_name, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var track Track
		err := rows.Scan(&track.ID, &track.Title, &track.Author, &track.Encoded)
		if err != nil {
			slog.Error("failed to parse track from database", slog.Any("err", err))
			continue
		}

		// decode encoded lavalink track into lavalink.Track

		tracks = append(tracks, track)
	}
	return tracks, nil
}
