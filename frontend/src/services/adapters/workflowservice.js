import { studioApiCall, getActiveStudioUrl } from './http-client.js';
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
 * Compute entity path from parent path and name
 */
function computeEntityPath(db, parentId, name) {
  if (!parentId || parentId === '') {
    return '/' + name + '/';
  }
  const parent = queryOne(db, 'SELECT entity_path FROM entity WHERE id = ?', [parentId]);
  if (parent && parent.entity_path) {
    return parent.entity_path + name + '/';
  }
  return '/' + name + '/';
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

  // Adds a workflow to an entity - creates entities and tasks based on workflow definition
  // This matches the Go implementation which creates actual entities and tasks
  AddWorkflow: async (projectPath, workflowId, name, entityTypeId, parentId) => {
    const projectName = getProjectName(projectPath);
    const now = Date.now();
    const createdAt = new Date().toISOString();
    const userId = getCurrentUserId();
    
    try {
      const db = await getDatabase(projectName);
      
      // Get the workflow definition
      const workflow = queryOne(db, 'SELECT * FROM workflow WHERE id = ?', [workflowId]);
      if (!workflow) {
        throw new Error('Workflow not found');
      }
      
      // Get workflow entities (child entity templates)
      const workflowEntities = query(db, 'SELECT * FROM workflow_entity WHERE workflow_id = ?', [workflowId]);
      
      // Get workflow tasks (task templates)
      const workflowTasks = query(db, 'SELECT * FROM workflow_task WHERE workflow_id = ?', [workflowId]);
      
      // Get workflow links (linked workflows)
      const workflowLinks = query(db, 'SELECT * FROM workflow_link WHERE workflow_id = ?', [workflowId]);
      
      // 1. Create the parent entity with the workflow name
      const parentEntityId = crypto.randomUUID();
      const parentEntityPath = computeEntityPath(db, parentId, name);
      execute(db, `
        INSERT INTO entity (id, created_at, mtime, name, description, entity_type_id, parent_id, entity_path, is_library, trashed, synced)
        VALUES (?, ?, ?, ?, '', ?, ?, ?, 0, 0, 0)
      `, [parentEntityId, createdAt, now, name, entityTypeId, parentId || '', parentEntityPath]);
      
      // 2. Create child entities from workflow entity templates
      for (const we of workflowEntities) {
        const childEntityId = crypto.randomUUID();
        const childEntityPath = computeEntityPath(db, parentEntityId, we.name);
        execute(db, `
          INSERT INTO entity (id, created_at, mtime, name, description, entity_type_id, parent_id, entity_path, is_library, trashed, synced)
          VALUES (?, ?, ?, ?, '', ?, ?, ?, 0, 0, 0)
        `, [childEntityId, createdAt, now, we.name, we.entity_type_id, parentEntityId, childEntityPath]);
      }
      
      // Get the default status (status with short_name = 'todo')
      const defaultStatus = queryOne(db, "SELECT id FROM status WHERE short_name = 'todo' LIMIT 1");
      const defaultStatusId = defaultStatus?.id || '';
      
      // 3. Create tasks from workflow task templates
      for (const wt of workflowTasks) {
        const taskId = crypto.randomUUID();
        const checkpointId = crypto.randomUUID();
        const groupId = crypto.randomUUID();
        
        // Get template for chunks and extension if available
        let chunks = '[]';
        let extension = wt.extension || '';
        if (wt.template_id) {
          const template = queryOne(db, 'SELECT chunks, extension FROM template WHERE id = ?', [wt.template_id]);
          if (template) {
            if (template.chunks) {
              chunks = template.chunks;
            }
            if (template.extension) {
              extension = template.extension;
            }
          }
        }
        
        // Create the task
        execute(db, `
          INSERT INTO task (id, mtime, created_at, name, description, extension, is_resource, 
                           status_id, task_type_id, entity_id, assignee_id, assigner_id, 
                           is_link, pointer, preview_id, trashed, synced)
          VALUES (?, ?, ?, ?, '', ?, ?, ?, ?, ?, '', '', ?, ?, '', 0, 0)
        `, [
          taskId, now, createdAt, wt.name, extension, wt.is_resource ? 1 : 0,
          defaultStatusId, wt.task_type_id, parentEntityId, wt.is_link ? 1 : 0, wt.pointer || ''
        ]);
        
        // Create initial checkpoint for the task
        execute(db, `
          INSERT INTO task_checkpoint (id, created_at, mtime, task_id, xxhash_checksum, 
                                       time_modified, file_size, chunks, comment, 
                                       author_id, group_id, preview_id, trashed, synced)
          VALUES (?, ?, ?, ?, '', ?, 0, ?, 'Asset created', ?, ?, '', 0, 0)
        `, [checkpointId, createdAt, now, taskId, now, chunks, userId, groupId]);
      }
      
      // 4. Recursively apply linked workflows
      for (const wl of workflowLinks) {
        if (wl.linked_workflow_id) {
          // Recursively call AddWorkflow for linked workflows
          await WorkflowService.AddWorkflow(
            projectPath, 
            wl.linked_workflow_id, 
            wl.name || workflow.name, 
            wl.entity_type_id || entityTypeId, 
            parentEntityId
          );
        }
      }
      
      await persistDatabase(projectName);
    } catch (error) {
      console.error('AddWorkflow error:', error);
      throw error;
    }
  },
};
