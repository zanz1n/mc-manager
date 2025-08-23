-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE instances (
    id bigint NOT NULL,
    user_id bigint,
    node_id bigint,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULt now(),
    last_launched timestamptz,
    name varchar(128) NOT NULL,
    description text NOT NULL,
    version varchar(16) NOT NULL,
    version_distro integer NOT NULL,
    maintenance boolean NOT NULL DEFAULT false,
    config jsonb NOT NULL,
    limits jsonb NOT NULL,

    PRIMARY KEY (id),

    CONSTRAINT instances_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON UPDATE CASCADE
    ON DELETE SET NULL,

    CONSTRAINT instances_node_id_fkey
    FOREIGN KEY (node_id) REFERENCES nodes(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT
);

CREATE INDEX instances_user_id_idx ON instances(user_id);
CREATE INDEX instances_node_id_idx ON instances(node_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE IF EXISTS instances;

-- +goose StatementEnd
