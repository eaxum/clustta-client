import { studioApiCall, getActiveStudioUrl } from './http-client.js';
import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
}

/**
 * Convert database row to dependency type object
 */
function rowToDependencyType(row) {
  if (!row) return null;
  return {
    ...row,
    synced: !!row.synced,
  };
}

export const DependencyTypeService = {
  // Returns all dependency types for a project
  GetDependencyTypes: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const rows = query(db, 'SELECT * FROM dependency_type ORDER BY name');
      return rows.map(rowToDependencyType);
    } catch (error) {
      console.error('GetDependencyTypes error:', error);
      return [];
    }
  },

  // Creates a new dependency type
  CreateDependencyType: async (projectPath, type) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/dependency-type`, 'POST', type);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO dependency_type (id, mtime, name, synced)
        VALUES (?, ?, ?, 1)
      `, [result.id, Date.now(), result.name]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('CreateDependencyType local insert error:', error);
    }
    
    return result;
  },

  // Updates an existing dependency type
  UpdateDependencyType: async (projectPath, type) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/dependency-type/${type.id}`, 'PUT', type);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `UPDATE dependency_type SET name = ?, mtime = ? WHERE id = ?`, 
        [type.name, Date.now(), type.id]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdateDependencyType local update error:', error);
    }
    
    return result;
  },

  // Deletes a dependency type
  DeleteDependencyType: async (projectPath, typeId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    await studioApiCall(studioUrl, `/${projectName}/dependency-type/${typeId}`, 'DELETE');
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM dependency_type WHERE id = ?', [typeId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteDependencyType local update error:', error);
    }
  },
};
