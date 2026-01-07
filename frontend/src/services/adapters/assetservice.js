import { studioApiCall, getActiveStudioUrl } from './http-client.js';
import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
}

/**
 * Convert database row to task object with proper types
 * Basic version without enrichment
 */
function rowToTaskBasic(row) {
  if (!row) return null;
  return {
    ...row,
    is_resource: !!row.is_resource,
    is_link: !!row.is_link,
    trashed: !!row.trashed,
    is_trashed: !!row.trashed,
    synced: !!row.synced,
    type: 'task',
    // Web mode specific - no local file tracking
    file_status: 'normal',
  };
}

/**
 * Convert database row to task object with enriched data (status, tags, type)
 */
function rowToTask(row, statusMap = {}, taskTypeMap = {}, taskTagsMap = {}, tagMap = {}, taskDependenciesMap = {}, entityDependenciesMap = {}) {
  if (!row) return null;
  const status = statusMap[row.status_id] || {};
  const taskType = taskTypeMap[row.task_type_id] || {};
  const tagIds = taskTagsMap[row.id] || [];
  const tags = tagIds.map(tagId => tagMap[tagId]?.name || tagId).filter(Boolean);
  
  // Get dependencies for this task
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

/**
 * Build lookup maps from database for enriching tasks
 */
function buildLookupMaps(db) {
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
  
  // Map task_id -> [dependency objects]
  const taskDependencyRows = query(db, 'SELECT * FROM task_dependency');
  const taskDependenciesMap = {};
  for (const td of taskDependencyRows) {
    if (!taskDependenciesMap[td.task_id]) taskDependenciesMap[td.task_id] = [];
    taskDependenciesMap[td.task_id].push({
      id: td.dependency_id,
      type_id: td.dependency_type_id
    });
  }
  
  // Map task_id -> [entity dependency objects]
  const entityDependencyRows = query(db, 'SELECT * FROM entity_dependency');
  const entityDependenciesMap = {};
  for (const ed of entityDependencyRows) {
    if (!entityDependenciesMap[ed.task_id]) entityDependenciesMap[ed.task_id] = [];
    entityDependenciesMap[ed.task_id].push({
      id: ed.entity_id,
      type_id: ed.dependency_type_id
    });
  }
  
  return { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap };
}

export const AssetService = {
  // Returns all assets in a project
  GetAssets: async (projectPath, includeTrashed = false) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap } = buildLookupMaps(db);
      
      const trashedCondition = includeTrashed ? '' : 'WHERE trashed = 0';
      const rows = query(db, `SELECT * FROM task ${trashedCondition}`);
      return rows.map(row => rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap));
    } catch (error) {
      console.error('GetAssets error:', error);
      return [];
    }
  },

  // Returns a specific asset by ID
  GetAsset: async (projectPath, assetId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap } = buildLookupMaps(db);
      
      const row = queryOne(db, 'SELECT * FROM task WHERE id = ?', [assetId]);
      return rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap) || {};
    } catch (error) {
      console.error('GetAsset error:', error);
      return {};
    }
  },

  // Creates a new asset
  CreateAsset: async (projectPath, asset) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    // Create on server first
    const result = await studioApiCall(studioUrl, `/${projectName}/task`, 'POST', asset);
    
    // Insert into local SQLite
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO task (id, mtime, created_at, name, description, extension, is_resource, 
                         status_id, task_type_id, entity_id, assignee_id, assigner_id, 
                         is_link, pointer, preview_id, trashed, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1)
      `, [
        result.id, Date.now(), result.created_at || new Date().toISOString(), 
        result.name, result.description, result.extension, result.is_resource ? 1 : 0,
        result.status_id, result.task_type_id, result.entity_id, result.assignee_id,
        result.assigner_id, result.is_link ? 1 : 0, result.pointer, result.preview_id
      ]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('CreateAsset local insert error:', error);
    }
    
    return rowToTask(result);
  },

  // Updates an existing asset
  UpdateAsset: async (projectPath, asset) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    // Update on server
    const result = await studioApiCall(studioUrl, `/${projectName}/task/${asset.id}`, 'PUT', asset);
    
    // Update local SQLite
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        UPDATE task SET name = ?, description = ?, extension = ?, is_resource = ?,
                       status_id = ?, task_type_id = ?, entity_id = ?, assignee_id = ?,
                       assigner_id = ?, is_link = ?, pointer = ?, preview_id = ?, mtime = ?
        WHERE id = ?
      `, [
        asset.name, asset.description, asset.extension, asset.is_resource ? 1 : 0,
        asset.status_id, asset.task_type_id, asset.entity_id, asset.assignee_id,
        asset.assigner_id, asset.is_link ? 1 : 0, asset.pointer, asset.preview_id,
        Date.now(), asset.id
      ]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdateAsset local update error:', error);
    }
    
    return rowToTask(result);
  },

  // Deletes an asset
  DeleteAsset: async (projectPath, assetId, moveToTrash = false) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const endpoint = moveToTrash 
      ? `/${projectName}/task/${assetId}/trash`
      : `/${projectName}/task/${assetId}`;
    
    await studioApiCall(studioUrl, endpoint, 'DELETE');
    
    // Update local SQLite
    try {
      const db = await getDatabase(projectName);
      if (moveToTrash) {
        execute(db, 'UPDATE task SET trashed = 1 WHERE id = ?', [assetId]);
      } else {
        execute(db, 'DELETE FROM task WHERE id = ?', [assetId]);
      }
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteAsset local update error:', error);
    }
  },

  // Changes an asset's collection (entity)
  ChangeAssetCollection: async (projectPath, assetId, newEntityId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    await studioApiCall(studioUrl, `/${projectName}/task/${assetId}/entity`, 'PUT', { entity_id: newEntityId });
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE task SET entity_id = ?, mtime = ? WHERE id = ?', [newEntityId, Date.now(), assetId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeAssetCollection local update error:', error);
    }
  },

  // Adds a task dependency
  AddAssetDependency: async (projectPath, taskId, dependencyId, dependencyTypeId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/task/${taskId}/dependency`, 'POST', {
      dependency_id: dependencyId,
      dependency_type_id: dependencyTypeId
    });
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO task_dependency (id, mtime, task_id, dependency_id, dependency_type_id, synced)
        VALUES (?, ?, ?, ?, ?, 1)
      `, [result.id || crypto.randomUUID(), Date.now(), taskId, dependencyId, dependencyTypeId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('AddAssetDependency local insert error:', error);
    }
    
    return result;
  },

  // Adds an entity dependency
  AddEntityDependency: async (projectPath, taskId, entityId, dependencyTypeId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/task/${taskId}/entity-dependency`, 'POST', {
      entity_id: entityId,
      dependency_type_id: dependencyTypeId
    });
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO entity_dependency (id, mtime, task_id, entity_id, dependency_type_id, synced)
        VALUES (?, ?, ?, ?, ?, 1)
      `, [result.id || crypto.randomUUID(), Date.now(), taskId, entityId, dependencyTypeId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('AddEntityDependency local insert error:', error);
    }
    
    return result;
  },

  // Returns all asset types for a project
  GetAssetTypes: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      return query(db, 'SELECT * FROM task_type');
    } catch (error) {
      console.error('GetAssetTypes error:', error);
      return [];
    }
  },

  // Creates a new asset type
  CreateAssetType: async (projectPath, type) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/task-type`, 'POST', type);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'INSERT INTO task_type (id, mtime, name, icon, synced) VALUES (?, ?, ?, ?, 1)', 
        [result.id, Date.now(), result.name, result.icon]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('CreateAssetType local insert error:', error);
    }
    
    return result;
  },

  // Updates an existing asset type
  UpdateAssetType: async (projectPath, type) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/task-type/${type.id}`, 'PUT', type);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE task_type SET name = ?, icon = ?, mtime = ? WHERE id = ?',
        [type.name, type.icon, Date.now(), type.id]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdateAssetType local update error:', error);
    }
    
    return result;
  },

  // Deletes an asset type
  DeleteAssetType: async (projectPath, typeId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    await studioApiCall(studioUrl, `/${projectName}/task-type/${typeId}`, 'DELETE');
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM task_type WHERE id = ?', [typeId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteAssetType local update error:', error);
    }
  },

  // Returns the state of an asset (file status)
  // In web mode, always returns 'normal' since we don't track local file states
  GetAssetState: async (projectPath, assetId) => {
    return 'normal';
  },

  // Returns full task/entity objects for a list of dependency IDs
  GetAssetDependencies: async (projectPath, dependencyIds) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      if (!dependencyIds || dependencyIds.length === 0) {
        return [];
      }
      
      const result = [];
      
      // Build lookup maps for task enrichment
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap } = buildLookupMaps(db);
      
      // Build entity type map
      const entityTypeRows = query(db, 'SELECT * FROM entity_type');
      const entityTypeMap = {};
      for (const et of entityTypeRows) {
        entityTypeMap[et.id] = et;
      }
      
      // Try to find tasks first
      const placeholders = dependencyIds.map(() => '?').join(',');
      const taskQuery = `SELECT * FROM task WHERE id IN (${placeholders}) AND trashed = 0 ORDER BY name`;
      const taskRows = query(db, taskQuery, dependencyIds);
      
      const foundTaskIds = new Set();
      for (const row of taskRows) {
        foundTaskIds.add(row.id);
        const task = rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap);
        result.push(task);
      }
      
      // Find IDs that didn't match any tasks - they might be entities
      const missingIds = dependencyIds.filter(id => !foundTaskIds.has(id));
      
      if (missingIds.length > 0) {
        const entityPlaceholders = missingIds.map(() => '?').join(',');
        const entityQuery = `SELECT * FROM entity WHERE id IN (${entityPlaceholders}) AND trashed = 0 ORDER BY name`;
        const entityRows = query(db, entityQuery, missingIds);
        
        for (const row of entityRows) {
          const entity = {
            ...row,
            trashed: !!row.trashed,
            is_trashed: !!row.trashed,
            synced: !!row.synced,
            type: 'entity',
            entity_type: entityTypeMap[row.entity_type_id]?.name || '',
            entity_type_icon: entityTypeMap[row.entity_type_id]?.icon || '',
          };
          result.push(entity);
        }
      }
      
      return result;
    } catch (error) {
      console.error('GetAssetDependencies error:', error);
      return [];
    }
  },
};
