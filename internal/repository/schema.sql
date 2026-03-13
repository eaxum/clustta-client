CREATE TABLE IF NOT EXISTS config (
    name TEXT PRIMARY KEY NOT NULL COLLATE NOCASE,
    value CLOB,
    mtime INTEGER NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TABLE IF NOT EXISTS preview (
    hash TEXT PRIMARY KEY,
    preview BLOB,
    extension TEXT DEFAULT '' NOT NULL
);

CREATE TABLE IF NOT EXISTS template (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL COLLATE NOCASE,
    extension TEXT NOT NULL,
    xxhash_checksum TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    chunks TEXT NOT NULL,
    trashed BOOLEAN DEFAULT 0 NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
	UNIQUE (name, extension),
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS template_update AFTER UPDATE ON template
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE template SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS template_delete AFTER DELETE ON template
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'template', 0);
END;


CREATE TABLE IF NOT EXISTS workflow (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT UNIQUE NOT NULL COLLATE NOCASE,
    synced BOOLEAN DEFAULT 0 NOT NULL
);


CREATE TRIGGER IF NOT EXISTS workflow_update AFTER UPDATE ON workflow
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE workflow SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS workflow_delete AFTER DELETE ON workflow
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'workflow', 0);
END;

CREATE TABLE IF NOT EXISTS workflow_collection (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL COLLATE NOCASE,
    workflow_id TEXT NOT NULL,
    collection_type_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (workflow_id) REFERENCES workflow(id),
    FOREIGN KEY (collection_type_id) REFERENCES collection_type(id),
    UNIQUE (name, workflow_id),
    CHECK( typeof(workflow_id)='text' AND length(workflow_id)>=1),
    CHECK( typeof(collection_type_id)='text' AND length(collection_type_id)>=1),
    CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS workflow_collection_update AFTER UPDATE ON workflow_collection
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE workflow_collection SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS workflow_collection_delete AFTER DELETE ON workflow_collection
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'workflow_collection', 0);
END;

CREATE TABLE IF NOT EXISTS workflow_asset (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL COLLATE NOCASE,
    workflow_id TEXT NOT NULL,
    is_resource BOOLEAN DEFAULT 0 NOT NULL,
	is_link BOOLEAN DEFAULT 0 NOT NULL,
	pointer TEXT DEFAULT '' NOT NULL,
    template_id TEXT NOT NULL,
    asset_type_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (workflow_id) REFERENCES workflow(id),
    FOREIGN KEY (template_id) REFERENCES template(id),
    FOREIGN KEY (asset_type_id) REFERENCES asset_type(id),
    UNIQUE (name, workflow_id),
	CHECK( typeof(workflow_id)='text' AND length(workflow_id)>=1),
	CHECK( typeof(template_id)='text' AND length(template_id)>=1),
	CHECK( typeof(asset_type_id)='text' AND length(asset_type_id)>=1),
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS workflow_asset_update AFTER UPDATE ON workflow_asset
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE workflow_asset SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS workflow_asset_delete AFTER DELETE ON workflow_asset
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'workflow_asset', 0);
END;

CREATE TABLE IF NOT EXISTS workflow_link (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL COLLATE NOCASE,
    collection_type_id TEXT NOT NULL,
    workflow_id TEXT NOT NULL,
    linked_workflow_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (workflow_id) REFERENCES workflow(id),
    FOREIGN KEY (linked_workflow_id) REFERENCES workflow(id),
    FOREIGN KEY (collection_type_id) REFERENCES collection_type(id),
    UNIQUE (workflow_id, linked_workflow_id, name),
    CHECK( typeof(collection_type_id)='text' AND length(collection_type_id)>=1)
);


CREATE TRIGGER IF NOT EXISTS workflow_link_update AFTER UPDATE ON workflow_link
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE workflow_link SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS workflow_link_delete AFTER DELETE ON workflow_link
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'workflow_link', 0);
END;


CREATE TRIGGER IF NOT EXISTS prevent_circular_link
BEFORE INSERT ON workflow_link
FOR EACH ROW
BEGIN
    WITH RECURSIVE link_chain AS (
        -- Start with existing links
        SELECT workflow_id, linked_workflow_id
        FROM workflow_link
        
        UNION ALL
        
        -- Follow the chain
        SELECT cc.workflow_id, wc.linked_workflow_id
        FROM link_chain cc
        JOIN workflow_link wc ON cc.linked_workflow_id = wc.workflow_id
    )
    SELECT RAISE(ABORT, 'Circular link detected')
    WHERE EXISTS (
        SELECT 1 
        FROM link_chain
        WHERE workflow_id = NEW.linked_workflow_id 
        AND linked_workflow_id = NEW.workflow_id
    );
END;


CREATE TABLE IF NOT EXISTS tag (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT UNIQUE NOT NULL COLLATE NOCASE,
    synced BOOLEAN DEFAULT 0 NOT NULL,
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS tag_update AFTER UPDATE ON tag
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE tag SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS tag_delete AFTER DELETE ON tag
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'tag', 0);
END;

CREATE TABLE IF NOT EXISTS collection (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL COLLATE NOCASE,
    collection_path TEXT DEFAULT '' NOT NULL,
    description TEXT,
    collection_type_id TEXT NOT NULL,
    parent_id TEXT NOT NULL,
	trashed BOOLEAN DEFAULT 0 NOT NULL,
    preview_id TEXT DEFAULT '' NOT NULL,
	synced BOOLEAN DEFAULT 0 NOT NULL,
	is_library BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (collection_type_id) REFERENCES collection_type(id),
    FOREIGN KEY (parent_id) REFERENCES collection(id),
    FOREIGN KEY (preview_id) REFERENCES preview(hash),
    UNIQUE (name, parent_id),
	CHECK( typeof(collection_type_id)='text' AND length(collection_type_id)>=1),
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS collection_update AFTER UPDATE ON collection
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE collection SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS collection_delete AFTER DELETE ON collection
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'collection', 0);
END;

-- Trigger to maintain materialized path on INSERT
CREATE TRIGGER IF NOT EXISTS collection_path_insert 
AFTER INSERT ON collection
FOR EACH ROW
BEGIN
    UPDATE collection
    SET collection_path = 
        CASE
        WHEN NEW.parent_id = '' OR NEW.parent_id IS NULL THEN '/' || NEW.name || '/'
        ELSE (
            SELECT collection_path || NEW.name || '/' FROM collection WHERE id = NEW.parent_id
        )
        END
    WHERE id = NEW.id;
END;

-- Updated collection path trigger to handle orphaned collections
CREATE TRIGGER IF NOT EXISTS collection_path_update 
AFTER UPDATE OF name, parent_id ON collection
FOR EACH ROW
WHEN OLD.name != NEW.name OR OLD.parent_id != NEW.parent_id
BEGIN
    -- Recalculate this collection's path
  UPDATE collection
  SET collection_path =
    CASE
      WHEN NEW.parent_id IS NULL THEN '/' || NEW.name || '/'
      ELSE COALESCE(
        (SELECT collection_path || NEW.name || '/' FROM collection WHERE id = NEW.parent_id),
        '/' || NEW.name || '/'
      )
    END
  WHERE id = NEW.id;

  -- Recalculate all descendant paths
  UPDATE collection
  SET collection_path =
    (SELECT collection_path FROM collection WHERE id = NEW.id) || substr(collection_path, length(OLD.collection_path) + 1)
  WHERE collection_path LIKE OLD.collection_path || '%'
    AND id != NEW.id;
END;


CREATE TABLE IF NOT EXISTS collection_assignee (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    collection_id TEXT NOT NULL,
    assignee_id TEXT DEFAULT '' NOT NULL,
    assigner_id TEXT DEFAULT '' NOT NULL,
	synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (collection_id) REFERENCES collection(id),
    FOREIGN KEY (assignee_id) REFERENCES user(id),
    FOREIGN KEY (assigner_id) REFERENCES user(id),
    UNIQUE (collection_id, assignee_id),
	CHECK( typeof(collection_id)='text' AND length(collection_id)>=1)
);

CREATE TRIGGER IF NOT EXISTS collection_assignee_update AFTER UPDATE ON collection_assignee
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE collection_assignee SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS collection_assignee_delete AFTER DELETE ON collection_assignee
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'collection_assignee', 0);
END;

CREATE TABLE IF NOT EXISTS collection_type (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL UNIQUE COLLATE NOCASE,
    icon TEXT NOT NULL UNIQUE COLLATE NOCASE,
    synced BOOLEAN DEFAULT 0 NOT NULL,
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS collection_type_update AFTER UPDATE ON collection_type
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE collection_type SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS collection_type_delete AFTER DELETE ON collection_type
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'collection_type', 0);
END;

CREATE TABLE IF NOT EXISTS asset (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL COLLATE NOCASE,
    description TEXT DEFAULT '' NOT NULL,
    extension TEXT NOT NULL,
    is_resource BOOLEAN DEFAULT 0 NOT NULL,
	is_link BOOLEAN DEFAULT 0 NOT NULL,
	pointer TEXT DEFAULT '' NOT NULL,
    status_id TEXT NOT NULL,
    asset_type_id TEXT NOT NULL,
    collection_id TEXT DEFAULT '' NOT NULL,
	assignee_id TEXT DEFAULT '' NOT NULL,
	assigner_id TEXT DEFAULT '' NOT NULL,
    preview_id TEXT DEFAULT '' NOT NULL,
    trashed BOOLEAN DEFAULT 0 NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (preview_id) REFERENCES preview(hash),
    FOREIGN KEY (status_id) REFERENCES status(id),
    FOREIGN KEY (asset_type_id) REFERENCES asset_type(id),
    FOREIGN KEY (collection_id) REFERENCES collection(id),
	FOREIGN KEY (assignee_id) REFERENCES user(id),
	FOREIGN KEY (assigner_id) REFERENCES user(id),
    UNIQUE (name, collection_id, extension),
	CHECK( typeof(asset_type_id)='text' AND length(asset_type_id)>=1),
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS asset_update AFTER UPDATE ON asset
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE asset SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS asset_delete AFTER DELETE ON asset
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'asset', 0);
END;

CREATE TABLE IF NOT EXISTS asset_type (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL UNIQUE COLLATE NOCASE,
    icon TEXT NOT NULL UNIQUE COLLATE NOCASE,
    synced BOOLEAN DEFAULT 0 NOT NULL,
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS asset_type_update AFTER UPDATE ON asset_type
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE asset_type SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS asset_type_delete AFTER DELETE ON asset_type
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'asset_type', 0);
END;

CREATE TABLE IF NOT EXISTS dependency_type (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT UNIQUE NOT NULL COLLATE NOCASE,
    synced BOOLEAN DEFAULT 0 NOT NULL,
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS dependency_type_update AFTER UPDATE ON dependency_type
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE dependency_type SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS dependency_type_delete AFTER DELETE ON dependency_type
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'dependency_type', 0);
END;

CREATE TABLE IF NOT EXISTS collection_dependency (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    asset_id TEXT NOT NULL,
    dependency_id TEXT NOT NULL,
    dependency_type_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (asset_id) REFERENCES asset(id),
    FOREIGN KEY (dependency_id) REFERENCES collection(id),
    FOREIGN KEY (dependency_type_id) REFERENCES dependency_type(id),
    UNIQUE (asset_id, dependency_id)
);

CREATE TRIGGER IF NOT EXISTS collection_dependency_update AFTER UPDATE ON collection_dependency
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE collection_dependency SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS collection_dependency_delete AFTER DELETE ON collection_dependency
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'collection_dependency', 0);
END;

CREATE TABLE IF NOT EXISTS asset_dependency (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    asset_id TEXT NOT NULL,
    dependency_id TEXT NOT NULL,
    dependency_type_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (asset_id) REFERENCES asset(id),
    FOREIGN KEY (dependency_id) REFERENCES asset(id),
    FOREIGN KEY (dependency_type_id) REFERENCES dependency_type(id),
    UNIQUE (asset_id, dependency_id)
);

CREATE TRIGGER IF NOT EXISTS asset_dependency_update AFTER UPDATE ON asset_dependency
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE asset_dependency SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS asset_dependency_delete AFTER DELETE ON asset_dependency
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'asset_dependency', 0);
END;

CREATE TABLE IF NOT EXISTS "status" (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT UNIQUE NOT NULL COLLATE NOCASE,
    short_name TEXT UNIQUE NOT NULL COLLATE NOCASE,
    color TEXT NOT NULL DEFAULT '#cccccc',
    synced BOOLEAN DEFAULT 0 NOT NULL,
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS status_update AFTER UPDATE ON status
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE status SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS status_delete AFTER DELETE ON status
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'status', 0);
END;

CREATE TABLE IF NOT EXISTS asset_tag (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    asset_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (asset_id) REFERENCES asset(id),
    FOREIGN KEY (tag_id) REFERENCES tag(id),
    UNIQUE (asset_id, tag_id)
);

CREATE TRIGGER IF NOT EXISTS asset_tag_update AFTER UPDATE ON asset_tag
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE asset_tag SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS asset_tag_delete AFTER DELETE ON asset_tag
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'asset_tag', 0);
END;

CREATE TABLE IF NOT EXISTS asset_checkpoint (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    mtime INTEGER NOT NULL,
    asset_id TEXT NOT NULL,
    xxhash_checksum TEXT NOT NULL,
    time_modified INTEGER NOT NULL,
    file_size INTEGER NOT NULL,
    chunks TEXT NOT NULL,
    comment TEXT DEFAULT '' NOT NULL,
    author_id TEXT NOT NULL,
    group_id TEXT DEFAULT '' NOT NULL,
    preview_id TEXT DEFAULT '' NOT NULL,
    trashed BOOLEAN DEFAULT 0 NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (preview_id) REFERENCES preview(hash),
    FOREIGN KEY (asset_id) REFERENCES asset(id),
    FOREIGN KEY (author_id) REFERENCES user(id)
);

CREATE TRIGGER IF NOT EXISTS asset_checkpoint_update AFTER UPDATE ON asset_checkpoint
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE asset_checkpoint SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS asset_checkpoint_delete AFTER DELETE ON asset_checkpoint
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'asset_checkpoint', 0);
END;

CREATE TABLE IF NOT EXISTS chunk (
    hash TEXT PRIMARY KEY NOT NULL,
    data BLOB NOT NULL,
    size INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS "role" (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT UNIQUE NOT NULL COLLATE NOCASE,
    synced BOOLEAN DEFAULT 0 NOT NULL,

    view_collection BOOLEAN DEFAULT FALSE NOT NULL,
    create_collection BOOLEAN DEFAULT FALSE NOT NULL,
    update_collection BOOLEAN DEFAULT FALSE NOT NULL,
    delete_collection BOOLEAN DEFAULT FALSE NOT NULL,

    view_asset BOOLEAN DEFAULT FALSE NOT NULL,
    create_asset BOOLEAN DEFAULT FALSE NOT NULL,
    update_asset BOOLEAN DEFAULT FALSE NOT NULL,
    delete_asset BOOLEAN DEFAULT FALSE NOT NULL,
    
    view_template BOOLEAN DEFAULT FALSE NOT NULL,
    create_template BOOLEAN DEFAULT FALSE NOT NULL,
    update_template BOOLEAN DEFAULT FALSE NOT NULL,
    delete_template BOOLEAN DEFAULT FALSE NOT NULL,
    
	view_checkpoint BOOLEAN DEFAULT FALSE NOT NULL,
	create_checkpoint BOOLEAN DEFAULT FALSE NOT NULL,
	delete_checkpoint BOOLEAN DEFAULT FALSE NOT NULL,

    pull_chunk BOOLEAN DEFAULT FALSE NOT NULL,

    assign_asset BOOLEAN DEFAULT FALSE NOT NULL,
    unassign_asset BOOLEAN DEFAULT FALSE NOT NULL,

    add_user BOOLEAN DEFAULT FALSE NOT NULL,
    remove_user BOOLEAN DEFAULT FALSE NOT NULL,
    change_role BOOLEAN DEFAULT FALSE NOT NULL,


    change_status BOOLEAN DEFAULT FALSE NOT NULL,
    set_done_asset BOOLEAN DEFAULT FALSE NOT NULL,
    set_retake_asset BOOLEAN DEFAULT FALSE NOT NULL,

    view_done_asset BOOLEAN DEFAULT FALSE NOT NULL,

    manage_dependencies BOOLEAN DEFAULT FALSE NOT NULL,
    
    CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS role_update AFTER UPDATE ON role
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE role SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS role_delete AFTER DELETE ON role
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'role', 0);
END;

CREATE TABLE IF NOT EXISTS user (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    added_at DATETIME NOT NULL,
    first_name TEXT COLLATE NOCASE,
    last_name TEXT COLLATE NOCASE,
    username TEXT UNIQUE COLLATE NOCASE,
    email TEXT NOT NULL UNIQUE COLLATE NOCASE,
    photo BLOB,
    role_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (role_id) REFERENCES role(id),
	CHECK( typeof(first_name)='text' AND length(first_name)>=1),
	CHECK( typeof(last_name)='text' AND length(last_name)>=1),
	CHECK( typeof(username)='text' AND length(username)>=1),
	CHECK( typeof(email)='text' AND length(email)>=1)
);

CREATE TRIGGER IF NOT EXISTS user_update AFTER UPDATE ON user
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE user SET synced = 0 WHERE id = NEW.id;
END;

DROP TRIGGER IF EXISTS user_delete;

CREATE TRIGGER IF NOT EXISTS user_delete AFTER DELETE ON user
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced)
    SELECT OLD.id, unixepoch(), 'user', 0
    WHERE NOT EXISTS (
        SELECT 1 FROM tomb WHERE id = OLD.id
    );
END;

CREATE TABLE IF NOT EXISTS tomb (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    table_name NOT NULL COLLATE NOCASE,
    synced BOOLEAN DEFAULT 0 NOT NULL
);

DROP VIEW IF EXISTS collection_hierarchy;

CREATE VIEW collection_hierarchy AS
WITH RECURSIVE collection_hierarchy_cte AS (
    SELECT 
        id, 
        name, 
        parent_id, 
        '/' || name || '/' AS collection_path
    FROM 
        collection 
    WHERE 
        parent_id = '' OR parent_id IS NULL 

    UNION ALL

    SELECT 
        e.id, 
        e.name, 
        e.parent_id, 
        eh.collection_path || e.name || '/' AS collection_path
    FROM 
        collection e
    JOIN 
        collection_hierarchy_cte eh ON e.parent_id = eh.id
)
SELECT * FROM collection_hierarchy_cte;

DROP VIEW IF EXISTS collection_assignees;
CREATE VIEW collection_assignees AS
SELECT 
    collection_assignee.collection_id,
    json_group_array(collection_assignee.assignee_id) AS assignee_ids
FROM 
    collection_assignee
GROUP BY 
    collection_assignee.collection_id;

DROP VIEW IF EXISTS full_collection;
CREATE VIEW full_collection AS
SELECT 
    collection.*,
    collection_type.name AS collection_type_name,
    collection_type.icon AS collection_type_icon,
    preview.preview AS preview,
    IFNULL(ea.assignee_ids, '[]') as assignee_ids
FROM 
    collection
LEFT JOIN 
    preview ON collection.preview_id = preview.hash 
JOIN 
    collection_type ON collection.collection_type_id = collection_type.id
LEFT JOIN
    collection_assignees ea ON collection.id = ea.collection_id;

DROP VIEW IF EXISTS asset_assignees;
CREATE VIEW asset_assignees AS
SELECT 
    asset.id AS asset_id,
    COALESCE(assignee.first_name, '') || ' ' || COALESCE(assignee.last_name, '') as assignee_name,
    IFNULL(assignee.email, '') as assignee_email,
    COALESCE(assigner.first_name, '') || ' ' || COALESCE(assigner.last_name, '') as assigner_name,
    IFNULL(assigner.email, '') as assigner_email
FROM 
    asset
LEFT JOIN 
    user assignee ON asset.assignee_id = assignee.id
LEFT JOIN 
    user assigner ON asset.assigner_id = assigner.id;

DROP VIEW IF EXISTS asset_tags;
CREATE VIEW asset_tags AS
SELECT 
    asset_tag.asset_id,
    json_group_array(json_object(
        'id', tag.id,
        'name', tag.name
    )) AS tags
FROM 
    asset_tag
LEFT JOIN 
    tag ON asset_tag.tag_id = tag.id
GROUP BY 
    asset_tag.asset_id;

DROP VIEW IF EXISTS asset_dependencies;
CREATE VIEW asset_dependencies AS
SELECT 
    ad.asset_id,
    json_group_array(json_object(
        'id', ad.dependency_id,
        'type_id', ad.dependency_type_id,
        'type_name', dt.name
    )) AS dependencies
FROM 
    asset_dependency ad
LEFT JOIN 
    dependency_type dt ON ad.dependency_type_id = dt.id
GROUP BY 
    ad.asset_id;

DROP VIEW IF EXISTS asset_collection_dependencies;
CREATE VIEW asset_collection_dependencies AS
SELECT 
    cd.asset_id,
    json_group_array(json_object(
        'id', cd.dependency_id,
        'type_id', cd.dependency_type_id,
        'type_name', dt.name
    )) AS collection_dependencies
FROM 
    collection_dependency cd
LEFT JOIN 
    dependency_type dt ON cd.dependency_type_id = dt.id
GROUP BY 
    cd.asset_id;

-- 2. Improved main full_asset view
DROP VIEW IF EXISTS full_asset;
CREATE VIEW full_asset AS
WITH asset_base AS (
    SELECT 
        t.*,
        tt.icon AS asset_type_icon,
        tt.name AS asset_type_name,
        IFNULL(e.name, '') AS collection_name,
        IFNULL(p.extension, '') AS preview_extension,
        p.preview,
        CASE 
            WHEN IFNULL(e.collection_path, '') = '' THEN '/' || t.name 
            ELSE e.collection_path || t.name 
        END AS asset_path,
        IFNULL(e.collection_path, '') AS collection_path,
        -- Include user data directly here, only when needed
        CASE WHEN t.assignee_id != '' THEN 
            COALESCE(assignee.first_name, '') || ' ' || COALESCE(assignee.last_name, '') 
            ELSE '' END as assignee_name,
        CASE WHEN t.assignee_id != '' THEN IFNULL(assignee.email, '') ELSE '' END as assignee_email,
        CASE WHEN t.assigner_id != '' THEN 
            COALESCE(assigner.first_name, '') || ' ' || COALESCE(assigner.last_name, '') 
            ELSE '' END as assigner_name,
        CASE WHEN t.assigner_id != '' THEN IFNULL(assigner.email, '') ELSE '' END as assigner_email
    FROM 
        asset t
    JOIN 
        asset_type tt ON t.asset_type_id = tt.id
    LEFT JOIN 
        preview p ON t.preview_id = p.hash 
    LEFT JOIN  
        collection e ON t.collection_id = e.id
    LEFT JOIN 
        user assignee ON t.assignee_id != '' AND t.assignee_id = assignee.id
    LEFT JOIN 
        user assigner ON t.assigner_id != '' AND t.assigner_id = assigner.id
)
SELECT 
    tb.*,
    IFNULL(tt.tags, '[]') as tags,
    IFNULL(td.dependencies, '[]') as dependencies,
    IFNULL(tcd.collection_dependencies, '[]') as collection_dependencies
FROM 
    asset_base tb
LEFT JOIN 
    asset_tags tt ON tb.id = tt.asset_id
LEFT JOIN 
    asset_dependencies td ON tb.id = td.asset_id
LEFT JOIN 
    asset_collection_dependencies tcd ON tb.id = tcd.asset_id;


CREATE INDEX IF NOT EXISTS idx_asset_assignee ON asset(assignee_id);
CREATE INDEX IF NOT EXISTS idx_asset_assigner ON asset(assigner_id);
CREATE INDEX IF NOT EXISTS idx_asset_collection ON asset(collection_id);
CREATE INDEX IF NOT EXISTS idx_asset_preview ON asset(preview_id);
CREATE INDEX IF NOT EXISTS idx_asset_type ON asset(asset_type_id);
CREATE INDEX IF NOT EXISTS idx_asset_tag_asset ON asset_tag(asset_id);
CREATE INDEX IF NOT EXISTS idx_asset_tag_tag ON asset_tag(tag_id);
CREATE INDEX IF NOT EXISTS idx_asset_dependency_asset ON asset_dependency(asset_id);
CREATE INDEX IF NOT EXISTS idx_collection_dependency_asset ON collection_dependency(asset_id);
CREATE INDEX IF NOT EXISTS idx_collection_parent ON collection(parent_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- INTEGRATION TABLES
-- External integration mappings (Kitsu, ClickUp, ShotGrid, etc.)
-- ═══════════════════════════════════════════════════════════════════════════

-- Project integration link: which external project is this Clustta project linked to?
-- CONSTRAINT: Only ONE row allowed (one integration per project)
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

-- Collection mappings: external hierarchy items → Clustta Collections
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
    FOREIGN KEY (collection_id) REFERENCES collection(id) ON DELETE SET NULL
);

-- Asset mappings: external assets → Clustta Assets
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
    FOREIGN KEY (asset_id) REFERENCES asset(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_integration_collection_mapping_collection ON integration_collection_mapping(collection_id);
CREATE INDEX IF NOT EXISTS idx_integration_collection_mapping_external ON integration_collection_mapping(integration_id, external_id);
CREATE INDEX IF NOT EXISTS idx_integration_asset_mapping_asset ON integration_asset_mapping(asset_id);
CREATE INDEX IF NOT EXISTS idx_integration_asset_mapping_external ON integration_asset_mapping(integration_id, external_id);

-- Integration triggers to set synced = 0 on mtime change
CREATE TRIGGER IF NOT EXISTS integration_project_update AFTER UPDATE ON integration_project
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE integration_project SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS integration_collection_mapping_update AFTER UPDATE ON integration_collection_mapping
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE integration_collection_mapping SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS integration_asset_mapping_update AFTER UPDATE ON integration_asset_mapping
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE integration_asset_mapping SET synced = 0 WHERE id = NEW.id;
END;