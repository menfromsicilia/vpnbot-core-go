-- Enable WAL mode for better concurrency
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 10000;
PRAGMA temp_store = MEMORY;

-- Servers table
CREATE TABLE IF NOT EXISTS servers (
    country_code TEXT NOT NULL,
    city_name TEXT NOT NULL,
    ext_name TEXT,
    endpoint TEXT PRIMARY KEY,
    inbound_type TEXT NOT NULL DEFAULT 'trojan',
    active BOOLEAN NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_servers_active ON servers(active);
CREATE INDEX IF NOT EXISTS idx_servers_inbound_type ON servers(inbound_type);

-- User nodes tracking table
CREATE TABLE IF NOT EXISTS user_nodes (
    user_id TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    inbound TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, endpoint, inbound),
    FOREIGN KEY (endpoint) REFERENCES servers(endpoint) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_nodes_user_id ON user_nodes(user_id);
CREATE INDEX IF NOT EXISTS idx_user_nodes_endpoint ON user_nodes(endpoint);

-- Pending deletions table (for tracking failed deletion attempts)
CREATE TABLE IF NOT EXISTS pending_deletions (
    user_id TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    inbound TEXT NOT NULL,
    attempts INT DEFAULT 1,
    last_attempt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    error_message TEXT,
    PRIMARY KEY (user_id, endpoint, inbound)
);

CREATE INDEX IF NOT EXISTS idx_pending_deletions_endpoint ON pending_deletions(endpoint);
CREATE INDEX IF NOT EXISTS idx_pending_deletions_created ON pending_deletions(created_at);

