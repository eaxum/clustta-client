import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
}

/**
 * Compute entity_path based on parent's path and entity name
 * Mirrors the SQLite trigger logic from schema.sql
 * @param {object} db - Database connection
 * @param {string} parentId - Parent entity ID (empty string for root)
 * @param {string} name - Entity name
 * @returns {string} - Computed entity path like '/Parent/Child/'
 */
function computeEntityPath(db, parentId, name) {
  if (!parentId || parentId === '') {
    // Root level entity
    return '/' + name + '/';
  }
  // Get parent's entity_path
  const parent = queryOne(db, 'SELECT entity_path FROM entity WHERE id = ?', [parentId]);
  if (parent && parent.entity_path) {
    return parent.entity_path + name + '/';
  }
  // Fallback if parent not found
  return '/' + name + '/';
}

/**
 * Convert database row to entity object with proper types
 * Optionally enriches with entity type info
 */
function rowToEntity(row, entityTypeMap = {}) {
  if (!row) return null;
  const entityType = entityTypeMap[row.entity_type_id] || {};
  return {
    ...row,
    trashed: !!row.trashed,
    is_trashed: !!row.trashed,
    is_library: !!row.is_library,
    synced: !!row.synced,
    type: 'entity',
    entity_type: entityType.name || '',
    entity_type_icon: entityType.icon || '',
  };
}

/**
 * Convert database row to task object with proper types
 * Optionally enriches with status, tags, entity type info, and dependencies
 * Also computes entity_path and task_path from entity
 */
function rowToTask(row, statusMap = {}, taskTypeMap = {}, taskTagsMap = {}, tagMap = {}, taskDependenciesMap = {}, entityDependenciesMap = {}, entityMap = {}) {
  if (!row) return null;
  const status = statusMap[row.status_id] || {};
  const taskType = taskTypeMap[row.task_type_id] || {};
  const tagIds = taskTagsMap[row.id] || [];
  const tags = tagIds.map(tagId => tagMap[tagId]?.name || tagId).filter(Boolean);
  const dependencies = taskDependenciesMap[row.id] || [];
  const entity_dependencies = entityDependenciesMap[row.id] || [];
  
  // Compute entity_path and task_path from parent entity (mirrors full_task view)
  const entity = entityMap[row.entity_id] || {};
  const entityPath = entity.entity_path || '';
  const taskPath = entityPath ? entityPath + row.name : '/' + row.name;
  const entityName = entity.name || '';
  
  return {
    ...row,
    is_resource: !!row.is_resource,
    is_link: !!row.is_link,
    trashed: !!row.trashed,
    is_trashed: !!row.trashed,
    synced: !!row.synced,
    type: 'task',
    // Status object
    status: status,
    status_name: status.name || '',
    status_short_name: status.short_name || '',
    status_color: status.color || '',
    // Task type info
    task_type: taskType.name || '',
    task_type_name: taskType.name || '',
    task_type_icon: taskType.icon || '',
    // Tags array
    tags: tags,
    // Dependencies arrays
    dependencies: dependencies,
    entity_dependencies: entity_dependencies,
    // Entity-related paths
    entity_path: entityPath,
    task_path: taskPath,
    entity_name: entityName,
    // Web mode specific - no local file tracking
    file_status: 'normal',
  };
}

export const CollectionService = {
  // Returns all collections for a project
  GetCollections: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      // Build entity type lookup map
      const entityTypeRows = query(db, 'SELECT * FROM entity_type');
      const entityTypeMap = {};
      for (const et of entityTypeRows) {
        entityTypeMap[et.id] = et;
      }
      
      const rows = query(db, 'SELECT * FROM entity WHERE trashed = 0');
      return rows.map(row => rowToEntity(row, entityTypeMap));
    } catch (error) {
      console.error('GetCollections error:', error);
      return [];
    }
  },

  // Returns a specific collection by ID
  GetCollection: async (projectPath, collectionId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      // Build entity type lookup map
      const entityTypeRows = query(db, 'SELECT * FROM entity_type');
      const entityTypeMap = {};
      for (const et of entityTypeRows) {
        entityTypeMap[et.id] = et;
      }
      
      const row = queryOne(db, 'SELECT * FROM entity WHERE id = ?', [collectionId]);
      return rowToEntity(row, entityTypeMap) || {};
    } catch (error) {
      console.error('GetCollection error:', error);
      return {};
    }
  },

  // Alias for GetCollection - retrieves a collection by its ID
  GetCollectionByID: async (projectPath, entityId) => {
    // Handle root as empty string
    const normalizedId = entityId === 'root' ? '' : entityId;
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      // Build entity type lookup map
      const entityTypeRows = query(db, 'SELECT * FROM entity_type');
      const entityTypeMap = {};
      for (const et of entityTypeRows) {
        entityTypeMap[et.id] = et;
      }
      
      const row = queryOne(db, 'SELECT * FROM entity WHERE id = ?', [normalizedId]);
      return rowToEntity(row, entityTypeMap) || {};
    } catch (error) {
      console.error('GetCollectionByID error:', error);
      return {};
    }
  },

  // Returns a specific collection by its path
  GetCollectionByPath: async (projectPath, entityPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      // Build entity type lookup map
      const entityTypeRows = query(db, 'SELECT * FROM entity_type');
      const entityTypeMap = {};
      for (const et of entityTypeRows) {
        entityTypeMap[et.id] = et;
      }
      
      // Handle root path
      const normalizedPath = entityPath === '/' ? '' : entityPath;
      
      const row = queryOne(db, 'SELECT * FROM entity WHERE entity_path = ?', [normalizedPath]);
      return rowToEntity(row, entityTypeMap) || {};
    } catch (error) {
      console.error('GetCollectionByPath error:', error);
      return {};
    }
  },

  // Returns all children of a collection (entities, tasks, untracked items)
  GetCollectionChildren: async (projectPath, entityId, projectWorkingDir, entityFolderPath, ignoreList, isUntracked) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      // Build lookup maps for enrichment
      const entityTypeRows = query(db, 'SELECT * FROM entity_type');
      const entityTypeMap = {};
      for (const et of entityTypeRows) {
        entityTypeMap[et.id] = et;
      }
      
      const statusRows = query(db, 'SELECT * FROM status');
      const statusMap = {};
      for (const s of statusRows) {
        statusMap[s.id] = s;
      }
      
      const taskTypeRows = query(db, 'SELECT * FROM task_type');
      const taskTypeMap = {};
      for (const tt of taskTypeRows) {
        taskTypeMap[tt.id] = tt;
      }
      
      const tagRows = query(db, 'SELECT * FROM tag');
      const tagMap = {};
      for (const t of tagRows) {
        tagMap[t.id] = t;
      }
      
      const taskTagRows = query(db, 'SELECT * FROM task_tag');
      const taskTagsMap = {};
      for (const tt of taskTagRows) {
        if (!taskTagsMap[tt.task_id]) {
          taskTagsMap[tt.task_id] = [];
        }
        taskTagsMap[tt.task_id].push(tt.tag_id);
      }
      
      // Build task dependency map
      const taskDependencyRows = query(db, 'SELECT * FROM task_dependency');
      const taskDependenciesMap = {};
      for (const td of taskDependencyRows) {
        if (!taskDependenciesMap[td.task_id]) {
          taskDependenciesMap[td.task_id] = [];
        }
        taskDependenciesMap[td.task_id].push({
          id: td.dependency_id,
          type_id: td.dependency_type_id
        });
      }
      
      // Build entity dependency map
      const entityDependencyRows = query(db, 'SELECT * FROM entity_dependency');
      const entityDependenciesMap = {};
      for (const ed of entityDependencyRows) {
        if (!entityDependenciesMap[ed.task_id]) {
          entityDependenciesMap[ed.task_id] = [];
        }
        entityDependenciesMap[ed.task_id].push({
          id: ed.entity_id,
          type_id: ed.dependency_type_id
        });
      }
      
      // Get child entities
      const parentCondition = entityId === 'root' 
        ? "(parent_id IS NULL OR parent_id = '')"
        : `parent_id = '${entityId}'`;
      
      const entityRows = query(db, `SELECT * FROM entity WHERE ${parentCondition} AND trashed = 0`);
      const entities = entityRows.map(row => rowToEntity(row, entityTypeMap));
      
      // Build entityMap for task_path computation
      const allEntityRows = query(db, 'SELECT id, name, entity_path FROM entity');
      const entityMapForTasks = {};
      for (const e of allEntityRows) {
        entityMapForTasks[e.id] = e;
      }
      
      // Get child tasks
      const taskCondition = entityId === 'root'
        ? "(entity_id IS NULL OR entity_id = '')"
        : `entity_id = '${entityId}'`;
      
      const taskRows = query(db, `SELECT * FROM task WHERE ${taskCondition} AND trashed = 0`);
      const tasks = taskRows.map(row => rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMapForTasks));
      
      return {
        entities,
        tasks,
        untracked_tasks: [],      // No untracked items in web mode
        untracked_entities: [],   // No untracked items in web mode
      };
    } catch (error) {
      console.error('GetCollectionChildren error:', error);
      return {
        entities: [],
        tasks: [],
        untracked_tasks: [],
        untracked_entities: [],
      };
    }
  },

  // Returns state flags for a collection's children
  GetCollectionChildrenState: async (projectPath, entityId, projectWorkingDir, ignoreList) => {
    // In web mode, we don't track local file states
    return {
      modified_tasks: [],
      outdated_tasks: [],
      rebuildable_tasks: [],
      normal_tasks: [],
      untracked_files: [],
      untracked_folders: [],
    };
  },

  // Returns collection state flags (has_untracked, has_modified, has_outdated, has_rebuildable)
  GetCollectionStateFlags: async (projectPath, entityId, projectWorkingDir, ignoreList) => {
    // In web mode, we don't track local file states, so return all false
    return {
      has_untracked: false,
      has_modified: false,
      has_outdated: false,
      has_rebuildable: false,
    };
  },

  // Changes the parent of a collection (local-first approach)
  ChangeCollectionParent: async (projectPath, entityId, newParentId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      
      // Get entity's current name to recompute path
      const entity = queryOne(db, 'SELECT name, entity_path FROM entity WHERE id = ?', [entityId]);
      if (!entity) {
        throw new Error('Entity not found');
      }
      
      // Compute new entity_path based on new parent
      const newEntityPath = computeEntityPath(db, newParentId, entity.name);
      const oldEntityPath = entity.entity_path;
      
      // Update entity with new parent and path
      execute(db, 'UPDATE entity SET parent_id = ?, entity_path = ?, mtime = ?, synced = 0 WHERE id = ?', 
        [newParentId, newEntityPath, Date.now(), entityId]);
      
      // Update all descendants' entity_paths
      if (oldEntityPath) {
        const descendants = query(db, "SELECT id, name, entity_path FROM entity WHERE entity_path LIKE ? AND id != ?", 
          [oldEntityPath + '%', entityId]);
        for (const desc of descendants) {
          const newDescPath = newEntityPath + desc.entity_path.substring(oldEntityPath.length);
          execute(db, 'UPDATE entity SET entity_path = ?, mtime = ?, synced = 0 WHERE id = ?', 
            [newDescPath, Date.now(), desc.id]);
        }
      }
      
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeCollectionParent error:', error);
      throw error;
    }
  },

  // Creates a new collection (local-first approach)
  // Signature matches Wails: CreateCollection(projectPath, name, description, entityTypeId, parentId, previewPath, isLibrary)
  CreateCollection: async (projectPath, name, description, entityTypeId, parentId, previewPath, isLibrary) => {
    const projectName = getProjectName(projectPath);
    const id = crypto.randomUUID();
    const now = Date.now();
    const createdAt = new Date().toISOString();
    
    try {
      const db = await getDatabase(projectName);
      
      // Compute entity_path based on parent (mimics SQLite trigger behavior)
      const entityPath = computeEntityPath(db, parentId, name || '');
      
      execute(db, `
        INSERT INTO entity (id, created_at, mtime, name, description, entity_type_id, parent_id, entity_path, preview_id, is_library, trashed, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
      `, [id, createdAt, now, name || '', description || '', entityTypeId || '', parentId || '', entityPath, previewPath || '', isLibrary ? 1 : 0]);
      await persistDatabase(projectName);
      
      // Build entity type lookup map
      const entityTypeRows = query(db, 'SELECT * FROM entity_type');
      const entityTypeMap = {};
      for (const et of entityTypeRows) {
        entityTypeMap[et.id] = et;
      }
      
      // Return the created entity
      const row = queryOne(db, 'SELECT * FROM entity WHERE id = ?', [id]);
      return rowToEntity(row, entityTypeMap);
    } catch (error) {
      console.error('CreateCollection error:', error);
      throw error;
    }
  },

  // Updates an existing collection (local-first approach)
  UpdateCollection: async (projectPath, collection) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        UPDATE entity SET name = ?, parent_id = ?, entity_type_id = ?, entity_path = ?, mtime = ?, synced = 0
        WHERE id = ?
      `, [collection.name || '', collection.parent_id || '', collection.entity_type_id || '', collection.entity_path || '', Date.now(), collection.id]);
      await persistDatabase(projectName);
      
      // Build entity type lookup map
      const entityTypeRows = query(db, 'SELECT * FROM entity_type');
      const entityTypeMap = {};
      for (const et of entityTypeRows) {
        entityTypeMap[et.id] = et;
      }
      
      // Return the updated entity
      const row = queryOne(db, 'SELECT * FROM entity WHERE id = ?', [collection.id]);
      return rowToEntity(row, entityTypeMap);
    } catch (error) {
      console.error('UpdateCollection error:', error);
      throw error;
    }
  },

  // Deletes a collection (local-first approach)
  DeleteCollection: async (projectPath, collectionId, moveToTrash = false) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      if (moveToTrash) {
        execute(db, 'UPDATE entity SET trashed = 1, mtime = ?, synced = 0 WHERE id = ?', [Date.now(), collectionId]);
      } else {
        execute(db, 'DELETE FROM entity WHERE id = ?', [collectionId]);
      }
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteCollection error:', error);
      throw error;
    }
  },

  // Rebuilds all items in a collection
  Rebuild: async (projectUri, projectUrl, collectionId) => {
    // In web mode, rebuild is not supported (no file system)
    console.warn('Rebuild not available in web mode (no file system)');
    return {};
  },

  // Returns all collection types for a project
  GetCollectionTypes: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      return query(db, 'SELECT * FROM entity_type');
    } catch (error) {
      console.error('GetCollectionTypes error:', error);
      return [];
    }
  },

  // Creates a new collection type (local-first approach)
  CreateCollectionType: async (projectPath, type) => {
    const projectName = getProjectName(projectPath);
    const id = type.id || crypto.randomUUID();
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'INSERT INTO entity_type (id, mtime, name, icon, synced) VALUES (?, ?, ?, ?, 0)', 
        [id, Date.now(), type.name || '', type.icon || '']);
      await persistDatabase(projectName);
      
      return { id, name: type.name, icon: type.icon };
    } catch (error) {
      console.error('CreateCollectionType error:', error);
      throw error;
    }
  },

  // Updates an existing collection type (local-first approach)
  UpdateCollectionType: async (projectPath, type) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE entity_type SET name = ?, icon = ?, mtime = ?, synced = 0 WHERE id = ?',
        [type.name || '', type.icon || '', Date.now(), type.id]);
      await persistDatabase(projectName);
      
      return { id: type.id, name: type.name, icon: type.icon };
    } catch (error) {
      console.error('UpdateCollectionType error:', error);
      throw error;
    }
  },

  // Deletes a collection type (local-first approach)
  DeleteCollectionType: async (projectPath, typeId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM entity_type WHERE id = ?', [typeId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteCollectionType error:', error);
      throw error;
    }
  },

  // Renames an existing collection (local-first approach)
  RenameCollection: async (projectPath, entityId, newName) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      
      // Get entity's current info to recompute path
      const entity = queryOne(db, 'SELECT name, parent_id, entity_path FROM entity WHERE id = ?', [entityId]);
      if (!entity) {
        throw new Error('Entity not found');
      }
      
      // Compute new entity_path with the new name
      const newEntityPath = computeEntityPath(db, entity.parent_id, newName);
      const oldEntityPath = entity.entity_path;
      
      // Update entity with new name and path
      execute(db, 'UPDATE entity SET name = ?, entity_path = ?, mtime = ?, synced = 0 WHERE id = ?', 
        [newName, newEntityPath, Date.now(), entityId]);
      
      // Update all descendants' entity_paths
      if (oldEntityPath) {
        const descendants = query(db, "SELECT id, entity_path FROM entity WHERE entity_path LIKE ? AND id != ?", 
          [oldEntityPath + '%', entityId]);
        for (const desc of descendants) {
          const newDescPath = newEntityPath + desc.entity_path.substring(oldEntityPath.length);
          execute(db, 'UPDATE entity SET entity_path = ?, mtime = ?, synced = 0 WHERE id = ?', 
            [newDescPath, Date.now(), desc.id]);
        }
      }
      
      await persistDatabase(projectName);
      
      // Build entity type lookup map
      const entityTypeRows = query(db, 'SELECT * FROM entity_type');
      const entityTypeMap = {};
      for (const et of entityTypeRows) {
        entityTypeMap[et.id] = et;
      }
      
      // Return the updated entity
      const row = queryOne(db, 'SELECT * FROM entity WHERE id = ?', [entityId]);
      return rowToEntity(row, entityTypeMap);
    } catch (error) {
      console.error('RenameCollection error:', error);
      throw error;
    }
  },

  // Changes the type of a collection (local-first approach)
  ChangeType: async (projectPath, entityId, entityTypeId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE entity SET entity_type_id = ?, mtime = ?, synced = 0 WHERE id = ?', [entityTypeId, Date.now(), entityId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeType error:', error);
      throw error;
    }
  },

  // Toggles the library flag on a collection (local-first approach)
  ChangeIsLibrary: async (projectPath, entityId, isLibrary) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE entity SET is_library = ?, mtime = ?, synced = 0 WHERE id = ?', 
        [isLibrary ? 1 : 0, Date.now(), entityId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeIsLibrary error:', error);
      throw error;
    }
  },

  // Assigns a user to a collection (local-first approach)
  // Uses entity_assignee junction table
  Assign: async (projectPath, entityId, userId) => {
    const projectName = getProjectName(projectPath);
    const id = crypto.randomUUID();
    const now = Date.now();
    try {
      const db = await getDatabase(projectName);
      // Check if already assigned
      const existing = queryOne(db, 'SELECT id FROM entity_assignee WHERE entity_id = ? AND assignee_id = ?', [entityId, userId]);
      if (!existing) {
        execute(db, `
          INSERT INTO entity_assignee (id, mtime, entity_id, assignee_id, assigner_id, synced)
          VALUES (?, ?, ?, ?, ?, 0)
        `, [id, now, entityId, userId, '']);
        await persistDatabase(projectName);
      }
    } catch (error) {
      console.error('Assign error:', error);
      throw error;
    }
  },

  // Unassigns a user from a collection (local-first approach)
  // Uses entity_assignee junction table
  Unassign: async (projectPath, entityId, userId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM entity_assignee WHERE entity_id = ? AND assignee_id = ?', [entityId, userId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('Unassign error:', error);
      throw error;
    }
  },

  // Updates the preview image for a collection (local-first approach)
  UpdatePreview: async (projectPath, entityId, previewPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE entity SET preview_id = ?, mtime = ?, synced = 0 WHERE id = ?', 
        [previewPath || '', Date.now(), entityId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdatePreview error:', error);
      throw error;
    }
  },

  // Reveals a collection in the file explorer - not available in web mode
  RevealCollection: async (projectPath, entityId) => {
    console.warn('RevealCollection not available in web mode (no file system)');
    return {};
  },

  // Reverts collections - not available in web mode
  RevertCollections: async (projectPath, entityIds) => {
    console.warn('RevertCollections not available in web mode (no file system)');
    return {};
  },

  // Returns items for checkpoint - not available in web mode
  GetItemsForCheckpoint: async (projectPath, entityId, targetPath, projectWorkingDir, ignoreList) => {
    console.warn('GetItemsForCheckpoint not available in web mode (no file system)');
    return [];
  },

  // Returns outdated items in collection - not available in web mode
  GetOutdatedItemsInCollection: async (projectPath, entityId, projectWorkingDir, ignoreList) => {
    console.warn('GetOutdatedItemsInCollection not available in web mode (no file system)');
    return [];
  },

  // Returns collection count
  GetCollectionCount: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const row = queryOne(db, 'SELECT COUNT(*) as count FROM entity WHERE trashed = 0');
      return row?.count || 0;
    } catch (error) {
      console.error('GetCollectionCount error:', error);
      return 0;
    }
  },

  // Returns tasks for a collection
  GetCollectionTasks: async (projectPath, entityId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      // Build lookup maps for enrichment
      const statusRows = query(db, 'SELECT * FROM status');
      const statusMap = {};
      for (const s of statusRows) {
        statusMap[s.id] = s;
      }
      
      const taskTypeRows = query(db, 'SELECT * FROM task_type');
      const taskTypeMap = {};
      for (const tt of taskTypeRows) {
        taskTypeMap[tt.id] = tt;
      }
      
      const tagRows = query(db, 'SELECT * FROM tag');
      const tagMap = {};
      for (const t of tagRows) {
        tagMap[t.id] = t;
      }
      
      const taskTagRows = query(db, 'SELECT * FROM task_tag');
      const taskTagsMap = {};
      for (const tt of taskTagRows) {
        if (!taskTagsMap[tt.task_id]) {
          taskTagsMap[tt.task_id] = [];
        }
        taskTagsMap[tt.task_id].push(tt.tag_id);
      }
      
      // Build task dependency map
      const taskDependencyRows = query(db, 'SELECT * FROM task_dependency');
      const taskDependenciesMap = {};
      for (const td of taskDependencyRows) {
        if (!taskDependenciesMap[td.task_id]) {
          taskDependenciesMap[td.task_id] = [];
        }
        taskDependenciesMap[td.task_id].push({
          id: td.dependency_id,
          type_id: td.dependency_type_id
        });
      }
      
      // Build entity dependency map
      const entityDependencyRows = query(db, 'SELECT * FROM entity_dependency');
      const entityDependenciesMap = {};
      for (const ed of entityDependencyRows) {
        if (!entityDependenciesMap[ed.task_id]) {
          entityDependenciesMap[ed.task_id] = [];
        }
        entityDependenciesMap[ed.task_id].push({
          id: ed.entity_id,
          type_id: ed.dependency_type_id
        });
      }
      
      // Build entityMap for task_path computation
      const entityRows = query(db, 'SELECT id, name, entity_path FROM entity');
      const entityMapForTasks = {};
      for (const e of entityRows) {
        entityMapForTasks[e.id] = e;
      }
      
      // Get tasks for this entity
      const taskCondition = entityId === 'root' || !entityId
        ? "(entity_id IS NULL OR entity_id = '')"
        : `entity_id = '${entityId}'`;
      
      const taskRows = query(db, `SELECT * FROM task WHERE ${taskCondition} AND trashed = 0`);
      return taskRows.map(row => rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMapForTasks));
    } catch (error) {
      console.error('GetCollectionTasks error:', error);
      return [];
    }
  },

  // Creates multiple collections at once (local-first approach)
  // Signature matches Wails: CreateCollections(projectPath, name, description, entityTypeId, parentId)
  CreateCollections: async (projectPath, name, description, entityTypeId, parentId) => {
    const projectName = getProjectName(projectPath);
    const id = crypto.randomUUID();
    const now = Date.now();
    const createdAt = new Date().toISOString();
    
    try {
      const db = await getDatabase(projectName);
      
      // Compute entity_path based on parent (mimics SQLite trigger behavior)
      const entityPath = computeEntityPath(db, parentId, name || '');
      
      execute(db, `
        INSERT INTO entity (id, created_at, mtime, name, description, entity_type_id, parent_id, entity_path, preview_id, is_library, trashed, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
      `, [id, createdAt, now, name || '', description || '', entityTypeId || '', parentId || '', entityPath, '', 0]);
      await persistDatabase(projectName);
      
      // Build entity type lookup map
      const entityTypeRows = query(db, 'SELECT * FROM entity_type');
      const entityTypeMap = {};
      for (const et of entityTypeRows) {
        entityTypeMap[et.id] = et;
      }
      
      // Return the created entity
      const row = queryOne(db, 'SELECT * FROM entity WHERE id = ?', [id]);
      return [rowToEntity(row, entityTypeMap)];
    } catch (error) {
      console.error('CreateCollections error:', error);
      throw error;
    }
  },

  // Checks if a user is assigned to a collection or any of its ancestor collections.
  IsUserAssignedToCollectionOrAncestor: async (projectPath, entityId, userId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);

      // Check the entity itself first.
      const direct = queryOne(db,
        `SELECT EXISTS(SELECT 1 FROM entity_assignee WHERE entity_id = ? AND assignee_id = ?) AS result`,
        [entityId, userId]
      );
      if (direct?.result === 1) return true;

      // Check ancestor entities recursively.
      const ancestor = queryOne(db, `
        WITH RECURSIVE ancestors AS (
          SELECT parent_id FROM entity WHERE id = ? AND parent_id != ''
          UNION ALL
          SELECT e.parent_id FROM entity e
          JOIN ancestors a ON e.id = a.parent_id
          WHERE a.parent_id != ''
        )
        SELECT EXISTS(
          SELECT 1 FROM entity_assignee ea
          JOIN ancestors a ON ea.entity_id = a.parent_id
          WHERE ea.assignee_id = ?
        ) AS result
      `, [entityId, userId]);
      return ancestor?.result === 1;
    } catch (error) {
      console.error('IsUserAssignedToCollectionOrAncestor error:', error);
      return false;
    }
  },
};
