-- +goose Up
BEGIN;

CREATE TABLE categories (
    id       SERIAL PRIMARY KEY,
    user_id  BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name     CITEXT      NOT NULL,
    emoji    TEXT,
    CONSTRAINT uq_user_cat UNIQUE (user_id, name)
);

CREATE TABLE transactions (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INT         REFERENCES categories(id),
    amount      NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    currency    CHAR(3)     NOT NULL,
    note        TEXT,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ
);

CREATE INDEX idx_tx_user_created ON transactions (user_id, created_at DESC);
CREATE INDEX idx_tx_note_trgm    ON transactions USING gin (note gin_trgm_ops);

COMMIT;

-- +goose Down
BEGIN;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS categories;
COMMIT;
