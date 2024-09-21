CREATE TABLE IF NOT EXISTS users
(
    id          BIGINT      PRIMARY KEY,
    username    VARCHAR
);

CREATE TABLE IF NOT EXISTS playlists
(
    id          SERIAL      PRIMARY KEY,
    name        VARCHAR     NOT NULL,
    user_id     BIGINT      NOT NULL,
    created_at  TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id)   REFERENCES users(id),
    UNIQUE (name, user_id)
);

CREATE TABLE IF NOT EXISTS tracks
(
    id          SERIAL     PRIMARY KEY,
    title       VARCHAR,
    author      VARCHAR,
    encoded     VARCHAR
);

CREATE TABLE IF NOT EXISTS playlist_tracks
(
    track_id        INT       NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    playlist_id     INT       NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    PRIMARY KEY (track_id, playlist_id)
);