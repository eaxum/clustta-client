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
 * Convert database row to asset object with proper types
 * Basic version without enrichment
 */
function rowToAssetBasic(row) {
  if (!row) return null;
  return {
    ...row,
    is_resource: !!row.is_resource,
    is_link: !!row.is_link,
    trashed: !!row.trashed,
    is_trashed: !!row.trashed,
    synced: !!row.synced,
    type: 'asset',
    // Web mode specific - no local file tracking
    file_status: 'normal',
  };
}

/**
 * Convert database row to asset object with enriched data (status, tags, type)
 * Also computes collection_path and asset_path from collection
 */
function rowToAsset(row, statusMap = {}, assetTypeMap = {}, assetTagsMap = {}, tagMap = {}, assetDependenciesMap = {}, collectionDependenciesMap = {}, collectionMap = {}) {
  if (!row) return null;
  const status = statusMap[row.status_id] || {};
  const assetType = assetTypeMap[row.asset_type_id] || {};
  const tagIds = assetTagsMap[row.id] || [];
  const tags = tagIds.map(tagId => tagMap[tagId]?.name || tagId).filter(Boolean);
  
  // Get dependencies for this asset
  const dependencies = assetDependenciesMap[row.id] || [];
  const collection_dependencies = collectionDependenciesMap[row.id] || [];
  
  // Compute collection_path and asset_path from parent collection (mirrors full_asset view)
  const collection = collectionMap[row.collection_id] || {};
  const collectionPath = collection.collection_path || '';
  const assetPath = collectionPath ? collectionPath + row.name : '/' + row.name;
  const collectionName = collection.name || '';
  
  return {
    ...row,
    is_resource: !!row.is_resource,
    is_link: !!row.is_link,
    trashed: !!row.trashed,
    is_trashed: !!row.trashed,
    synced: !!row.synced,
    type: 'asset',
    // Status object
    status: status,
    status_name: status.name || '',
    status_short_name: status.short_name || '',
    status_color: status.color || '',
    // Asset type info
    asset_type: assetType.name || '',
    asset_type_name: assetType.name || '',
    asset_type_icon: assetType.icon || '',
    // Tags array
    tags: tags,
    // Dependencies arrays
    dependencies: dependencies,
    collection_dependencies: collection_dependencies,
    // Collection-related paths
    collection_path: collectionPath,
    asset_path: assetPath,
    collection_name: collectionName,
    // Web mode specific - no local file tracking
    file_status: 'normal',
  };
}

/**
 * Build lookup maps from database for enriching assets
 */
function buildLookupMaps(db) {
  const statusRows = query(db, 'SELECT * FROM status');
  const statusMap = {};
  for (const s of statusRows) {
    statusMap[s.id] = s;
  }
  
  const assetTypeRows = query(db, 'SELECT * FROM asset_type');
  const assetTypeMap = {};
  for (const tt of assetTypeRows) {
    assetTypeMap[tt.id] = tt;
  }
  
  const tagRows = query(db, 'SELECT * FROM tag');
  const tagMap = {};
  for (const t of tagRows) {
    tagMap[t.id] = t;
  }
  
  const assetTagRows = query(db, 'SELECT * FROM asset_tag');
  const assetTagsMap = {};
  for (const tt of assetTagRows) {
    if (!assetTagsMap[tt.asset_id]) {
      assetTagsMap[tt.asset_id] = [];
    }
    assetTagsMap[tt.asset_id].push(tt.tag_id);
  }
  
  // Map asset_id -> [dependency objects]
  const assetDependencyRows = query(db, 'SELECT * FROM asset_dependency');
  const assetDependenciesMap = {};
  for (const td of assetDependencyRows) {
    if (!assetDependenciesMap[td.asset_id]) assetDependenciesMap[td.asset_id] = [];
    assetDependenciesMap[td.asset_id].push({
      id: td.dependency_id,
      type_id: td.dependency_type_id
    });
  }
  
  // Map asset_id -> [collection dependency objects]
  const collectionDependencyRows = query(db, 'SELECT * FROM collection_dependency');
  const collectionDependenciesMap = {};
  for (const ed of collectionDependencyRows) {
    if (!collectionDependenciesMap[ed.asset_id]) collectionDependenciesMap[ed.asset_id] = [];
    collectionDependenciesMap[ed.asset_id].push({
      id: ed.collection_id,
      type_id: ed.dependency_type_id
    });
  }
  
  // Map collection_id -> collection (for computing asset_path/collection_path)
  const collectionRows = query(db, 'SELECT id, name, collection_path FROM collection');
  const collectionMap = {};
  for (const e of collectionRows) {
    collectionMap[e.id] = e;
  }
  
  return { statusMap, assetTypeMap, tagMap, assetTagsMap, assetDependenciesMap, collectionDependenciesMap, collectionMap };
}

export const AssetService = {
  // Returns all assets in a project
  GetAssets: async (projectPath, includeTrashed = false) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const { statusMap, assetTypeMap, tagMap, assetTagsMap, assetDependenciesMap, collectionDependenciesMap, collectionMap } = buildLookupMaps(db);
      
      const trashedCondition = includeTrashed ? '' : 'WHERE trashed = 0';
      const rows = query(db, `SELECT * FROM asset ${trashedCondition}`);
      return rows.map(row => rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMap));
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
      const { statusMap, assetTypeMap, tagMap, assetTagsMap, assetDependenciesMap, collectionDependenciesMap, collectionMap } = buildLookupMaps(db);
      
      const row = queryOne(db, 'SELECT * FROM asset WHERE id = ?', [assetId]);
      return rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMap) || {};
    } catch (error) {
      console.error('GetAsset error:', error);
      return {};
    }
  },

  // Creates a new asset (local-first approach)
  // Signature matches Wails: CreateAsset(projectPath, name, description, assetTypeId, collectionId, isResource, templateId, templateFilePath, pointer, isLink, tags, previewPath, comment)
  CreateAsset: async (projectPath, name, description, assetTypeId, collectionId, isResource, templateId, templateFilePath, pointer, isLink, tags, previewPath, comment) => {
    const projectName = getProjectName(projectPath);
    
    // Generate IDs client-side for local-first
    const assetId = crypto.randomUUID();
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
      
      // Insert asset into local SQLite with synced=0 (dirty)
      execute(db, `
        INSERT INTO asset (id, mtime, created_at, name, description, extension, is_resource, 
                         status_id, asset_type_id, collection_id, assignee_id, assigner_id, 
                         is_link, pointer, preview_id, trashed, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
      `, [
        assetId, now, createdAt, 
        name || '', description || '', extension, isResource ? 1 : 0,
        statusId, assetTypeId || '', collectionId || '', '',
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
          // Create asset_tag association
          const assetTagId = crypto.randomUUID();
          execute(db, 'INSERT INTO asset_tag (id, mtime, asset_id, tag_id, synced) VALUES (?, ?, ?, ?, 0)', 
            [assetTagId, now, assetId, tag.id]);
        }
      }
      
      // Create initial checkpoint if not a link and template was provided
      // This mirrors the Go implementation which creates a checkpoint with the template's chunks
      if (!isLink && templateId && templateChunks) {
        const checkpointComment = comment || 'Asset created';
        const authorId = getCurrentUserId();
        execute(db, `
          INSERT INTO asset_checkpoint (id, created_at, mtime, asset_id, xxhash_checksum, 
                                       time_modified, file_size, chunks, comment, 
                                       author_id, group_id, preview_id, trashed, synced)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
        `, [
          checkpointId, createdAt, now, assetId, templateChecksum,
          Math.floor(now / 1000), templateFileSize, templateChunks, checkpointComment,
          authorId, checkpointGroupId, '', // author_id from current user, group_id for batch, no preview
        ]);
      }
      
      await persistDatabase(projectName);
      
      // Return the created asset object
      const { statusMap, assetTypeMap, tagMap, assetTagsMap, assetDependenciesMap, collectionDependenciesMap, collectionMap } = buildLookupMaps(db);
      const row = queryOne(db, 'SELECT * FROM asset WHERE id = ?', [assetId]);
      return rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMap);
    } catch (error) {
      console.error('CreateAsset error:', error);
      throw error;
    }
  },

  // Updates an existing asset (local-first approach)
  // Signature matches Wails binding: UpdateAsset(projectPath, assetId, name, assetTypeId, isResource, pointer, tags)
  UpdateAsset: async (projectPath, assetId, name, assetTypeId, isResource, pointer, tags) => {
    const projectName = getProjectName(projectPath);
    
    // Validate assetId is present
    if (!assetId) {
      console.error('UpdateAsset error: assetId is required');
      throw new Error('assetId is required for update');
    }
    
    try {
      const db = await getDatabase(projectName);
      
      execute(db, `
        UPDATE asset SET name = ?, asset_type_id = ?, is_resource = ?, pointer = ?, mtime = ?, synced = 0
        WHERE id = ?
      `, [
        name ?? '', 
        assetTypeId ?? null, 
        isResource ? 1 : 0, 
        pointer ?? '',
        Date.now(), 
        assetId
      ]);
      await persistDatabase(projectName);
      
      // TODO: Handle tags update separately if needed
      
      // Return the updated asset
      const { statusMap, assetTypeMap, tagMap, assetTagsMap, assetDependenciesMap, collectionDependenciesMap, collectionMap } = buildLookupMaps(db);
      const row = queryOne(db, 'SELECT * FROM asset WHERE id = ?', [assetId]);
      return rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMap);
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
        execute(db, 'UPDATE asset SET trashed = 1, synced = 0, mtime = ? WHERE id = ?', [Date.now(), assetId]);
      } else {
        execute(db, 'DELETE FROM asset WHERE id = ?', [assetId]);
      }
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteAsset error:', error);
      throw error;
    }
  },

  // Changes an asset's collection (collection) - local-first approach
  ChangeAssetCollection: async (projectPath, assetId, newCollectionId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE asset SET collection_id = ?, mtime = ?, synced = 0 WHERE id = ?', [newCollectionId, Date.now(), assetId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeAssetCollection error:', error);
      throw error;
    }
  },

  // Adds a asset dependency (local-first approach)
  AddAssetDependency: async (projectPath, assetId, dependencyId, dependencyTypeId) => {
    const projectName = getProjectName(projectPath);
    const id = crypto.randomUUID();
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO asset_dependency (id, mtime, asset_id, dependency_id, dependency_type_id, synced)
        VALUES (?, ?, ?, ?, ?, 0)
      `, [id, Date.now(), assetId, dependencyId, dependencyTypeId]);
      await persistDatabase(projectName);
      
      return { id, asset_id: assetId, dependency_id: dependencyId, dependency_type_id: dependencyTypeId };
    } catch (error) {
      console.error('AddAssetDependency error:', error);
      throw error;
    }
  },

  // Adds an collection dependency (local-first approach)
  AddCollectionDependency: async (projectPath, assetId, collectionId, dependencyTypeId) => {
    const projectName = getProjectName(projectPath);
    const id = crypto.randomUUID();
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO collection_dependency (id, mtime, asset_id, collection_id, dependency_type_id, synced)
        VALUES (?, ?, ?, ?, ?, 0)
      `, [id, Date.now(), assetId, collectionId, dependencyTypeId]);
      await persistDatabase(projectName);
      
      return { id, asset_id: assetId, collection_id: collectionId, dependency_type_id: dependencyTypeId };
    } catch (error) {
      console.error('AddCollectionDependency error:', error);
      throw error;
    }
  },

  // Returns all asset types for a project
  GetAssetTypes: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      return query(db, 'SELECT * FROM asset_type');
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
      execute(db, 'INSERT INTO asset_type (id, mtime, name, icon, synced) VALUES (?, ?, ?, ?, 0)', 
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
      execute(db, 'UPDATE asset_type SET name = ?, icon = ?, mtime = ?, synced = 0 WHERE id = ?',
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
      execute(db, 'DELETE FROM asset_type WHERE id = ?', [typeId]);
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

  // Returns full asset/collection objects for a list of dependency IDs
  GetAssetDependencies: async (projectPath, dependencyIds) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      if (!dependencyIds || dependencyIds.length === 0) {
        return [];
      }
      
      const result = [];
      
      // Build lookup maps for asset enrichment
      const { statusMap, assetTypeMap, tagMap, assetTagsMap, assetDependenciesMap, collectionDependenciesMap, collectionMap } = buildLookupMaps(db);
      
      // Build collection type map
      const collectionTypeRows = query(db, 'SELECT * FROM collection_type');
      const collectionTypeMap = {};
      for (const et of collectionTypeRows) {
        collectionTypeMap[et.id] = et;
      }
      
      // Try to find assets first
      const placeholders = dependencyIds.map(() => '?').join(',');
      const assetQuery = `SELECT * FROM asset WHERE id IN (${placeholders}) AND trashed = 0 ORDER BY name`;
      const assetRows = query(db, assetQuery, dependencyIds);
      
      const foundAssetIds = new Set();
      for (const row of assetRows) {
        foundAssetIds.add(row.id);
        const asset = rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMap);
        result.push(asset);
      }
      
      // Find IDs that didn't match any assets - they might be collections
      const missingIds = dependencyIds.filter(id => !foundAssetIds.has(id));
      
      if (missingIds.length > 0) {
        const collectionPlaceholders = missingIds.map(() => '?').join(',');
        const collectionQuery = `SELECT * FROM collection WHERE id IN (${collectionPlaceholders}) AND trashed = 0 ORDER BY name`;
        const collectionRows = query(db, collectionQuery, missingIds);
        
        for (const row of collectionRows) {
          const collection = {
            ...row,
            trashed: !!row.trashed,
            is_trashed: !!row.trashed,
            synced: !!row.synced,
            type: 'collection',
            collection_type: collectionTypeMap[row.collection_type_id]?.name || '',
            collection_type_icon: collectionTypeMap[row.collection_type_id]?.icon || '',
          };
          result.push(collection);
        }
      }
      
      return result;
    } catch (error) {
      console.error('GetAssetDependencies error:', error);
      return [];
    }
  },

  // Renames an asset (local-first approach)
  RenameAsset: async (projectPath, assetId, newName) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE asset SET name = ?, mtime = ?, synced = 0 WHERE id = ?', [newName, Date.now(), assetId]);
      await persistDatabase(projectName);
      
      // Return updated asset
      const { statusMap, assetTypeMap, tagMap, assetTagsMap, assetDependenciesMap, collectionDependenciesMap, collectionMap } = buildLookupMaps(db);
      const row = queryOne(db, 'SELECT * FROM asset WHERE id = ?', [assetId]);
      return rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMap);
    } catch (error) {
      console.error('RenameAsset error:', error);
      throw error;
    }
  },

  // Duplicates an asset (local-first approach)
  DuplicateAsset: async (projectPath, sourceAssetId) => {
    const projectName = getProjectName(projectPath);
    const newAssetId = crypto.randomUUID();
    const now = Date.now();
    const createdAt = new Date().toISOString();
    
    try {
      const db = await getDatabase(projectName);
      
      // Get source asset
      const source = queryOne(db, 'SELECT * FROM asset WHERE id = ?', [sourceAssetId]);
      if (!source) throw new Error('Source asset not found');
      
      // Generate unique duplicate name (like Go: baseName + "-duplicate", increment if exists)
      const baseName = source.name;
      let duplicateName = `${baseName}-duplicate`;
      let counter = 1;
      while (queryOne(db, 'SELECT id FROM asset WHERE name = ? AND collection_id = ?', [duplicateName, source.collection_id])) {
        counter++;
        duplicateName = `${baseName}-duplicate-${counter}`;
      }
      
      // Insert duplicated asset with new ID
      execute(db, `
        INSERT INTO asset (id, mtime, created_at, name, description, extension, is_resource, 
                         status_id, asset_type_id, collection_id, assignee_id, assigner_id, 
                         is_link, pointer, preview_id, trashed, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
      `, [
        newAssetId, now, createdAt, 
        duplicateName, source.description || '', source.extension || '', source.is_resource,
        source.status_id || '', source.asset_type_id || '', source.collection_id || '', '',
        '', source.is_link, source.pointer || '', ''
      ]);
      
      // Copy tags from source asset
      const sourceTags = query(db, 'SELECT tag_id FROM asset_tag WHERE asset_id = ?', [sourceAssetId]);
      for (const tag of sourceTags) {
        const newTagId = crypto.randomUUID();
        execute(db, 'INSERT INTO asset_tag (id, mtime, asset_id, tag_id, synced) VALUES (?, ?, ?, ?, 0)', 
          [newTagId, now, newAssetId, tag.tag_id]);
      }
      
      // Copy latest checkpoint from source asset
      const latestCheckpoint = queryOne(db, 
        'SELECT * FROM asset_checkpoint WHERE asset_id = ? ORDER BY created_at DESC LIMIT 1', 
        [sourceAssetId]
      );
      
      if (latestCheckpoint) {
        const newCheckpointId = crypto.randomUUID();
        const newGroupId = crypto.randomUUID();
        const authorId = getCurrentUserId();
        
        execute(db, `
          INSERT INTO asset_checkpoint (id, created_at, mtime, asset_id, xxhash_checksum, 
                                       time_modified, file_size, chunks, comment, 
                                       author_id, group_id, preview_id, trashed, synced)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
        `, [
          newCheckpointId, createdAt, now, newAssetId,
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
      
      // Return the new asset
      const { statusMap, assetTypeMap, tagMap, assetTagsMap, assetDependenciesMap, collectionDependenciesMap, collectionMap } = buildLookupMaps(db);
      const row = queryOne(db, 'SELECT * FROM asset WHERE id = ?', [newAssetId]);
      return rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMap);
    } catch (error) {
      console.error('DuplicateAsset error:', error);
      throw error;
    }
  },

  // Changes the status of one or more assets (remote-first approach)
  ChangeStatus: async (projectPath, assetIds, statusId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      for (const assetId of assetIds) {
        execute(db, 'UPDATE asset SET status_id = ?, mtime = ?, synced = 0 WHERE id = ?', [statusId, Date.now(), assetId]);
      }
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeStatus error:', error);
      throw error;
    }
  },

  // Changes the type of an asset (local-first approach)
  ChangeAssetType: async (projectPath, assetId, assetTypeId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE asset SET asset_type_id = ?, mtime = ?, synced = 0 WHERE id = ?', [assetTypeId, Date.now(), assetId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeAssetType error:', error);
      throw error;
    }
  },

  // Toggles whether an asset is a asset or resource (local-first approach)
  ToggleIsAsset: async (projectPath, assetId, isAsset) => {
    const projectName = getProjectName(projectPath);
    
    // isAsset means it's NOT a resource
    const isResource = !isAsset;
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE asset SET is_resource = ?, mtime = ?, synced = 0 WHERE id = ?', [isResource ? 1 : 0, Date.now(), assetId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ToggleIsAsset error:', error);
      throw error;
    }
  },

  // Toggles whether assets are resources (local-first approach)
  ToggleIsResource: async (projectPath, assetIds, isResource) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      const placeholders = assetIds.map(() => '?').join(',');
      execute(db, `UPDATE asset SET is_resource = ?, mtime = ?, synced = 0 WHERE id IN (${placeholders})`, [isResource ? 1 : 0, Date.now(), ...assetIds]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ToggleIsResource error:', error);
      throw error;
    }
  },

  // Assigns a user to an asset (local-first approach)
  AssignAsset: async (projectPath, assetId, userId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE asset SET assignee_id = ?, mtime = ?, synced = 0 WHERE id = ?', [userId, Date.now(), assetId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('AssignAsset error:', error);
      throw error;
    }
  },

  // Unassigns a user from an asset (local-first approach)
  UnassignAsset: async (projectPath, assetId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE asset SET assignee_id = NULL, mtime = ?, synced = 0 WHERE id = ?', [Date.now(), assetId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UnassignAsset error:', error);
      throw error;
    }
  },

  // Unassigns multiple assets (local-first approach)
  UnassignAssets: async (projectPath, assetIds) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      const placeholders = assetIds.map(() => '?').join(',');
      execute(db, `UPDATE asset SET assignee_id = NULL, mtime = ?, synced = 0 WHERE id IN (${placeholders})`, [Date.now(), ...assetIds]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UnassignAssets error:', error);
      throw error;
    }
  },

  // Removes a asset dependency (local-first approach)
  RemoveAssetDependency: async (projectPath, assetId, dependencyId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM asset_dependency WHERE asset_id = ? AND dependency_id = ?', [assetId, dependencyId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('RemoveAssetDependency error:', error);
      throw error;
    }
  },

  // Removes an collection dependency (local-first approach)
  RemoveCollectionDependency: async (projectPath, assetId, collectionId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM collection_dependency WHERE asset_id = ? AND collection_id = ?', [assetId, collectionId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('RemoveCollectionDependency error:', error);
      throw error;
    }
  },

  // Adds a preview to an asset (local-first approach)
  // In web mode, preview path would be a URL/blob reference
  AddPreview: async (projectPath, assetId, previewPath) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE asset SET preview_id = ?, mtime = ?, synced = 0 WHERE id = ?', [previewPath, Date.now(), assetId]);
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
      const { statusMap, assetTypeMap, tagMap, assetTagsMap, assetDependenciesMap, collectionDependenciesMap, collectionMap } = buildLookupMaps(db);
      
      const row = queryOne(db, 'SELECT * FROM asset WHERE id = ?', [assetId]);
      return rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMap) || {};
    } catch (error) {
      console.error('GetAssetByID error:', error);
      return {};
    }
  },

  // Returns asset by its asset_path.
  GetAssetByPath: async (projectPath, assetPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const { statusMap, assetTypeMap, tagMap, assetTagsMap, assetDependenciesMap, collectionDependenciesMap, collectionMap } = buildLookupMaps(db);

      const rows = query(db, 'SELECT * FROM asset WHERE trashed = 0');
      for (const row of rows) {
        const asset = rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMap);
        if (asset && asset.asset_path === assetPath) {
          return asset;
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
      const row = queryOne(db, 'SELECT COUNT(*) as count FROM asset WHERE trashed = 0');
      return row?.count || 0;
    } catch (error) {
      console.error('GetAssetCount error:', error);
      return 0;
    }
  },

  // Returns all asset-type assets
  GetAssetAssets: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const { statusMap, assetTypeMap, tagMap, assetTagsMap, assetDependenciesMap, collectionDependenciesMap, collectionMap } = buildLookupMaps(db);
      
      const rows = query(db, 'SELECT * FROM asset WHERE is_resource = 0 AND trashed = 0');
      return rows.map(row => rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMap));
    } catch (error) {
      console.error('GetAssetAssets error:', error);
      return [];
    }
  },

  // Returns file status for an asset - not available in web mode
  AssetFileStatus: async (projectPath, assetId) => {
    return 'normal';
  },

  // Returns file statuses for multiple assets - not available in web mode
  AssetFilesStatus: async (projectPath, assetIds) => {
    const result = {};
    for (const id of assetIds) {
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
  RevealAsset: async (projectPath, assetId) => {
    console.warn('RevealAsset not available in web mode (no file system)');
    return {};
  },

  // Returns recursive dependencies for a asset
  GetRecursiveDependencies: async (projectPath, assetId, maxDepth = 5) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const { statusMap, assetTypeMap, tagMap, assetTagsMap, assetDependenciesMap, collectionDependenciesMap, collectionMap } = buildLookupMaps(db);
      
      const visited = new Set();
      const result = [];
      
      const collectDependencies = (id, depth) => {
        if (depth > maxDepth || visited.has(id)) return;
        visited.add(id);
        
        const deps = assetDependenciesMap[id] || [];
        for (const dep of deps) {
          const row = queryOne(db, 'SELECT * FROM asset WHERE id = ? AND trashed = 0', [dep.id]);
          if (row) {
            result.push(rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMap));
            collectDependencies(dep.id, depth + 1);
          }
        }
      };
      
      collectDependencies(assetId, 0);
      return result;
    } catch (error) {
      console.error('GetRecursiveDependencies error:', error);
      return [];
    }
  },
};
