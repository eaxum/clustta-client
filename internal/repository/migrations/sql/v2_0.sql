CREATE TABLE IF NOT EXISTS project_storage (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    mode TEXT NOT NULL CHECK (mode IN ('compact', 'deflated', 'object_storage')),
    updated_at INTEGER NOT NULL
);

INSERT OR IGNORE INTO project_storage (id, mode, updated_at)
VALUES (1, 'compact', unixepoch());

CREATE TABLE IF NOT EXISTS project_storage_conversion (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    source_mode TEXT NOT NULL CHECK (source_mode IN ('compact', 'deflated')),
    target_mode TEXT NOT NULL CHECK (target_mode IN ('compact', 'deflated')),
    status TEXT NOT NULL CHECK (status IN ('running', 'failed', 'cleanup_failed', 'completed')),
    total_chunks INTEGER NOT NULL DEFAULT 0,
    processed_chunks INTEGER NOT NULL DEFAULT 0,
    required_bytes INTEGER NOT NULL DEFAULT 0,
    processed_bytes INTEGER NOT NULL DEFAULT 0,
    error TEXT NOT NULL DEFAULT '',
    started_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS chunk_ref (
    hash TEXT PRIMARY KEY NOT NULL,
    storage_key TEXT NOT NULL,
    size INTEGER NOT NULL,
    created_at INTEGER NOT NULL
);
