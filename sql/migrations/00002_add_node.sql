-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE nodes (
    id bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULt now(),
    name varchar(128) NOT NULL,
    description text NOT NULL,
    maintenance boolean NOT NULL DEFAULT false,
    token varchar(128) NOT NULL,
    endpoint varchar(128) NOT NULL,
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
