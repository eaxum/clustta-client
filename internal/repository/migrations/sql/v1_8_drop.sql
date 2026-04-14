-- Drop old-name (entity/task) views
DROP VIEW IF EXISTS entity_hierarchy;
DROP VIEW IF EXISTS entity_assignees;
DROP VIEW IF EXISTS full_entity;
DROP VIEW IF EXISTS task_assignees;
DROP VIEW IF EXISTS task_tags;
DROP VIEW IF EXISTS task_dependencies;
DROP VIEW IF EXISTS task_entity_dependencies;
DROP VIEW IF EXISTS full_task;

-- Drop new-name (collection/asset) views 
DROP VIEW IF EXISTS collection_hierarchy;
DROP VIEW IF EXISTS collection_assignees;
DROP VIEW IF EXISTS full_collection;
DROP VIEW IF EXISTS asset_assignees;
DROP VIEW IF EXISTS asset_tags;
DROP VIEW IF EXISTS asset_dependencies;
DROP VIEW IF EXISTS asset_collection_dependencies;
DROP VIEW IF EXISTS full_asset;

-- Drop old-name triggers
DROP TRIGGER IF EXISTS entity_update;
DROP TRIGGER IF EXISTS entity_delete;
DROP TRIGGER IF EXISTS entity_path_insert;
DROP TRIGGER IF EXISTS entity_path_update;
DROP TRIGGER IF EXISTS entity_assignee_update;
DROP TRIGGER IF EXISTS entity_assignee_delete;
DROP TRIGGER IF EXISTS entity_type_update;
DROP TRIGGER IF EXISTS entity_type_delete;
DROP TRIGGER IF EXISTS task_update;
DROP TRIGGER IF EXISTS task_delete;
DROP TRIGGER IF EXISTS task_type_update;
DROP TRIGGER IF EXISTS task_type_delete;
DROP TRIGGER IF EXISTS entity_dependency_update;
DROP TRIGGER IF EXISTS entity_dependency_delete;
DROP TRIGGER IF EXISTS task_dependency_update;
DROP TRIGGER IF EXISTS task_dependency_delete;
DROP TRIGGER IF EXISTS task_tag_update;
DROP TRIGGER IF EXISTS task_tag_delete;
DROP TRIGGER IF EXISTS task_checkpoint_update;
DROP TRIGGER IF EXISTS task_checkpoint_delete;
DROP TRIGGER IF EXISTS workflow_entity_update;
DROP TRIGGER IF EXISTS workflow_entity_delete;
DROP TRIGGER IF EXISTS workflow_task_update;
DROP TRIGGER IF EXISTS workflow_task_delete;

-- Drop new-name triggers 
DROP TRIGGER IF EXISTS collection_update;
DROP TRIGGER IF EXISTS collection_delete;
DROP TRIGGER IF EXISTS collection_path_insert;
DROP TRIGGER IF EXISTS collection_path_update;
DROP TRIGGER IF EXISTS collection_assignee_update;
DROP TRIGGER IF EXISTS collection_assignee_delete;
DROP TRIGGER IF EXISTS collection_type_update;
DROP TRIGGER IF EXISTS collection_type_delete;
DROP TRIGGER IF EXISTS collection_dependency_update;
DROP TRIGGER IF EXISTS collection_dependency_delete;
DROP TRIGGER IF EXISTS asset_update;
DROP TRIGGER IF EXISTS asset_delete;
DROP TRIGGER IF EXISTS asset_type_update;
DROP TRIGGER IF EXISTS asset_type_delete;
DROP TRIGGER IF EXISTS asset_dependency_update;
DROP TRIGGER IF EXISTS asset_dependency_delete;
DROP TRIGGER IF EXISTS asset_tag_update;
DROP TRIGGER IF EXISTS asset_tag_delete;
DROP TRIGGER IF EXISTS asset_checkpoint_update;
DROP TRIGGER IF EXISTS asset_checkpoint_delete;
DROP TRIGGER IF EXISTS workflow_collection_update;
DROP TRIGGER IF EXISTS workflow_collection_delete;
DROP TRIGGER IF EXISTS workflow_asset_update;
DROP TRIGGER IF EXISTS workflow_asset_delete;

-- Drop old-name indexes
DROP INDEX IF EXISTS idx_task_assignee;
DROP INDEX IF EXISTS idx_task_assigner;
DROP INDEX IF EXISTS idx_task_entity;
DROP INDEX IF EXISTS idx_task_preview;
DROP INDEX IF EXISTS idx_task_type;
DROP INDEX IF EXISTS idx_task_tag_task;
DROP INDEX IF EXISTS idx_task_tag_tag;
DROP INDEX IF EXISTS idx_task_dependency_task;
DROP INDEX IF EXISTS idx_entity_dependency_task;
DROP INDEX IF EXISTS idx_entity_parent;
DROP INDEX IF EXISTS idx_entity_path;

-- Drop new-name indexes 
DROP INDEX IF EXISTS idx_asset_assignee;
DROP INDEX IF EXISTS idx_asset_assigner;
DROP INDEX IF EXISTS idx_asset_collection;
DROP INDEX IF EXISTS idx_asset_preview;
DROP INDEX IF EXISTS idx_asset_type;
DROP INDEX IF EXISTS idx_asset_tag_asset;
DROP INDEX IF EXISTS idx_asset_tag_tag;
DROP INDEX IF EXISTS idx_asset_dependency_asset;
DROP INDEX IF EXISTS idx_collection_dependency_asset;
DROP INDEX IF EXISTS idx_collection_parent;
