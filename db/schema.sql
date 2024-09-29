-- psql -h localhost -p 5432 -U musiccatgo -d playlist_db

CREATE TABLE IF NOT EXISTS users
(
    id          BIGINT              PRIMARY KEY,
    username    VARCHAR(255) 
);

CREATE TABLE IF NOT EXISTS playlists
(
    id          SERIAL              PRIMARY KEY,
    name        VARCHAR(255)        NOT NULL,
    user_id     BIGINT              NOT NULL,
    created_at  TIMESTAMP           DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id)           REFERENCES users(id),
    UNIQUE (name, user_id)
);

CREATE TABLE IF NOT EXISTS tracks
(
    id          VARCHAR(255)     PRIMARY KEY,
    title       VARCHAR(255),
    author      VARCHAR(255),
    encoded     TEXT
);

CREATE TABLE IF NOT EXISTS playlist_tracks
(
    track_id        VARCHAR(255)        NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    playlist_id     INT                 NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    PRIMARY KEY (track_id, playlist_id)
);