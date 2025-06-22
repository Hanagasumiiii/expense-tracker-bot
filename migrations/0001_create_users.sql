-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE users (
    id           BIGSERIAL PRIMARY KEY,
    tg_id        BIGINT    UNIQUE NOT NULL,
    first_name   CITEXT,
    currency_def CHAR(3)   DEFAULT 'EUR'
);

-- +goose Down
DROP TABLE users;
DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS citext;