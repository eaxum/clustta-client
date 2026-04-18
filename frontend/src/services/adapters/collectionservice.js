import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
}

/**
 * Compute collection_path based on parent's path and collection name
 * Mirrors the SQLite trigger logic from schema.sql
 * @param {object} db - Database connection
 * @param {string} parentId - Parent collection ID (empty string for root)
 * @param {string} name - Collection name
 * @returns {string} - Computed collection path like '/Parent/Child/'
 */
function computeCollectionPath(db, parentId, name) {
  if (!parentId || parentId === '') {
    // Root level collection
    return '/' + name + '/';
  }
  // Get parent's collection_path
  const parent = queryOne(db, 'SELECT collection_path FROM collection WHERE id = ?', [parentId]);
  if (parent && parent.collection_path) {
    return parent.collection_path + name + '/';
  }
  // Fallback if parent not found
  return '/' + name + '/';
}

/**
 * Convert database row to collection object with proper types
 * Optionally enriches with collection type info
 */
function rowToCollection(row, collectionTypeMap = {}) {
  if (!row) return null;
  const collectionType = collectionTypeMap[row.collection_type_id] || {};
  return {
    ...row,
    trashed: !!row.trashed,
    is_trashed: !!row.trashed,
    is_shared: !!row.is_shared,
    synced: !!row.synced,
    type: 'collection',
    collection_type: collectionType.name || '',
    collection_type_icon: collectionType.icon || '',
  };
}

/**
 * Convert database row to asset object with proper types
 * Optionally enriches with status, tags, collection type info, and dependencies
 * Also computes collection_path and asset_path from collection
 */
function rowToAsset(row, statusMap = {}, assetTypeMap = {}, assetTagsMap = {}, tagMap = {}, assetDependenciesMap = {}, collectionDependenciesMap = {}, collectionMap = {}) {
  if (!row) return null;
  const status = statusMap[row.status_id] || {};
  const assetType = assetTypeMap[row.asset_type_id] || {};
  const tagIds = assetTagsMap[row.id] || [];
  const tags = tagIds.map(tagId => tagMap[tagId]?.name || tagId).filter(Boolean);
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

export const CollectionService = {
  // Returns all collections for a project
  GetCollections: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      // Build collection type lookup map
      const collectionTypeRows = query(db, 'SELECT * FROM collection_type');
      const collectionTypeMap = {};
      for (const et of collectionTypeRows) {
        collectionTypeMap[et.id] = et;
      }
      
      const rows = query(db, 'SELECT * FROM collection WHERE trashed = 0');
      return rows.map(row => rowToCollection(row, collectionTypeMap));
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
      
      // Build collection type lookup map
      const collectionTypeRows = query(db, 'SELECT * FROM collection_type');
      const collectionTypeMap = {};
      for (const et of collectionTypeRows) {
        collectionTypeMap[et.id] = et;
      }
      
      const row = queryOne(db, 'SELECT * FROM collection WHERE id = ?', [collectionId]);
      return rowToCollection(row, collectionTypeMap) || {};
    } catch (error) {
      console.error('GetCollection error:', error);
      return {};
    }
  },

  // Alias for GetCollection - retrieves a collection by its ID
  GetCollectionByID: async (projectPath, collectionId) => {
    // Handle root as empty string
    const normalizedId = collectionId === 'root' ? '' : collectionId;
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      // Build collection type lookup map
      const collectionTypeRows = query(db, 'SELECT * FROM collection_type');
      const collectionTypeMap = {};
      for (const et of collectionTypeRows) {
        collectionTypeMap[et.id] = et;
      }
      
      const row = queryOne(db, 'SELECT * FROM collection WHERE id = ?', [normalizedId]);
      return rowToCollection(row, collectionTypeMap) || {};
    } catch (error) {
      console.error('GetCollectionByID error:', error);
      return {};
    }
  },

  // Returns a specific collection by its path
  GetCollectionByPath: async (projectPath, collectionPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      // Build collection type lookup map
      const collectionTypeRows = query(db, 'SELECT * FROM collection_type');
      const collectionTypeMap = {};
      for (const et of collectionTypeRows) {
        collectionTypeMap[et.id] = et;
      }
      
      // Handle root path
      const normalizedPath = collectionPath === '/' ? '' : collectionPath;
      
      const row = queryOne(db, 'SELECT * FROM collection WHERE collection_path = ?', [normalizedPath]);
      return rowToCollection(row, collectionTypeMap) || {};
    } catch (error) {
      console.error('GetCollectionByPath error:', error);
      return {};
    }
  },

  // Returns all children of a collection (collections, assets, untracked items)
  GetCollectionChildren: async (projectPath, collectionId, projectWorkingDir, collectionFolderPath, ignoreList, isUntracked) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      // Build lookup maps for enrichment
      const collectionTypeRows = query(db, 'SELECT * FROM collection_type');
      const collectionTypeMap = {};
      for (const et of collectionTypeRows) {
        collectionTypeMap[et.id] = et;
      }
      
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
      
      // Build asset dependency map
      const assetDependencyRows = query(db, 'SELECT * FROM asset_dependency');
      const assetDependenciesMap = {};
      for (const td of assetDependencyRows) {
        if (!assetDependenciesMap[td.asset_id]) {
          assetDependenciesMap[td.asset_id] = [];
        }
        assetDependenciesMap[td.asset_id].push({
          id: td.dependency_id,
          type_id: td.dependency_type_id
        });
      }
      
      // Build collection dependency map
      const collectionDependencyRows = query(db, 'SELECT * FROM collection_dependency');
      const collectionDependenciesMap = {};
      for (const ed of collectionDependencyRows) {
        if (!collectionDependenciesMap[ed.asset_id]) {
          collectionDependenciesMap[ed.asset_id] = [];
        }
        collectionDependenciesMap[ed.asset_id].push({
          id: ed.collection_id,
          type_id: ed.dependency_type_id
        });
      }
      
      // Get child collections
      const parentCondition = collectionId === 'root' 
        ? "(parent_id IS NULL OR parent_id = '')"
        : `parent_id = '${collectionId}'`;
      
      const collectionRows = query(db, `SELECT * FROM collection WHERE ${parentCondition} AND trashed = 0`);
      const collections = collectionRows.map(row => rowToCollection(row, collectionTypeMap));
      
      // Build collectionMap for asset_path computation
      const allCollectionRows = query(db, 'SELECT id, name, collection_path FROM collection');
      const collectionMapForAssets = {};
      for (const e of allCollectionRows) {
        collectionMapForAssets[e.id] = e;
      }
      
      // Get child assets
      const assetCondition = collectionId === 'root'
        ? "(collection_id IS NULL OR collection_id = '')"
        : `collection_id = '${collectionId}'`;
      
      const assetRows = query(db, `SELECT * FROM asset WHERE ${assetCondition} AND trashed = 0`);
      const assets = assetRows.map(row => rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMapForAssets));
      
      return {
        collections,
        assets,
        untracked_assets: [],      // No untracked items in web mode
        untracked_collections: [],   // No untracked items in web mode
      };
    } catch (error) {
      console.error('GetCollectionChildren error:', error);
      return {
        collections: [],
        assets: [],
        untracked_assets: [],
        untracked_collections: [],
      };
    }
  },

  // Returns state flags for a collection's children
  GetCollectionChildrenState: async (projectPath, collectionId, projectWorkingDir, ignoreList) => {
    // In web mode, we don't track local file states
    return {
      modified_assets: [],
      outdated_assets: [],
      rebuildable_assets: [],
      normal_assets: [],
      untracked_files: [],
      untracked_folders: [],
    };
  },

  // Returns collection state flags (has_untracked, has_modified, has_outdated, has_rebuildable)
  GetCollectionStateFlags: async (projectPath, collectionId, projectWorkingDir, ignoreList) => {
    // In web mode, we don't track local file states, so return all false
    return {
      has_untracked: false,
      has_modified: false,
      has_outdated: false,
      has_rebuildable: false,
    };
  },

  // Changes the parent of a collection (local-first approach)
  ChangeCollectionParent: async (projectPath, collectionId, newParentId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      
      // Get collection's current name to recompute path
      const collection = queryOne(db, 'SELECT name, collection_path FROM collection WHERE id = ?', [collectionId]);
      if (!collection) {
        throw new Error('Collection not found');
      }
      
      // Compute new collection_path based on new parent
      const newCollectionPath = computeCollectionPath(db, newParentId, collection.name);
      const oldCollectionPath = collection.collection_path;
      
      // Update collection with new parent and path
      execute(db, 'UPDATE collection SET parent_id = ?, collection_path = ?, mtime = ?, synced = 0 WHERE id = ?', 
        [newParentId, newCollectionPath, Date.now(), collectionId]);
      
      // Update all descendants' collection_paths
      if (oldCollectionPath) {
        const descendants = query(db, "SELECT id, name, collection_path FROM collection WHERE collection_path LIKE ? AND id != ?", 
          [oldCollectionPath + '%', collectionId]);
        for (const desc of descendants) {
          const newDescPath = newCollectionPath + desc.collection_path.substring(oldCollectionPath.length);
          execute(db, 'UPDATE collection SET collection_path = ?, mtime = ?, synced = 0 WHERE id = ?', 
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
  // Signature matches Wails: CreateCollection(projectPath, name, description, collectionTypeId, parentId, previewPath, isShared)
  CreateCollection: async (projectPath, name, description, collectionTypeId, parentId, previewPath, isShared) => {
    const projectName = getProjectName(projectPath);
    const id = crypto.randomUUID();
    const now = Date.now();
    const createdAt = new Date().toISOString();
    
    try {
      const db = await getDatabase(projectName);
      
      // Compute collection_path based on parent (mimics SQLite trigger behavior)
      const collectionPath = computeCollectionPath(db, parentId, name || '');
      
      execute(db, `
        INSERT INTO collection (id, created_at, mtime, name, description, collection_type_id, parent_id, collection_path, preview_id, is_shared, trashed, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
      `, [id, createdAt, now, name || '', description || '', collectionTypeId || '', parentId || '', collectionPath, previewPath || '', isShared ? 1 : 0]);
      await persistDatabase(projectName);
      
      // Build collection type lookup map
      const collectionTypeRows = query(db, 'SELECT * FROM collection_type');
      const collectionTypeMap = {};
      for (const et of collectionTypeRows) {
        collectionTypeMap[et.id] = et;
      }
      
      // Return the created collection
      const row = queryOne(db, 'SELECT * FROM collection WHERE id = ?', [id]);
      return rowToCollection(row, collectionTypeMap);
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
        UPDATE collection SET name = ?, parent_id = ?, collection_type_id = ?, collection_path = ?, mtime = ?, synced = 0
        WHERE id = ?
      `, [collection.name || '', collection.parent_id || '', collection.collection_type_id || '', collection.collection_path || '', Date.now(), collection.id]);
      await persistDatabase(projectName);
      
      // Build collection type lookup map
      const collectionTypeRows = query(db, 'SELECT * FROM collection_type');
      const collectionTypeMap = {};
      for (const et of collectionTypeRows) {
        collectionTypeMap[et.id] = et;
      }
      
      // Return the updated collection
      const row = queryOne(db, 'SELECT * FROM collection WHERE id = ?', [collection.id]);
      return rowToCollection(row, collectionTypeMap);
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
        execute(db, 'UPDATE collection SET trashed = 1, mtime = ?, synced = 0 WHERE id = ?', [Date.now(), collectionId]);
      } else {
        execute(db, 'DELETE FROM collection WHERE id = ?', [collectionId]);
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
      return query(db, 'SELECT * FROM collection_type');
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
      execute(db, 'INSERT INTO collection_type (id, mtime, name, icon, synced) VALUES (?, ?, ?, ?, 0)', 
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
      execute(db, 'UPDATE collection_type SET name = ?, icon = ?, mtime = ?, synced = 0 WHERE id = ?',
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
      execute(db, 'DELETE FROM collection_type WHERE id = ?', [typeId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteCollectionType error:', error);
      throw error;
    }
  },

  // Renames an existing collection (local-first approach)
  RenameCollection: async (projectPath, collectionId, newName) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      
      // Get collection's current info to recompute path
      const collection = queryOne(db, 'SELECT name, parent_id, collection_path FROM collection WHERE id = ?', [collectionId]);
      if (!collection) {
        throw new Error('Collection not found');
      }
      
      // Compute new collection_path with the new name
      const newCollectionPath = computeCollectionPath(db, collection.parent_id, newName);
      const oldCollectionPath = collection.collection_path;
      
      // Update collection with new name and path
      execute(db, 'UPDATE collection SET name = ?, collection_path = ?, mtime = ?, synced = 0 WHERE id = ?', 
        [newName, newCollectionPath, Date.now(), collectionId]);
      
      // Update all descendants' collection_paths
      if (oldCollectionPath) {
        const descendants = query(db, "SELECT id, collection_path FROM collection WHERE collection_path LIKE ? AND id != ?", 
          [oldCollectionPath + '%', collectionId]);
        for (const desc of descendants) {
          const newDescPath = newCollectionPath + desc.collection_path.substring(oldCollectionPath.length);
          execute(db, 'UPDATE collection SET collection_path = ?, mtime = ?, synced = 0 WHERE id = ?', 
            [newDescPath, Date.now(), desc.id]);
        }
      }
      
      await persistDatabase(projectName);
      
      // Build collection type lookup map
      const collectionTypeRows = query(db, 'SELECT * FROM collection_type');
      const collectionTypeMap = {};
      for (const et of collectionTypeRows) {
        collectionTypeMap[et.id] = et;
      }
      
      // Return the updated collection
      const row = queryOne(db, 'SELECT * FROM collection WHERE id = ?', [collectionId]);
      return rowToCollection(row, collectionTypeMap);
    } catch (error) {
      console.error('RenameCollection error:', error);
      throw error;
    }
  },

  // Changes the type of a collection (local-first approach)
  ChangeType: async (projectPath, collectionId, collectionTypeId) => {
    const projectName = getProjectName(projectPath);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE collection SET collection_type_id = ?, mtime = ?, synced = 0 WHERE id = ?', [collectionTypeId, Date.now(), collectionId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeType error:', error);
      throw error;
    }
  },

  // Toggles the shared flag on a collection (local-first approach)
  ChangeIsShared: async (projectPath, collectionId, isShared) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE collection SET is_shared = ?, mtime = ?, synced = 0 WHERE id = ?', 
        [isShared ? 1 : 0, Date.now(), collectionId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('ChangeIsShared error:', error);
      throw error;
    }
  },

  // Assigns a user to a collection (local-first approach)
  // Uses collection_assignee junction table
  Assign: async (projectPath, collectionId, userId) => {
    const projectName = getProjectName(projectPath);
    const id = crypto.randomUUID();
    const now = Date.now();
    try {
      const db = await getDatabase(projectName);
      // Check if already assigned
      const existing = queryOne(db, 'SELECT id FROM collection_assignee WHERE collection_id = ? AND assignee_id = ?', [collectionId, userId]);
      if (!existing) {
        execute(db, `
          INSERT INTO collection_assignee (id, mtime, collection_id, assignee_id, assigner_id, synced)
          VALUES (?, ?, ?, ?, ?, 0)
        `, [id, now, collectionId, userId, '']);
        await persistDatabase(projectName);
      }
    } catch (error) {
      console.error('Assign error:', error);
      throw error;
    }
  },

  // Unassigns a user from a collection (local-first approach)
  // Uses collection_assignee junction table
  Unassign: async (projectPath, collectionId, userId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM collection_assignee WHERE collection_id = ? AND assignee_id = ?', [collectionId, userId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('Unassign error:', error);
      throw error;
    }
  },

  // Updates the preview image for a collection (local-first approach)
  UpdatePreview: async (projectPath, collectionId, previewPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      execute(db, 'UPDATE collection SET preview_id = ?, mtime = ?, synced = 0 WHERE id = ?', 
        [previewPath || '', Date.now(), collectionId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdatePreview error:', error);
      throw error;
    }
  },

  // Reveals a collection in the file explorer - not available in web mode
  RevealCollection: async (projectPath, collectionId) => {
    console.warn('RevealCollection not available in web mode (no file system)');
    return {};
  },

  // Reverts collections - not available in web mode
  RevertCollections: async (projectPath, collectionIds) => {
    console.warn('RevertCollections not available in web mode (no file system)');
    return {};
  },

  // Returns items for checkpoint - not available in web mode
  GetItemsForCheckpoint: async (projectPath, collectionId, targetPath, projectWorkingDir, ignoreList) => {
    console.warn('GetItemsForCheckpoint not available in web mode (no file system)');
    return [];
  },

  // Returns outdated items in collection - not available in web mode
  GetOutdatedItemsInCollection: async (projectPath, collectionId, projectWorkingDir, ignoreList) => {
    console.warn('GetOutdatedItemsInCollection not available in web mode (no file system)');
    return [];
  },

  // Returns collection count
  GetCollectionCount: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const row = queryOne(db, 'SELECT COUNT(*) as count FROM collection WHERE trashed = 0');
      return row?.count || 0;
    } catch (error) {
      console.error('GetCollectionCount error:', error);
      return 0;
    }
  },

  // Returns assets for a collection
  GetCollectionAssets: async (projectPath, collectionId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      // Build lookup maps for enrichment
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
      
      // Build asset dependency map
      const assetDependencyRows = query(db, 'SELECT * FROM asset_dependency');
      const assetDependenciesMap = {};
      for (const td of assetDependencyRows) {
        if (!assetDependenciesMap[td.asset_id]) {
          assetDependenciesMap[td.asset_id] = [];
        }
        assetDependenciesMap[td.asset_id].push({
          id: td.dependency_id,
          type_id: td.dependency_type_id
        });
      }
      
      // Build collection dependency map
      const collectionDependencyRows = query(db, 'SELECT * FROM collection_dependency');
      const collectionDependenciesMap = {};
      for (const ed of collectionDependencyRows) {
        if (!collectionDependenciesMap[ed.asset_id]) {
          collectionDependenciesMap[ed.asset_id] = [];
        }
        collectionDependenciesMap[ed.asset_id].push({
          id: ed.collection_id,
          type_id: ed.dependency_type_id
        });
      }
      
      // Build collectionMap for asset_path computation
      const collectionRows = query(db, 'SELECT id, name, collection_path FROM collection');
      const collectionMapForAssets = {};
      for (const e of collectionRows) {
        collectionMapForAssets[e.id] = e;
      }
      
      // Get assets for this collection
      const assetCondition = collectionId === 'root' || !collectionId
        ? "(collection_id IS NULL OR collection_id = '')"
        : `collection_id = '${collectionId}'`;
      
      const assetRows = query(db, `SELECT * FROM asset WHERE ${assetCondition} AND trashed = 0`);
      return assetRows.map(row => rowToAsset(row, statusMap, assetTypeMap, assetTagsMap, tagMap, assetDependenciesMap, collectionDependenciesMap, collectionMapForAssets));
    } catch (error) {
      console.error('GetCollectionAssets error:', error);
      return [];
    }
  },

  // Creates multiple collections at once (local-first approach)
  // Signature matches Wails: CreateCollections(projectPath, name, description, collectionTypeId, parentId)
  CreateCollections: async (projectPath, name, description, collectionTypeId, parentId) => {
    const projectName = getProjectName(projectPath);
    const id = crypto.randomUUID();
    const now = Date.now();
    const createdAt = new Date().toISOString();
    
    try {
      const db = await getDatabase(projectName);
      
      // Compute collection_path based on parent (mimics SQLite trigger behavior)
      const collectionPath = computeCollectionPath(db, parentId, name || '');
      
      execute(db, `
        INSERT INTO collection (id, created_at, mtime, name, description, collection_type_id, parent_id, collection_path, preview_id, is_shared, trashed, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
      `, [id, createdAt, now, name || '', description || '', collectionTypeId || '', parentId || '', collectionPath, '', 0]);
      await persistDatabase(projectName);
      
      // Build collection type lookup map
      const collectionTypeRows = query(db, 'SELECT * FROM collection_type');
      const collectionTypeMap = {};
      for (const et of collectionTypeRows) {
        collectionTypeMap[et.id] = et;
      }
      
      // Return the created collection
      const row = queryOne(db, 'SELECT * FROM collection WHERE id = ?', [id]);
      return [rowToCollection(row, collectionTypeMap)];
    } catch (error) {
      console.error('CreateCollections error:', error);
      throw error;
    }
  },

  // Checks if a user is assigned to a collection or any of its ancestor collections.
  IsUserAssignedToCollectionOrAncestor: async (projectPath, collectionId, userId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);

      // Check the collection itself first.
      const direct = queryOne(db,
        `SELECT EXISTS(SELECT 1 FROM collection_assignee WHERE collection_id = ? AND assignee_id = ?) AS result`,
        [collectionId, userId]
      );
      if (direct?.result === 1) return true;

      // Check ancestor collections recursively.
      const ancestor = queryOne(db, `
        WITH RECURSIVE ancestors AS (
          SELECT parent_id FROM collection WHERE id = ? AND parent_id != ''
          UNION ALL
          SELECT e.parent_id FROM collection e
          JOIN ancestors a ON e.id = a.parent_id
          WHERE a.parent_id != ''
        )
        SELECT EXISTS(
          SELECT 1 FROM collection_assignee ea
          JOIN ancestors a ON ea.collection_id = a.parent_id
          WHERE ea.assignee_id = ?
        ) AS result
      `, [collectionId, userId]);
      return ancestor?.result === 1;
    } catch (error) {
      console.error('IsUserAssignedToCollectionOrAncestor error:', error);
      return false;
    }
  },
};
