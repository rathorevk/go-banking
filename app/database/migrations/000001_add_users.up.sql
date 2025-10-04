CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    inserted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert predefined users as required by test task
INSERT INTO users (id, username, full_name, email) VALUES
(1, 'user1', 'Test User 1', 'user1@example.com'),
(2, 'user2', 'Test User 2', 'user2@example.com'),
(3, 'user3', 'Test User 3', 'user3@example.com');

-- Reset the sequence to avoid conflicts with future inserts
SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT MAX(id) FROM users));