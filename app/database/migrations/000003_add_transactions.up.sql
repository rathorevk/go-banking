CREATE TABLE transactions (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id BIGINT NOT NULL REFERENCES accounts(id),
    amount DECIMAL(10, 2) NOT NULL,
    source VARCHAR(20) NOT NULL CHECK (source IN ('game', 'server', 'payment')),
    type VARCHAR(20) NOT NULL CHECK (type IN ('win', 'lose')),
    inserted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index to quickly find transactions by account ID
CREATE INDEX idx_transactions_account_id ON transactions(account_id);
