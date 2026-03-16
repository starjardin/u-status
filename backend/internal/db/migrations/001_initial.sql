-- +goose Up
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    plan TEXT NOT NULL DEFAULT 'free',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE monitors (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    interval_seconds INTEGER NOT NULL DEFAULT 60,
    status TEXT NOT NULL DEFAULT 'pending',
    alert_email INTEGER NOT NULL DEFAULT 1,
    is_public INTEGER NOT NULL DEFAULT 0,
    consecutive_failures INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_monitors_user_id ON monitors(user_id);

CREATE TABLE checks (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    status_code INTEGER,
    response_time_ms INTEGER,
    is_up INTEGER NOT NULL,
    error TEXT,
    checked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_checks_monitor_id ON checks(monitor_id);
CREATE INDEX idx_checks_checked_at ON checks(checked_at);

CREATE TABLE incidents (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    started_at DATETIME NOT NULL,
    resolved_at DATETIME,
    error TEXT
);

CREATE INDEX idx_incidents_monitor_id ON incidents(monitor_id);

-- +goose Down
DROP TABLE IF EXISTS incidents;
DROP TABLE IF EXISTS checks;
DROP TABLE IF EXISTS monitors;
DROP TABLE IF EXISTS users;
