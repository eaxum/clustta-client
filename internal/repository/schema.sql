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

CREATE TABLE IF NOT EXISTS workflow_entity (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL COLLATE NOCASE,
    workflow_id TEXT NOT NULL,
    entity_type_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (workflow_id) REFERENCES workflow(id),
    FOREIGN KEY (entity_type_id) REFERENCES entity_type(id),
    UNIQUE (name, workflow_id),
    CHECK( typeof(workflow_id)='text' AND length(workflow_id)>=1),
    CHECK( typeof(entity_type_id)='text' AND length(entity_type_id)>=1),
    CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS workflow_entity_update AFTER UPDATE ON workflow_entity
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE workflow_entity SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS workflow_entity_delete AFTER DELETE ON workflow_entity
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'workflow_entity', 0);
END;

CREATE TABLE IF NOT EXISTS workflow_task (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL COLLATE NOCASE,
    workflow_id TEXT NOT NULL,
    is_resource BOOLEAN DEFAULT 0 NOT NULL,
	is_link BOOLEAN DEFAULT 0 NOT NULL,
	pointer TEXT DEFAULT '' NOT NULL,
    template_id TEXT NOT NULL,
    task_type_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (workflow_id) REFERENCES workflow(id),
    FOREIGN KEY (template_id) REFERENCES template(id),
    FOREIGN KEY (task_type_id) REFERENCES task_type(id),
    UNIQUE (name, workflow_id),
	CHECK( typeof(workflow_id)='text' AND length(workflow_id)>=1),
	CHECK( typeof(template_id)='text' AND length(template_id)>=1),
	CHECK( typeof(task_type_id)='text' AND length(task_type_id)>=1),
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS workflow_task_update AFTER UPDATE ON workflow_task
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE workflow_task SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS workflow_task_delete AFTER DELETE ON workflow_task
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'workflow_task', 0);
END;

CREATE TABLE IF NOT EXISTS workflow_link (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL COLLATE NOCASE,
    entity_type_id TEXT NOT NULL,
    workflow_id TEXT NOT NULL,
    linked_workflow_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (workflow_id) REFERENCES workflow(id),
    FOREIGN KEY (linked_workflow_id) REFERENCES workflow(id),
    FOREIGN KEY (entity_type_id) REFERENCES entity_type(id),
    UNIQUE (workflow_id, linked_workflow_id, name),
    CHECK( typeof(entity_type_id)='text' AND length(entity_type_id)>=1)
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

CREATE TABLE IF NOT EXISTS entity (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL COLLATE NOCASE,
    entity_path TEXT DEFAULT '' NOT NULL,
    description TEXT,
    entity_type_id TEXT NOT NULL,
    parent_id TEXT NOT NULL,
	trashed BOOLEAN DEFAULT 0 NOT NULL,
    preview_id TEXT DEFAULT '' NOT NULL,
	synced BOOLEAN DEFAULT 0 NOT NULL,
	is_library BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (entity_type_id) REFERENCES entity_type(id),
    FOREIGN KEY (parent_id) REFERENCES entity(id),
    FOREIGN KEY (preview_id) REFERENCES preview(hash),
    UNIQUE (name, parent_id),
	CHECK( typeof(entity_type_id)='text' AND length(entity_type_id)>=1),
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS entity_update AFTER UPDATE ON entity
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE entity SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS entity_delete AFTER DELETE ON entity
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'entity', 0);
END;

-- Trigger to maintain materialized path on INSERT
CREATE TRIGGER IF NOT EXISTS entity_path_insert 
AFTER INSERT ON entity
FOR EACH ROW
BEGIN
    UPDATE entity
    SET entity_path = 
        CASE
        WHEN NEW.parent_id = '' OR NEW.parent_id IS NULL THEN '/' || NEW.name || '/'
        ELSE (
            SELECT entity_path || NEW.name || '/' FROM entity WHERE id = NEW.parent_id
        )
        END
    WHERE id = NEW.id;
END;

-- Updated entity path trigger to handle orphaned entities
CREATE TRIGGER IF NOT EXISTS entity_path_update 
AFTER UPDATE OF name, parent_id ON entity
FOR EACH ROW
WHEN OLD.name != NEW.name OR OLD.parent_id != NEW.parent_id
BEGIN
    -- Recalculate this entity's path
  UPDATE entity
  SET entity_path =
    CASE
      WHEN NEW.parent_id IS NULL THEN '/' || NEW.name || '/'
      ELSE COALESCE(
        (SELECT entity_path || NEW.name || '/' FROM entity WHERE id = NEW.parent_id),
        '/' || NEW.name || '/'
      )
    END
  WHERE id = NEW.id;

  -- Recalculate all descendant paths
  UPDATE entity
  SET entity_path =
    (SELECT entity_path FROM entity WHERE id = NEW.id) || substr(entity_path, length(OLD.entity_path) + 1)
  WHERE entity_path LIKE OLD.entity_path || '%'
    AND id != NEW.id;
END;


CREATE TABLE IF NOT EXISTS entity_assignee (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    entity_id TEXT NOT NULL,
    assignee_id TEXT DEFAULT '' NOT NULL,
    assigner_id TEXT DEFAULT '' NOT NULL,
	synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (entity_id) REFERENCES entity(id),
    FOREIGN KEY (assignee_id) REFERENCES user(id),
    FOREIGN KEY (assigner_id) REFERENCES user(id),
    UNIQUE (entity_id, assignee_id),
	CHECK( typeof(entity_id)='text' AND length(entity_id)>=1)
);

CREATE TRIGGER IF NOT EXISTS entity_assignee_update AFTER UPDATE ON entity_assignee
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE entity_assignee SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS entity_assignee_delete AFTER DELETE ON entity_assignee
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'entity_assignee', 0);
END;

CREATE TABLE IF NOT EXISTS entity_type (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL UNIQUE COLLATE NOCASE,
    icon TEXT NOT NULL UNIQUE COLLATE NOCASE,
    synced BOOLEAN DEFAULT 0 NOT NULL,
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS entity_type_update AFTER UPDATE ON entity_type
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE entity_type SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS entity_type_delete AFTER DELETE ON entity_type
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'entity_type', 0);
END;

CREATE TABLE IF NOT EXISTS task (
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
    task_type_id TEXT NOT NULL,
    entity_id TEXT DEFAULT '' NOT NULL,
	assignee_id TEXT DEFAULT '' NOT NULL,
	assigner_id TEXT DEFAULT '' NOT NULL,
    preview_id TEXT DEFAULT '' NOT NULL,
    trashed BOOLEAN DEFAULT 0 NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (preview_id) REFERENCES preview(hash),
    FOREIGN KEY (status_id) REFERENCES status(id),
    FOREIGN KEY (task_type_id) REFERENCES task_type(id),
    FOREIGN KEY (entity_id) REFERENCES entity(id),
	FOREIGN KEY (assignee_id) REFERENCES user(id),
	FOREIGN KEY (assigner_id) REFERENCES user(id),
    UNIQUE (name, entity_id, extension),
	CHECK( typeof(task_type_id)='text' AND length(task_type_id)>=1),
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS task_update AFTER UPDATE ON task
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE task SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS task_delete AFTER DELETE ON task
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'task', 0);
END;

CREATE TABLE IF NOT EXISTS task_type (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL UNIQUE COLLATE NOCASE,
    icon TEXT NOT NULL UNIQUE COLLATE NOCASE,
    synced BOOLEAN DEFAULT 0 NOT NULL,
	CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TRIGGER IF NOT EXISTS task_type_update AFTER UPDATE ON task_type
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE task_type SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS task_type_delete AFTER DELETE ON task_type
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'task_type', 0);
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

CREATE TABLE IF NOT EXISTS entity_dependency (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    task_id TEXT NOT NULL,
    dependency_id TEXT NOT NULL,
    dependency_type_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (task_id) REFERENCES task(id),
    FOREIGN KEY (dependency_id) REFERENCES entity(id),
    FOREIGN KEY (dependency_type_id) REFERENCES dependency_type(id),
    UNIQUE (task_id, dependency_id)
);

CREATE TRIGGER IF NOT EXISTS entity_dependency_update AFTER UPDATE ON entity_dependency
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE entity_dependency SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS entity_dependency_delete AFTER DELETE ON entity_dependency
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'entity_dependency', 0);
END;

CREATE TABLE IF NOT EXISTS task_dependency (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    task_id TEXT NOT NULL,
    dependency_id TEXT NOT NULL,
    dependency_type_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (task_id) REFERENCES task(id),
    FOREIGN KEY (dependency_id) REFERENCES task(id),
    FOREIGN KEY (dependency_type_id) REFERENCES dependency_type(id),
    UNIQUE (task_id, dependency_id)
);

CREATE TRIGGER IF NOT EXISTS task_dependency_update AFTER UPDATE ON task_dependency
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE task_dependency SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS task_dependency_delete AFTER DELETE ON task_dependency
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'task_dependency', 0);
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

CREATE TABLE IF NOT EXISTS task_tag (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    task_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    synced BOOLEAN DEFAULT 0 NOT NULL,
    FOREIGN KEY (task_id) REFERENCES task(id),
    FOREIGN KEY (tag_id) REFERENCES tag(id),
    UNIQUE (task_id, tag_id)
);

CREATE TRIGGER IF NOT EXISTS task_tag_update AFTER UPDATE ON task_tag
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE task_tag SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS task_tag_delete AFTER DELETE ON task_tag
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'task_tag', 0);
END;

CREATE TABLE IF NOT EXISTS task_checkpoint (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    mtime INTEGER NOT NULL,
    task_id TEXT NOT NULL,
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
    FOREIGN KEY (task_id) REFERENCES task(id),
    FOREIGN KEY (author_id) REFERENCES user(id)
);

CREATE TRIGGER IF NOT EXISTS task_checkpoint_update AFTER UPDATE ON task_checkpoint
FOR EACH ROW
WHEN OLD.mtime != NEW.mtime
BEGIN
    UPDATE task_checkpoint SET synced = 0 WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS task_checkpoint_delete AFTER DELETE ON task_checkpoint
FOR EACH ROW
BEGIN
    INSERT INTO tomb (id, mtime, table_name, synced) VALUES (OLD.id, unixepoch(), 'task_checkpoint', 0);
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

    view_entity BOOLEAN DEFAULT FALSE NOT NULL,
    create_entity BOOLEAN DEFAULT FALSE NOT NULL,
    update_entity BOOLEAN DEFAULT FALSE NOT NULL,
    delete_entity BOOLEAN DEFAULT FALSE NOT NULL,

    view_task BOOLEAN DEFAULT FALSE NOT NULL,
    create_task BOOLEAN DEFAULT FALSE NOT NULL,
    update_task BOOLEAN DEFAULT FALSE NOT NULL,
    delete_task BOOLEAN DEFAULT FALSE NOT NULL,
    
    view_template BOOLEAN DEFAULT FALSE NOT NULL,
    create_template BOOLEAN DEFAULT FALSE NOT NULL,
    update_template BOOLEAN DEFAULT FALSE NOT NULL,
    delete_template BOOLEAN DEFAULT FALSE NOT NULL,
    
	view_checkpoint BOOLEAN DEFAULT FALSE NOT NULL,
	create_checkpoint BOOLEAN DEFAULT FALSE NOT NULL,
	delete_checkpoint BOOLEAN DEFAULT FALSE NOT NULL,

    pull_chunk BOOLEAN DEFAULT FALSE NOT NULL,

    assign_task BOOLEAN DEFAULT FALSE NOT NULL,
    unassign_task BOOLEAN DEFAULT FALSE NOT NULL,

    add_user BOOLEAN DEFAULT FALSE NOT NULL,
    remove_user BOOLEAN DEFAULT FALSE NOT NULL,
    change_role BOOLEAN DEFAULT FALSE NOT NULL,


    change_status BOOLEAN DEFAULT FALSE NOT NULL,
    set_done_task BOOLEAN DEFAULT FALSE NOT NULL,
    set_retake_task BOOLEAN DEFAULT FALSE NOT NULL,

    view_done_task BOOLEAN DEFAULT FALSE NOT NULL,

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

DROP VIEW IF EXISTS entity_hierarchy;

CREATE VIEW entity_hierarchy AS
WITH RECURSIVE entity_hierarchy_cte AS (
    SELECT 
        id, 
        name, 
        parent_id, 
        '/' || name || '/' AS entity_path
    FROM 
        entity 
    WHERE 
        parent_id = '' OR parent_id IS NULL 

    UNION ALL

    SELECT 
        e.id, 
        e.name, 
        e.parent_id, 
        eh.entity_path || e.name || '/' AS entity_path
    FROM 
        entity e
    JOIN 
        entity_hierarchy_cte eh ON e.parent_id = eh.id
)
SELECT * FROM entity_hierarchy_cte;

DROP VIEW IF EXISTS entity_assignees;
CREATE VIEW entity_assignees AS
SELECT 
    entity_assignee.entity_id,
    json_group_array(entity_assignee.assignee_id) AS assignee_ids
FROM 
    entity_assignee
GROUP BY 
    entity_assignee.entity_id;

DROP VIEW IF EXISTS full_entity;
CREATE VIEW full_entity AS
SELECT 
    entity.*,
    entity_type.name AS entity_type_name,
    entity_type.icon AS entity_type_icon,
    preview.preview AS preview,
    IFNULL(ea.assignee_ids, '[]') as assignee_ids
FROM 
    entity
LEFT JOIN 
    preview ON entity.preview_id = preview.hash 
JOIN 
    entity_type ON entity.entity_type_id = entity_type.id
LEFT JOIN
    entity_assignees ea ON entity.id = ea.entity_id;

DROP VIEW IF EXISTS task_assignees;
CREATE VIEW task_assignees AS
SELECT 
    task.id AS task_id,
    COALESCE(assignee.first_name, '') || ' ' || COALESCE(assignee.last_name, '') as assignee_name,
    IFNULL(assignee.email, '') as assignee_email,
    COALESCE(assigner.first_name, '') || ' ' || COALESCE(assigner.last_name, '') as assigner_name,
    IFNULL(assigner.email, '') as assigner_email
FROM 
    task
LEFT JOIN 
    user assignee ON task.assignee_id = assignee.id
LEFT JOIN 
    user assigner ON task.assigner_id = assigner.id;

DROP VIEW IF EXISTS task_tags;
CREATE VIEW task_tags AS
SELECT 
    task_tag.task_id,
    json_group_array(json_object(
        'id', tag.id,
        'name', tag.name
    )) AS tags
FROM 
    task_tag
LEFT JOIN 
    tag ON task_tag.tag_id = tag.id
GROUP BY 
    task_tag.task_id;

DROP VIEW IF EXISTS task_dependencies;
CREATE VIEW task_dependencies AS
SELECT 
    td.task_id,
    json_group_array(json_object(
        'id', td.dependency_id,
        'type_id', td.dependency_type_id,
        'type_name', dt.name
    )) AS dependencies
FROM 
    task_dependency td
LEFT JOIN 
    dependency_type dt ON td.dependency_type_id = dt.id
GROUP BY 
    td.task_id;

DROP VIEW IF EXISTS task_entity_dependencies;
CREATE VIEW task_entity_dependencies AS
SELECT 
    ed.task_id,
    json_group_array(json_object(
        'id', ed.dependency_id,
        'type_id', ed.dependency_type_id,
        'type_name', dt.name
    )) AS entity_dependencies
FROM 
    entity_dependency ed
LEFT JOIN 
    dependency_type dt ON ed.dependency_type_id = dt.id
GROUP BY 
    ed.task_id;

-- 2. Improved main full_task view
DROP VIEW IF EXISTS full_task;
CREATE VIEW full_task AS
WITH task_base AS (
    SELECT 
        t.*,
        tt.icon AS task_type_icon,
        tt.name AS task_type_name,
        IFNULL(e.name, '') AS entity_name,
        IFNULL(p.extension, '') AS preview_extension,
        p.preview,
        CASE 
            WHEN IFNULL(e.entity_path, '') = '' THEN '/' || t.name 
            ELSE e.entity_path || t.name 
        END AS task_path,
        IFNULL(e.entity_path, '') AS entity_path,
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
        task t
    JOIN 
        task_type tt ON t.task_type_id = tt.id
    LEFT JOIN 
        preview p ON t.preview_id = p.hash 
    LEFT JOIN  
        entity e ON t.entity_id = e.id
    LEFT JOIN 
        user assignee ON t.assignee_id != '' AND t.assignee_id = assignee.id
    LEFT JOIN 
        user assigner ON t.assigner_id != '' AND t.assigner_id = assigner.id
)
SELECT 
    tb.*,
    IFNULL(tt.tags, '[]') as tags,
    IFNULL(td.dependencies, '[]') as dependencies,
    IFNULL(ted.entity_dependencies, '[]') as entity_dependencies
FROM 
    task_base tb
LEFT JOIN 
    task_tags tt ON tb.id = tt.task_id
LEFT JOIN 
    task_dependencies td ON tb.id = td.task_id
LEFT JOIN 
    task_entity_dependencies ted ON tb.id = ted.task_id;


CREATE INDEX IF NOT EXISTS idx_task_assignee ON task(assignee_id);
CREATE INDEX IF NOT EXISTS idx_task_assigner ON task(assigner_id);
CREATE INDEX IF NOT EXISTS idx_task_entity ON task(entity_id);
CREATE INDEX IF NOT EXISTS idx_task_preview ON task(preview_id);
CREATE INDEX IF NOT EXISTS idx_task_type ON task(task_type_id);
CREATE INDEX IF NOT EXISTS idx_task_tag_task ON task_tag(task_id);
CREATE INDEX IF NOT EXISTS idx_task_tag_tag ON task_tag(tag_id);
CREATE INDEX IF NOT EXISTS idx_task_dependency_task ON task_dependency(task_id);
CREATE INDEX IF NOT EXISTS idx_entity_dependency_task ON entity_dependency(task_id);
CREATE INDEX IF NOT EXISTS idx_entity_parent ON entity(parent_id);