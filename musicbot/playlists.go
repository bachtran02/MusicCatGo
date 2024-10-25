package musicbot

import (
	"context"
	"encoding/json"
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
	OwnerID   snowflake.ID `db:"owner_id"`
	CreatedAt time.Time    `db:"created_at"`
}

type PlaylistTrack struct {
	ID         string         `db:"id"`
	PlaylistID int            `db:"playlist_id"`
	Track      lavalink.Track `db:"track"`
	AddedAt    time.Time      `db:"added_at"`
}

func (d *DB) CreatePlaylist(ctx context.Context, userID snowflake.ID, username string, playlistName string) error {
	_, err := d.Pool.Exec(ctx, "INSERT INTO playlists (owner_id, name) VALUES ($1, $2)", userID, playlistName)

	// TODO: handle playlist with owner_id + name already existed
	return err
}

func (d *DB) RemovePlaylist(ctx context.Context, userID snowflake.ID, name string) error {
	_, err := d.Pool.Exec(ctx, "DELETE FROM playlists WHERE owner_id = $1 AND name = $2", userID, name)
	return err
}

func (d *DB) SearchPlaylist(ctx context.Context, userID snowflake.ID, query string, limit int) ([]Playlist, error) {
	var (
		playlists []Playlist
		rows      pgx.Rows
		err       error
	)

	if query == "" {
		rows, err = d.Pool.Query(ctx, "SELECT * FROM playlists WHERE owner_id = $1 LIMIT $2", userID, limit)
	} else {
		rows, err = d.Pool.Query(ctx, "SELECT * FROM playlists WHERE owner_id = $1 AND name ILIKE $2 || '%' LIMIT $3;", userID, query, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var playlist Playlist
		err := rows.Scan(&playlist.ID, &playlist.Name, &playlist.OwnerID, &playlist.CreatedAt)
		if err != nil {
			slog.Error("failed to parse playlist from database", slog.Any("err", err))
			continue
		}
		playlists = append(playlists, playlist)
	}
	return playlists, nil
}

func (d *DB) GetPlaylist(ctx context.Context, playlistID int) (Playlist, []PlaylistTrack, error) {
	var playlist Playlist

	row := d.Pool.QueryRow(ctx, "SELECT * FROM playlists WHERE id = $1", playlistID)

	if err := row.Scan(&playlist.ID, &playlist.Name, &playlist.OwnerID, &playlist.CreatedAt); err != nil {
		return Playlist{}, nil, err
	}

	var tracks []PlaylistTrack
	rows, err := d.Pool.Query(ctx, "SELECT * FROM playlist_tracks WHERE playlist_id = $1", playlistID)

	if err != nil {
		return playlist, nil, err
	}

	for rows.Next() {
		var (
			track    PlaylistTrack
			rawTrack json.RawMessage
			err      error
		)
		err = rows.Scan(&track.ID, &track.PlaylistID, &rawTrack, &track.AddedAt)
		if err != nil {
			slog.Error("failed to parse playlist track from database", slog.Any("err", err))
			continue
		}

		err = json.Unmarshal(rawTrack, &track.Track)
		if err != nil {
			slog.Error("failed to decode track object", slog.Any("err", err))
			continue
		}

		tracks = append(tracks, track)
	}
	return playlist, tracks, nil
}

func (d *DB) AddTracksToPlaylist(ctx context.Context, playlistID int, tracks []lavalink.Track) error {

	if len(tracks) == 0 {
		return nil
	}

	query := "INSERT INTO playlist_tracks (playlist_id, track) VALUES "
	values := []interface{}{}

	for i, track := range tracks {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($1, $%d)", i+2)
		values = append(values, track)
	}

	query += ";" // End the query

	values = append([]interface{}{playlistID}, values...)

	_, err := d.Pool.Exec(ctx, query, values...)
	return err
}
