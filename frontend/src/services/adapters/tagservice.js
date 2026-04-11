import { studioApiCall, getActiveStudioUrl } from './http-client.js';
import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
}

/**
 * Convert database row to tag object
 */
function rowToTag(row) {
  if (!row) return null;
  return {
    ...row,
    synced: !!row.synced,
  };
}

export const TagService = {
  // Returns all tags for a project
  GetTags: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const rows = query(db, 'SELECT * FROM tag ORDER BY name');
      return rows.map(rowToTag);
    } catch (error) {
      console.error('GetTags error:', error);
      return [];
    }
  },

  // Creates a new tag
  CreateTag: async (projectPath, tag) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/tag`, 'POST', tag);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO tag (id, mtime, name, color, synced)
        VALUES (?, ?, ?, ?, 1)
      `, [result.id, Date.now(), result.name, result.color || '']);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('CreateTag local insert error:', error);
    }
    
    return result;
  },

  // Updates an existing tag
  UpdateTag: async (projectPath, tag) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/tag/${tag.id}`, 'PUT', tag);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `UPDATE tag SET name = ?, color = ?, mtime = ? WHERE id = ?`, 
        [tag.name, tag.color || '', Date.now(), tag.id]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdateTag local update error:', error);
    }
    
    return result;
  },

  // Deletes a tag
  DeleteTag: async (projectPath, tagId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    await studioApiCall(studioUrl, `/${projectName}/tag/${tagId}`, 'DELETE');
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM tag WHERE id = ?', [tagId]);
      execute(db, 'DELETE FROM asset_tag WHERE tag_id = ?', [tagId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteTag local update error:', error);
    }
  },

  // Returns all tags for a specific asset
  GetAssetTags: async (projectPath, assetId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const rows = query(db, 'SELECT t.* FROM tag t INNER JOIN asset_tag at ON t.id = at.tag_id WHERE at.asset_id = ? ORDER BY t.name', [assetId]);
      return rows.map(rowToTag);
    } catch (error) {
      console.error('GetAssetTags error:', error);
      return [];
    }
  },

  // Adds a tag to an asset by name, creating the tag if it doesn't exist
  AddTagToAsset: async (projectPath, assetId, tagName) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();

    const result = await studioApiCall(studioUrl, `/${projectName}/asset/${assetId}/tag`, 'POST', { name: tagName });

    try {
      const db = await getDatabase(projectName);
      const rows = query(db, 'SELECT t.* FROM tag t INNER JOIN asset_tag at ON t.id = at.tag_id WHERE at.asset_id = ? ORDER BY t.name', [assetId]);
      return rows.map(rowToTag);
    } catch (error) {
      console.error('AddTagToAsset local query error:', error);
      return result || [];
    }
  },

  // Removes a tag from an asset by tag ID
  RemoveTagFromAsset: async (projectPath, assetId, tagId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();

    await studioApiCall(studioUrl, `/${projectName}/asset/${assetId}/tag/${tagId}`, 'DELETE');

    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM asset_tag WHERE asset_id = ? AND tag_id = ?', [assetId, tagId]);
      await persistDatabase(projectName);
      const rows = query(db, 'SELECT t.* FROM tag t INNER JOIN asset_tag at ON t.id = at.tag_id WHERE at.asset_id = ? ORDER BY t.name', [assetId]);
      return rows.map(rowToTag);
    } catch (error) {
      console.error('RemoveTagFromAsset local update error:', error);
      return [];
    }
  },
};
