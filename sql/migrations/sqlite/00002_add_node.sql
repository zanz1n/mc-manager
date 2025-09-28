-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE nodes (
    id integer NOT NULL,
    created_at timestamp NOT NULL DEFAULT (now()),
    updated_at timestamp NOT NULL DEFAULT (now()),
    name text NOT NULL,
    description text NOT NULL,
    maintenance boolean NOT NULL DEFAULT false,
    token text NOT NULL,
    endpoint text NOT NULL,
    endpoint_tls boolean NOT NULL,
    ftp_port integer NOT NULL,
    grpc_port integer NOT NULL,

    PRIMARY KEY (id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE IF EXISTS nodes;

-- +goose StatementEnd
