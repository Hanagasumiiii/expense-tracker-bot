-- +goose Up
CREATE TABLE transactions (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT REFERENCES users(id) ON DELETE CASCADE,
    category_id INT    REFERENCES categories(id),
    amount      NUMERIC(12,2) CHECK (amount > 0) NOT NULL,
    currency    CHAR(3) NOT NULL,
    note        TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ
);

CREATE INDEX idx_tx_user_created ON transactions (user_id, created_at DESC);
CREATE INDEX idx_tx_note_trgm ON transactions USING GIN (note gin_trgm_ops);

-- +goose Down
DROP TABLE transactions;