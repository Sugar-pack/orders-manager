-- +migrate Up
-- +migrate StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS orders (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    label varchar NOT NULL,
    created_at timestamp NOT NULL
);
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DROP TABLE IF EXISTS orders;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +migrate StatementEnd