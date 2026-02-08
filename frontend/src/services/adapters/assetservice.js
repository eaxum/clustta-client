import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';
import { STORAGE_KEYS } from './config.js';

/**
 * Get current user ID from localStorage
 */
function getCurrentUserId() {
  try {
    const user = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || '{}');
    return user.id || '';
  } catch {
    return '';
  }
}

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
 * Also computes entity_path and task_path from entity
 */
function rowToTask(row, statusMap = {}, taskTypeMap = {}, taskTagsMap = {}, tagMap = {}, taskDependenciesMap = {}, entityDependenciesMap = {}, entityMap = {}) {
  if (!row) return null;
  const status = statusMap[row.status_id] || {};
  const taskType = taskTypeMap[row.task_type_id] || {};
  const tagIds = taskTagsMap[row.id] || [];
  const tags = tagIds.map(tagId => tagMap[tagId]?.name || tagId).filter(Boolean);
  
  // Get dependencies for this task
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
  
  // Map entity_id -> entity (for computing task_path/entity_path)
  const entityRows = query(db, 'SELECT id, name, entity_path FROM entity');
  const entityMap = {};
  for (const e of entityRows) {
    entityMap[e.id] = e;
  }
  
  return { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap, entityMap };
}

export const AssetService = {
  // Returns all assets in a project
  GetAssets: async (projectPath, includeTrashed = false) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap, entityMap } = buildLookupMaps(db);
      
      const trashedCondition = includeTrashed ? '' : 'WHERE trashed = 0';
      const rows = query(db, `SELECT * FROM task ${trashedCondition}`);
      return rows.map(row => rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMap));
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
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap, entityMap } = buildLookupMaps(db);
      
      const row = queryOne(db, 'SELECT * FROM task WHERE id = ?', [assetId]);
      return rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMap) || {};
    } catch (error) {
      console.error('GetAsset error:', error);
      return {};
    }
  },

  // Creates a new asset (local-first approach)
  // Signature matches Wails: CreateAsset(projectPath, name, description, taskTypeId, entityId, isResource, templateId, templateFilePath, pointer, isLink, tags, previewPath, comment)
  CreateAsset: async (projectPath, name, description, taskTypeId, entityId, isResource, templateId, templateFilePath, pointer, isLink, tags, previewPath, comment) => {
    const projectName = getProjectName(projectPath);
    
    // Generate IDs client-side for local-first
    const taskId = crypto.randomUUID();
    const checkpointId = crypto.randomUUID();
    const checkpointGroupId = crypto.randomUUID();
    const now = Date.now();
    const createdAt = new Date().toISOString();
    
    // Template data for checkpoint creation
    let extension = '';
    let templateChunks = '';
    let templateChecksum = '';
    let templateFileSize = 0;
    
    try {
      const db = await getDatabase(projectName);
      
      // Get template data if templateId is provided (needed for checkpoint)
      if (templateId && !isLink) {
        const template = queryOne(db, 'SELECT * FROM template WHERE id = ?', [templateId]);
        if (template) {
          extension = template.extension || '';
          templateChunks = template.chunks || '';
          templateChecksum = template.xxhash_checksum || '';
          templateFileSize = template.file_size || 0;
        }
      }
      
      // Get default status (first status in list, or look for "todo"/"pending")
      let statusId = '';
      const statuses = query(db, 'SELECT id, name, short_name FROM status ORDER BY name');
      if (statuses.length > 0) {
        // Try to find a default status like "todo" or "pending", otherwise use first
        const defaultStatus = statuses.find(s => 
          s.short_name?.toLowerCase() === 'todo' || 
          s.name?.toLowerCase() === 'todo' || 
          s.name?.toLowerCase() === 'pending'
        ) || statuses[0];
        statusId = defaultStatus.id;
      }
      
      // Insert task into local SQLite with synced=0 (dirty)
      execute(db, `
        INSERT INTO task (id, mtime, created_at, name, description, extension, is_resource, 
                         status_id, task_type_id, entity_id, assignee_id, assigner_id, 
                         is_link, pointer, preview_id, trashed, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
      `, [
        taskId, now, createdAt, 
        name || '', description || '', extension, isResource ? 1 : 0,
        statusId, taskTypeId || '', entityId || '', '',
        '', isLink ? 1 : 0, pointer || '', previewPath || ''
      ]);
      
      // Handle tags if provided
      if (tags && Array.isArray(tags) && tags.length > 0) {
        for (const tagName of tags) {
          // Find or create tag
          let tag = queryOne(db, 'SELECT id FROM tag WHERE name = ?', [tagName]);
          if (!tag) {
            const tagId = crypto.randomUUID();
            execute(db, 'INSERT INTO tag (id, mtime, name, synced) VALUES (?, ?, ?, 0)', [tagId, now, tagName]);
            tag = { id: tagId };
          }
          // Create task_tag association
          const taskTagId = crypto.randomUUID();
          execute(db, 'INSERT INTO task_tag (id, mtime, task_id, tag_id, synced) VALUES (?, ?, ?, ?, 0)', 
            [taskTagId, now, taskId, tag.id]);
        }
      }
      
      // Create initial checkpoint if not a link and template was provided
      // This mirrors the Go implementation which creates a checkpoint with the template's chunks
      if (!isLink && templateId && templateChunks) {
        const checkpointComment = comment || 'Asset created';
        const authorId = getCurrentUserId();
        execute(db, `
          INSERT INTO task_checkpoint (id, created_at, mtime, task_id, xxhash_checksum, 
                                       time_modified, file_size, chunks, comment, 
                                       author_id, group_id, preview_id, trashed, synced)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
        `, [
          checkpointId, createdAt, now, taskId, templateChecksum,
          Math.floor(now / 1000), templateFileSize, templateChunks, checkpointComment,
          authorId, checkpointGroupId, '', // author_id from current user, group_id for batch, no preview
        ]);
      }
      
      await persistDatabase(projectName);
      
      // Return the created task object
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap, entityMap } = buildLookupMaps(db);
      const row = queryOne(db, 'SELECT * FROM task WHERE id = ?', [taskId]);
      return rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMap);
    } catch (error) {
      console.error('CreateAsset error:', error);
      throw error;
    }
  },

  // Updates an existing asset (local-first approach)
  // Signature matches Wails binding: UpdateAsset(projectPath, taskId, name, taskTypeId, isResource, pointer, tags)
  UpdateAsset: async (projectPath, taskId, name, taskTypeId, isResource, pointer, tags) => {
    const projectName = getProjectName(projectPath);
    
    // Validate taskId is present
    if (!taskId) {
      console.error('UpdateAsset error: taskId is required');
      throw new Error('taskId is required for update');
    }
    
    try {
      const db = await getDatabase(projectName);
      
      execute(db, `
        UPDATE task SET name = ?, task_type_id = ?, is_resource = ?, pointer = ?, mtime = ?, synced = 0
        WHERE id = ?
      `, [
        name ?? '', 
        taskTypeId ?? null, 
        isResource ? 1 : 0, 
        pointer ?? '',
        Date.now(), 
        taskId
      ]);
      await persistDatabase(projectName);
      
      // TODO: Handle tags update separately if needed
      
      // Return the updated task
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap, entityMap } = buildLookupMaps(db);
      const row = queryOne(db, 'SELECT * FROM task WHERE id = ?', [taskId]);
      return rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMap);
    } catch (error) {
      console.error('UpdateAsset error:', error);
      throw error;
    }
  },

  // Deletes an asset (local-first approach)
  DeleteAsset: async (projectPath, assetId, moveToTrash = false) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      if (moveToTrash) {
        execute(db, 'UPDATE task SET trashed = 1, synced = 0, mtime = ? WHERE id = ?', [Date.now(), assetId]);
      } else {
        execute(db, 'DELETE FROM task WHERE id = ?', [assetId]);
      }
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteAsset error:', error);
      throw error;
    }
  },

  // Changes an asset's collection (entity) - local-first approach
  ChangeAssetCollection: async (projectPath, assetId, newEntityId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE task SET entity_id = ?, mtime = ?, synced = 0 WHERE id = ?', [newEntityId, Date.now(), assetId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeAssetCollection error:', error);
      throw error;
    }
  },

  // Adds a task dependency (local-first approach)
  AddAssetDependency: async (projectPath, taskId, dependencyId, dependencyTypeId) => {
    const projectName = getProjectName(projectPath);
    const id = crypto.randomUUID();
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO task_dependency (id, mtime, task_id, dependency_id, dependency_type_id, synced)
        VALUES (?, ?, ?, ?, ?, 0)
      `, [id, Date.now(), taskId, dependencyId, dependencyTypeId]);
      await persistDatabase(projectName);
      
      return { id, task_id: taskId, dependency_id: dependencyId, dependency_type_id: dependencyTypeId };
    } catch (error) {
      console.error('AddAssetDependency error:', error);
      throw error;
    }
  },

  // Adds an entity dependency (local-first approach)
  AddEntityDependency: async (projectPath, taskId, entityId, dependencyTypeId) => {
    const projectName = getProjectName(projectPath);
    const id = crypto.randomUUID();
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO entity_dependency (id, mtime, task_id, entity_id, dependency_type_id, synced)
        VALUES (?, ?, ?, ?, ?, 0)
      `, [id, Date.now(), taskId, entityId, dependencyTypeId]);
      await persistDatabase(projectName);
      
      return { id, task_id: taskId, entity_id: entityId, dependency_type_id: dependencyTypeId };
    } catch (error) {
      console.error('AddEntityDependency error:', error);
      throw error;
    }
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

  // Creates a new asset type (local-first approach)
  CreateAssetType: async (projectPath, type) => {
    const projectName = getProjectName(projectPath);
    const id = type.id || crypto.randomUUID();
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'INSERT INTO task_type (id, mtime, name, icon, synced) VALUES (?, ?, ?, ?, 0)', 
        [id, Date.now(), type.name || '', type.icon || '']);
      await persistDatabase(projectName);
      
      return { id, name: type.name, icon: type.icon };
    } catch (error) {
      console.error('CreateAssetType error:', error);
      throw error;
    }
  },

  // Updates an existing asset type (local-first approach)
  UpdateAssetType: async (projectPath, type) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE task_type SET name = ?, icon = ?, mtime = ?, synced = 0 WHERE id = ?',
        [type.name || '', type.icon || '', Date.now(), type.id]);
      await persistDatabase(projectName);
      
      return { id: type.id, name: type.name, icon: type.icon };
    } catch (error) {
      console.error('UpdateAssetType error:', error);
      throw error;
    }
  },

  // Deletes an asset type (local-first approach)
  DeleteAssetType: async (projectPath, typeId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM task_type WHERE id = ?', [typeId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteAssetType error:', error);
      throw error;
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
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap, entityMap } = buildLookupMaps(db);
      
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
        const task = rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMap);
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

  // Renames an asset (local-first approach)
  RenameAsset: async (projectPath, taskId, newName) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE task SET name = ?, mtime = ?, synced = 0 WHERE id = ?', [newName, Date.now(), taskId]);
      await persistDatabase(projectName);
      
      // Return updated task
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap, entityMap } = buildLookupMaps(db);
      const row = queryOne(db, 'SELECT * FROM task WHERE id = ?', [taskId]);
      return rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMap);
    } catch (error) {
      console.error('RenameAsset error:', error);
      throw error;
    }
  },

  // Duplicates an asset (local-first approach)
  DuplicateAsset: async (projectPath, sourceTaskId) => {
    const projectName = getProjectName(projectPath);
    const newTaskId = crypto.randomUUID();
    const now = Date.now();
    const createdAt = new Date().toISOString();
    
    try {
      const db = await getDatabase(projectName);
      
      // Get source task
      const source = queryOne(db, 'SELECT * FROM task WHERE id = ?', [sourceTaskId]);
      if (!source) throw new Error('Source task not found');
      
      // Generate unique duplicate name (like Go: baseName + "-duplicate", increment if exists)
      const baseName = source.name;
      let duplicateName = `${baseName}-duplicate`;
      let counter = 1;
      while (queryOne(db, 'SELECT id FROM task WHERE name = ? AND entity_id = ?', [duplicateName, source.entity_id])) {
        counter++;
        duplicateName = `${baseName}-duplicate-${counter}`;
      }
      
      // Insert duplicated task with new ID
      execute(db, `
        INSERT INTO task (id, mtime, created_at, name, description, extension, is_resource, 
                         status_id, task_type_id, entity_id, assignee_id, assigner_id, 
                         is_link, pointer, preview_id, trashed, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
      `, [
        newTaskId, now, createdAt, 
        duplicateName, source.description || '', source.extension || '', source.is_resource,
        source.status_id || '', source.task_type_id || '', source.entity_id || '', '',
        '', source.is_link, source.pointer || '', ''
      ]);
      
      // Copy tags from source task
      const sourceTags = query(db, 'SELECT tag_id FROM task_tag WHERE task_id = ?', [sourceTaskId]);
      for (const tag of sourceTags) {
        const newTagId = crypto.randomUUID();
        execute(db, 'INSERT INTO task_tag (id, mtime, task_id, tag_id, synced) VALUES (?, ?, ?, ?, 0)', 
          [newTagId, now, newTaskId, tag.tag_id]);
      }
      
      // Copy latest checkpoint from source task
      const latestCheckpoint = queryOne(db, 
        'SELECT * FROM task_checkpoint WHERE task_id = ? ORDER BY created_at DESC LIMIT 1', 
        [sourceTaskId]
      );
      
      if (latestCheckpoint) {
        const newCheckpointId = crypto.randomUUID();
        const newGroupId = crypto.randomUUID();
        const authorId = getCurrentUserId();
        
        execute(db, `
          INSERT INTO task_checkpoint (id, created_at, mtime, task_id, xxhash_checksum, 
                                       time_modified, file_size, chunks, comment, 
                                       author_id, group_id, preview_id, trashed, synced)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
        `, [
          newCheckpointId, createdAt, now, newTaskId,
          latestCheckpoint.xxhash_checksum || '',
          latestCheckpoint.time_modified || now,
          latestCheckpoint.file_size || 0,
          latestCheckpoint.chunks || '[]',
          latestCheckpoint.comment || '',
          authorId,
          newGroupId,
          latestCheckpoint.preview_id || ''
        ]);
      }
      
      await persistDatabase(projectName);
      
      // Return the new task
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap, entityMap } = buildLookupMaps(db);
      const row = queryOne(db, 'SELECT * FROM task WHERE id = ?', [newTaskId]);
      return rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMap);
    } catch (error) {
      console.error('DuplicateAsset error:', error);
      throw error;
    }
  },

  // Changes the status of an asset (local-first approach)
  ChangeStatus: async (projectPath, taskId, statusId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE task SET status_id = ?, mtime = ?, synced = 0 WHERE id = ?', [statusId, Date.now(), taskId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeStatus error:', error);
      throw error;
    }
  },

  // Changes the type of an asset (local-first approach)
  ChangeAssetType: async (projectPath, taskId, taskTypeId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE task SET task_type_id = ?, mtime = ?, synced = 0 WHERE id = ?', [taskTypeId, Date.now(), taskId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeAssetType error:', error);
      throw error;
    }
  },

  // Toggles whether an asset is a task or resource (local-first approach)
  ToggleIsTask: async (projectPath, taskId, isTask) => {
    const projectName = getProjectName(projectPath);
    
    // isTask means it's NOT a resource
    const isResource = !isTask;
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE task SET is_resource = ?, mtime = ?, synced = 0 WHERE id = ?', [isResource ? 1 : 0, Date.now(), taskId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ToggleIsTask error:', error);
      throw error;
    }
  },

  // Toggles whether assets are resources (local-first approach)
  ToggleIsResource: async (projectPath, taskIds, isResource) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      const placeholders = taskIds.map(() => '?').join(',');
      execute(db, `UPDATE task SET is_resource = ?, mtime = ?, synced = 0 WHERE id IN (${placeholders})`, [isResource ? 1 : 0, Date.now(), ...taskIds]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ToggleIsResource error:', error);
      throw error;
    }
  },

  // Assigns a user to an asset (local-first approach)
  AssignAsset: async (projectPath, taskId, userId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE task SET assignee_id = ?, mtime = ?, synced = 0 WHERE id = ?', [userId, Date.now(), taskId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('AssignAsset error:', error);
      throw error;
    }
  },

  // Unassigns a user from an asset (local-first approach)
  UnassignAsset: async (projectPath, taskId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE task SET assignee_id = NULL, mtime = ?, synced = 0 WHERE id = ?', [Date.now(), taskId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UnassignAsset error:', error);
      throw error;
    }
  },

  // Unassigns multiple assets (local-first approach)
  UnassignAssets: async (projectPath, taskIds) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      const placeholders = taskIds.map(() => '?').join(',');
      execute(db, `UPDATE task SET assignee_id = NULL, mtime = ?, synced = 0 WHERE id IN (${placeholders})`, [Date.now(), ...taskIds]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UnassignAssets error:', error);
      throw error;
    }
  },

  // Removes a task dependency (local-first approach)
  RemoveAssetDependency: async (projectPath, taskId, dependencyId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM task_dependency WHERE task_id = ? AND dependency_id = ?', [taskId, dependencyId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('RemoveAssetDependency error:', error);
      throw error;
    }
  },

  // Removes an entity dependency (local-first approach)
  RemoveEntityDependency: async (projectPath, taskId, entityId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM entity_dependency WHERE task_id = ? AND entity_id = ?', [taskId, entityId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('RemoveEntityDependency error:', error);
      throw error;
    }
  },

  // Adds a preview to an asset (local-first approach)
  // In web mode, preview path would be a URL/blob reference
  AddPreview: async (projectPath, taskId, previewPath) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE task SET preview_id = ?, mtime = ?, synced = 0 WHERE id = ?', [previewPath, Date.now(), taskId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('AddPreview error:', error);
      throw error;
    }
  },

  // Returns asset by ID (alias)
  GetAssetByID: async (projectPath, assetId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap, entityMap } = buildLookupMaps(db);
      
      const row = queryOne(db, 'SELECT * FROM task WHERE id = ?', [assetId]);
      return rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMap) || {};
    } catch (error) {
      console.error('GetAssetByID error:', error);
      return {};
    }
  },

  // Returns asset by its task_path.
  GetAssetByPath: async (projectPath, taskPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap, entityMap } = buildLookupMaps(db);

      const rows = query(db, 'SELECT * FROM task WHERE trashed = 0');
      for (const row of rows) {
        const task = rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMap);
        if (task && task.task_path === taskPath) {
          return task;
        }
      }
      return {};
    } catch (error) {
      console.error('GetAssetByPath error:', error);
      return {};
    }
  },

  // Returns asset count
  GetAssetCount: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const row = queryOne(db, 'SELECT COUNT(*) as count FROM task WHERE trashed = 0');
      return row?.count || 0;
    } catch (error) {
      console.error('GetAssetCount error:', error);
      return 0;
    }
  },

  // Returns all task-type assets
  GetAssetTasks: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap, entityMap } = buildLookupMaps(db);
      
      const rows = query(db, 'SELECT * FROM task WHERE is_resource = 0 AND trashed = 0');
      return rows.map(row => rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMap));
    } catch (error) {
      console.error('GetAssetTasks error:', error);
      return [];
    }
  },

  // Returns file status for an asset - not available in web mode
  AssetFileStatus: async (projectPath, taskId) => {
    return 'normal';
  },

  // Returns file statuses for multiple assets - not available in web mode
  AssetFilesStatus: async (projectPath, taskIds) => {
    const result = {};
    for (const id of taskIds) {
      result[id] = 'normal';
    }
    return result;
  },

  // Returns states for all assets - not available in web mode
  GetAssetsStates: async (projectPath, projectWorkingDir, ignoreList) => {
    return {};
  },

  // Returns untracked files - not available in web mode
  GetUntrackedFiles: async (projectPath, projectWorkingDir, ignoreList) => {
    return [];
  },

  // Reveals asset in file explorer - not available in web mode
  RevealAsset: async (projectPath, taskId) => {
    console.warn('RevealAsset not available in web mode (no file system)');
    return {};
  },

  // Returns recursive dependencies for a task
  GetRecursiveDependencies: async (projectPath, taskId, maxDepth = 5) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const { statusMap, taskTypeMap, tagMap, taskTagsMap, taskDependenciesMap, entityDependenciesMap, entityMap } = buildLookupMaps(db);
      
      const visited = new Set();
      const result = [];
      
      const collectDependencies = (id, depth) => {
        if (depth > maxDepth || visited.has(id)) return;
        visited.add(id);
        
        const deps = taskDependenciesMap[id] || [];
        for (const dep of deps) {
          const row = queryOne(db, 'SELECT * FROM task WHERE id = ? AND trashed = 0', [dep.id]);
          if (row) {
            result.push(rowToTask(row, statusMap, taskTypeMap, taskTagsMap, tagMap, taskDependenciesMap, entityDependenciesMap, entityMap));
            collectDependencies(dep.id, depth + 1);
          }
        }
      };
      
      collectDependencies(taskId, 0);
      return result;
    } catch (error) {
      console.error('GetRecursiveDependencies error:', error);
      return [];
    }
  },
};
