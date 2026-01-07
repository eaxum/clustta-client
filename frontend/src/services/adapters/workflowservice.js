import { studioApiCall, getActiveStudioUrl } from './http-client.js';
import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
}

/**
 * Convert database row to workflow object, including links, entities, and tasks
 */
async function rowToWorkflow(db, row) {
  if (!row) return null;
  
  // Get workflow links
  const links = query(db, 'SELECT * FROM workflow_link WHERE workflow_id = ?', [row.id]);
  
  // Get workflow entities
  const entities = query(db, 'SELECT * FROM workflow_entity WHERE workflow_id = ?', [row.id]);
  
  // Get workflow tasks
  const tasks = query(db, 'SELECT * FROM workflow_task WHERE workflow_id = ?', [row.id]);
  
  return {
    ...row,
    synced: !!row.synced,
    links: links || [],
    entities: entities || [],
    tasks: tasks || [],
  };
}

export const WorkflowService = {
  // Returns all workflows for a project
  GetWorkflows: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const rows = query(db, 'SELECT * FROM workflow ORDER BY name');
      const workflows = await Promise.all(rows.map(row => rowToWorkflow(db, row)));
      return workflows;
    } catch (error) {
      console.error('GetWorkflows error:', error);
      return [];
    }
  },

  // Creates a new workflow
  CreateWorkflow: async (projectPath, workflow) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/workflow`, 'POST', workflow);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO workflow (id, mtime, name, synced)
        VALUES (?, ?, ?, 1)
      `, [result.id, Date.now(), result.name]);
      
      // Insert workflow links if provided
      if (result.links && result.links.length > 0) {
        for (const link of result.links) {
          execute(db, `
            INSERT INTO workflow_link (id, mtime, workflow_id, source_task_type_id, target_task_type_id, synced)
            VALUES (?, ?, ?, ?, ?, 1)
          `, [link.id, Date.now(), result.id, link.source_task_type_id, link.target_task_type_id]);
        }
      }
      
      await persistDatabase(projectName);
    } catch (error) {
      console.error('CreateWorkflow local insert error:', error);
    }
    
    return result;
  },

  // Updates an existing workflow
  UpdateWorkflow: async (projectPath, workflow) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/workflow/${workflow.id}`, 'PUT', workflow);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `UPDATE workflow SET name = ?, mtime = ? WHERE id = ?`, 
        [workflow.name, Date.now(), workflow.id]);
      
      // Update workflow links - delete existing and re-insert
      execute(db, 'DELETE FROM workflow_link WHERE workflow_id = ?', [workflow.id]);
      if (workflow.links && workflow.links.length > 0) {
        for (const link of workflow.links) {
          execute(db, `
            INSERT INTO workflow_link (id, mtime, workflow_id, source_task_type_id, target_task_type_id, synced)
            VALUES (?, ?, ?, ?, ?, 1)
          `, [link.id, Date.now(), workflow.id, link.source_task_type_id, link.target_task_type_id]);
        }
      }
      
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdateWorkflow local update error:', error);
    }
    
    return result;
  },

  // Deletes a workflow
  DeleteWorkflow: async (projectPath, workflowId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    await studioApiCall(studioUrl, `/${projectName}/workflow/${workflowId}`, 'DELETE');
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM workflow WHERE id = ?', [workflowId]);
      execute(db, 'DELETE FROM workflow_link WHERE workflow_id = ?', [workflowId]);
      execute(db, 'DELETE FROM workflow_entity WHERE workflow_id = ?', [workflowId]);
      execute(db, 'DELETE FROM workflow_task WHERE workflow_id = ?', [workflowId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteWorkflow local update error:', error);
    }
  },
};
