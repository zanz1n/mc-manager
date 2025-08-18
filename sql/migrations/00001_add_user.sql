-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE users (
    id bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULt now(),
    username varchar(32) NOT NULL,
    first_name varchar(32) NOT NULL,
    last_name varchar(32) NOT NULL,
    minecraft_user varchar(20) NOT NULL,
    email varchar(64) NOT NULL,
    admin boolean NOT NULL,
    two_fa boolean NOT NULL,
    password bytea NOT NULL,

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
