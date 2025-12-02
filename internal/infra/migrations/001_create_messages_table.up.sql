CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL CHECK (phone_number ~ '^[0-9+]+$'),
    content VARCHAR(160) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'sent', 'failed')),
    external_id VARCHAR(100),
    sent_time TIMESTAMP,
    attempt_count INT NOT NULL DEFAULT 0
);

-- Index for fast fetching pending messages
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_sent_time ON messages(sent_time);
