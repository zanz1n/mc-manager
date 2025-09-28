-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE users (
    id integer,
    created_at timestamp NOT NULL DEFAULT (now()),
    updated_at timestamp NOT NULL DEFAULT (now()),
    username text NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    minecraft_user text NOT NULL,
    email text NOT NULL,
    admin boolean NOT NULL,
    two_fa boolean NOT NULL,
    password blob NOT NULL,

    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX users_username_idx ON users(username);
CREATE UNIQUE INDEX users_email_idx ON users(email);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE IF EXISTS users;

-- +goose StatementEnd
