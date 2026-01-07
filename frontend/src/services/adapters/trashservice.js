import { studioApiCall, getActiveStudioUrl } from './http-client.js';
import { getDatabase, query, execute, persistDatabase } from './project-database.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
}

/**
 * Convert database row to trash item object
 */
function rowToTrashItem(row) {
  if (!row) return null;
  return {
    id: row.id,
    table_name: row.table_name,  // 'entity' or 'task'
    item_type: row.table_name,   // For compatibility
    dtime: row.dtime,
  };
}

export const TrashService = {
  // Returns all trashed items for a project (matches Go binding: GetTrashs)
  GetTrashs: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const rows = query(db, 'SELECT * FROM tomb ORDER BY dtime DESC');
      return rows.map(rowToTrashItem);
    } catch (error) {
      console.error('GetTrashs error:', error);
      return [];
    }
  },

  // Restores a deleted item by ID and type from the recycle bin
  Restore: async (projectPath, id, itemType) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/trash/${id}/restore`, 'POST', { item_type: itemType });
    
    // Remove from tomb table
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM tomb WHERE id = ? AND table_name = ?', [id, itemType]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('Restore local update error:', error);
    }
    
    return result;
  },
};
