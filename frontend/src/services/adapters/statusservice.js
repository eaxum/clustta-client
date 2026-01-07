import { studioApiCall, getActiveStudioUrl } from './http-client.js';
import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
}

/**
 * Convert database row to status object
 */
function rowToStatus(row) {
  if (!row) return null;
  return {
    ...row,
    synced: !!row.synced,
  };
}

export const StatusService = {
  // Returns all statuses for a project
  GetStatuses: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const rows = query(db, 'SELECT * FROM status ORDER BY name');
      return rows.map(rowToStatus);
    } catch (error) {
      console.error('GetStatuses error:', error);
      return [];
    }
  },

  // Creates a new status
  CreateStatus: async (projectPath, status) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/status`, 'POST', status);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO status (id, mtime, name, color, sort_order, synced)
        VALUES (?, ?, ?, ?, ?, 1)
      `, [result.id, Date.now(), result.name, result.color || '', result.sort_order || 0]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('CreateStatus local insert error:', error);
    }
    
    return result;
  },

  // Updates an existing status
  UpdateStatus: async (projectPath, status) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/status/${status.id}`, 'PUT', status);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `UPDATE status SET name = ?, color = ?, sort_order = ?, mtime = ? WHERE id = ?`, 
        [status.name, status.color || '', status.sort_order || 0, Date.now(), status.id]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdateStatus local update error:', error);
    }
    
    return result;
  },

  // Deletes a status
  DeleteStatus: async (projectPath, statusId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    await studioApiCall(studioUrl, `/${projectName}/status/${statusId}`, 'DELETE');
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM status WHERE id = ?', [statusId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteStatus local update error:', error);
    }
  },
};
