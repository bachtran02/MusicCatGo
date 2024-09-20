CREATE TABLE IF NOT EXISTS playlists
(
    id          VARCHAR      PRIMARY KEY,
    name        VARCHAR     NOT NULL,
    created_at  DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS tracks
(
    id          SERIAL    PRIMARY KEY,
    title       VARCHAR,
    author      VARCHAR,
    encoded     VARCHAR
);

CREATE TABLE IF NOT EXISTS users
(
    id          BIGINT,
    username    VARCHAR
);

CREATE TABLE IF NOT EXISTS playlist_tracks
(
    track_id        BIGINT
    playlist_id     INT NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
);