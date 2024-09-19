CREATE TABLE IF NOT EXISTS playlists
(
    id      SERIAL PRIMARY KEY,
    name    VARCHAR   NOT NULL,

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
    playlist_id     SERIAL  NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
);