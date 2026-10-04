CREATE TABLE IF NOT EXISTS pending_path_update (
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    current_local_path TEXT NOT NULL,
    PRIMARY KEY (entity_type, entity_id),
    CHECK (entity_type IN ('asset', 'collection'))
);
