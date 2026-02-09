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
      execute(db, 'DELETE FROM task_tag WHERE tag_id = ?', [tagId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteTag local update error:', error);
    }
  },
};
