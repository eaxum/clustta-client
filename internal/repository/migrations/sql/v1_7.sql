CREATE TABLE IF NOT EXISTS integration_project (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    integration_id TEXT NOT NULL,
    external_project_id TEXT NOT NULL,
    external_project_name TEXT DEFAULT '' NOT NULL,
    api_url TEXT DEFAULT '' NOT NULL,
    sync_options TEXT DEFAULT '{}' NOT NULL,
    linked_by_user_id TEXT DEFAULT '' NOT NULL,
    linked_at TEXT DEFAULT '' NOT NULL,
    enabled INTEGER DEFAULT 1 NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL
);

CREATE TABLE IF NOT EXISTS integration_collection_mapping (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    integration_id TEXT NOT NULL,
    external_id TEXT NOT NULL,
    external_type TEXT DEFAULT '' NOT NULL,
    external_name TEXT DEFAULT '' NOT NULL,
    external_parent_id TEXT DEFAULT '' NOT NULL,
    external_path TEXT DEFAULT '' NOT NULL,
    external_metadata TEXT DEFAULT '{}' NOT NULL,
    collection_id TEXT DEFAULT '' NOT NULL,
    synced_at TEXT DEFAULT '' NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    UNIQUE(integration_id, external_id),
    FOREIGN KEY (collection_id) REFERENCES entity(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS integration_asset_mapping (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    integration_id TEXT NOT NULL,
    external_id TEXT NOT NULL,
    external_name TEXT DEFAULT '' NOT NULL,
    external_parent_id TEXT DEFAULT '' NOT NULL,
    external_type TEXT DEFAULT '' NOT NULL,
    external_status TEXT DEFAULT '' NOT NULL,
    external_assignees TEXT DEFAULT '[]' NOT NULL,
    external_metadata TEXT DEFAULT '{}' NOT NULL,
    asset_id TEXT DEFAULT '' NOT NULL,
    last_pushed_checkpoint_id TEXT DEFAULT '' NOT NULL,
    synced_at TEXT DEFAULT '' NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    UNIQUE(integration_id, external_id),
    FOREIGN KEY (asset_id) REFERENCES task(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_integration_collection_mapping_collection
ON integration_collection_mapping(collection_id);

CREATE INDEX IF NOT EXISTS idx_integration_collection_mapping_external
ON integration_collection_mapping(integration_id, external_id);

CREATE INDEX IF NOT EXISTS idx_integration_asset_mapping_asset
ON integration_asset_mapping(asset_id);

CREATE INDEX IF NOT EXISTS idx_integration_asset_mapping_external
ON integration_asset_mapping(integration_id, external_id);

CREATE TRIGGER IF NOT EXISTS integration_project_update AFTER UPDATE ON integration_project
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE integration_project SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS integration_collection_mapping_update
AFTER UPDATE ON integration_collection_mapping
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE integration_collection_mapping SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS integration_asset_mapping_update
AFTER UPDATE ON integration_asset_mapping
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE integration_asset_mapping SET synced = 0 WHERE id = NEW.id;
END;
