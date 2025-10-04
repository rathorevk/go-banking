CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    balance DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'EUR',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    inserted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index to quickly find accounts by user ID
CREATE INDEX idx_accounts_user_id ON accounts(user_id);

-- This ensures a user can have only one account per currency
CREATE UNIQUE INDEX idx_unique_account_per_currency ON accounts(user_id, currency);

-- Insert predefined accounts with initial 0 balances as required by test task
INSERT INTO accounts (user_id, balance) VALUES
(1, 0.00),
(2, 0.00),
(3, 0.00);

-- Reset the sequence to avoid conflicts with future inserts
SELECT setval(pg_get_serial_sequence('accounts', 'id'), (SELECT MAX(id) FROM accounts));

