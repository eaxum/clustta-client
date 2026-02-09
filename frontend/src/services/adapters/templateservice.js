import { studioApiCall, getActiveStudioUrl } from './http-client.js';
import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
}

/**
 * Convert database row to template object
 */
function rowToTemplate(row) {
  if (!row) return null;
  return {
    ...row,
    file_size: Number(row.file_size || 0),
    trashed: !!row.trashed,
    synced: !!row.synced,
  };
}

export const TemplateService = {
  // Returns all templates for a project
  GetTemplates: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const rows = query(db, 'SELECT * FROM template WHERE trashed = 0 ORDER BY name');
      return rows.map(rowToTemplate);
    } catch (error) {
      console.error('GetTemplates error:', error);
      return [];
    }
  },

  // Creates a new template
  CreateTemplate: async (projectPath, template) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/template`, 'POST', template);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO template (id, mtime, name, extension, chunks, xxhash_checksum, file_size, trashed, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, 0, 1)
      `, [
        result.id, Date.now(), result.name, result.extension || '',
        result.chunks || '', result.xxhash_checksum || '', result.file_size || 0
      ]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('CreateTemplate local insert error:', error);
    }
    
    return result;
  },

  // Updates an existing template
  UpdateTemplate: async (projectPath, template) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/template/${template.id}`, 'PUT', template);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `UPDATE template SET name = ?, extension = ?, mtime = ? WHERE id = ?`, [
        template.name,
        template.extension || '',
        Date.now(),
        template.id
      ]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdateTemplate local update error:', error);
    }
    
    return result;
  },

  // Deletes a template
  DeleteTemplate: async (projectPath, templateId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    await studioApiCall(studioUrl, `/${projectName}/template/${templateId}`, 'DELETE');
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM template WHERE id = ?', [templateId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteTemplate local update error:', error);
    }
  },
};
