-- +migrate Up
-- +migrate StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS orders_db (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    label varchar NOT NULL,
    created_at varchar NOT NULL
);
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DROP TABLE IF EXISTS orders_db;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +migrate StatementEnd