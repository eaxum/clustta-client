import { studioApiCall, getActiveStudioUrl } from './http-client.js';
import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
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
    synced: !!row.synced,
    type: 'entity',
    entity_type: entityType.name || '',
    entity_type_icon: entityType.icon || '',
  };
}

/**
 * Convert database row to task object with proper types
 * Optionally enriches with status, tags, entity type info, and dependencies
 */
function rowToTask(row, statusMap = {}, taskTypeMap = {}, taskTagsMap = {}, tagMap = {}, taskDependenciesMap = {}, entityDependenciesMap = {}) {
  if (!row) return null;
  const status = statusMap[row.status_id] || {};
  const taskType = taskTypeMap[row.task_type_id] || {};
  const tagIds = taskTagsMap[row.id] || [];
  const tags = tagIds.map(tagId => tagMap[tagId]?.name || tagId).filter(Boolean);
  const dependencies = taskDependenciesMap[row.id] || [];
  const entity_dependencies = entityDependenciesMap[row.id] || [];
  
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
    status_color: status.color || '',
    // Task type info
    task_type: taskType.name || '',
    task_type_icon: taskType.icon || '',
    // Tags array
    tags: tags,
    // Dependencies arrays
    dependencies: dependencies,
    entity_dependencies: entity_dependencies,
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
      
      // Get child tasks
      const taskCondition = entityId === 'root'
        ? "(entity_id IS NULL OR entity_id = '')"
        : `entity_id = '${entityId}'`;
      
      const taskRows = query(db, `SELECT * FROM task WHERE ${taskCondition} AND trashed = 0`);
      const tasks = taskRows.map(row => rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap));
      
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

  // Changes the parent of a collection
  ChangeCollectionParent: async (projectPath, entityId, newParentId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    // Update server
    await studioApiCall(studioUrl, `/${projectName}/entity/${entityId}/parent`, 'PUT', { parent_id: newParentId });
    
    // Update local SQLite
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE entity SET parent_id = ? WHERE id = ?', [newParentId, entityId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeCollectionParent local update error:', error);
    }
  },

  // Creates a new collection
  CreateCollection: async (projectPath, collection) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    // Create on server first
    const result = await studioApiCall(studioUrl, `/${projectName}/entity`, 'POST', collection);
    
    // Insert into local SQLite
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO entity (id, mtime, name, parent_id, entity_type_id, entity_path, trashed, synced)
        VALUES (?, ?, ?, ?, ?, ?, 0, 1)
      `, [result.id, Date.now(), result.name, result.parent_id, result.entity_type_id, result.entity_path]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('CreateCollection local insert error:', error);
    }
    
    return result;
  },

  // Updates an existing collection
  UpdateCollection: async (projectPath, collection) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    // Update on server
    const result = await studioApiCall(studioUrl, `/${projectName}/entity/${collection.id}`, 'PUT', collection);
    
    // Update local SQLite
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        UPDATE entity SET name = ?, parent_id = ?, entity_type_id = ?, entity_path = ?, mtime = ?
        WHERE id = ?
      `, [collection.name, collection.parent_id, collection.entity_type_id, collection.entity_path, Date.now(), collection.id]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdateCollection local update error:', error);
    }
    
    return result;
  },

  // Deletes a collection
  DeleteCollection: async (projectPath, collectionId, moveToTrash = false) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const endpoint = moveToTrash 
      ? `/${projectName}/entity/${collectionId}/trash`
      : `/${projectName}/entity/${collectionId}`;
    
    const result = await studioApiCall(studioUrl, endpoint, 'DELETE');
    
    // Update local SQLite
    try {
      const db = await getDatabase(projectName);
      if (moveToTrash) {
        execute(db, 'UPDATE entity SET trashed = 1 WHERE id = ?', [collectionId]);
      } else {
        execute(db, 'DELETE FROM entity WHERE id = ?', [collectionId]);
      }
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteCollection local update error:', error);
    }
    
    return result;
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

  // Creates a new collection type
  CreateCollectionType: async (projectPath, type) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/entity-type`, 'POST', type);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'INSERT INTO entity_type (id, mtime, name, icon, synced) VALUES (?, ?, ?, ?, 1)', 
        [result.id, Date.now(), result.name, result.icon]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('CreateCollectionType local insert error:', error);
    }
    
    return result;
  },

  // Updates an existing collection type
  UpdateCollectionType: async (projectPath, type) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/entity-type/${type.id}`, 'PUT', type);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE entity_type SET name = ?, icon = ?, mtime = ? WHERE id = ?',
        [type.name, type.icon, Date.now(), type.id]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdateCollectionType local update error:', error);
    }
    
    return result;
  },

  // Deletes a collection type
  DeleteCollectionType: async (projectPath, typeId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    await studioApiCall(studioUrl, `/${projectName}/entity-type/${typeId}`, 'DELETE');
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM entity_type WHERE id = ?', [typeId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteCollectionType local update error:', error);
    }
  },
};
