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
 * Compute collection path from parent path and name
 */
function computeCollectionPath(db, parentId, name) {
  if (!parentId || parentId === '') {
    return '/' + name + '/';
  }
  const parent = queryOne(db, 'SELECT collection_path FROM collection WHERE id = ?', [parentId]);
  if (parent && parent.collection_path) {
    return parent.collection_path + name + '/';
  }
  return '/' + name + '/';
}

/**
 * Convert database row to workflow object, including links, collections, and assets
 */
async function rowToWorkflow(db, row) {
  if (!row) return null;
  
  // Get workflow links
  const links = query(db, 'SELECT * FROM workflow_link WHERE workflow_id = ?', [row.id]);
  
  // Get workflow collections
  const collections = query(db, 'SELECT * FROM workflow_collection WHERE workflow_id = ?', [row.id]);
  
  // Get workflow assets
  const assets = query(db, 'SELECT * FROM workflow_asset WHERE workflow_id = ?', [row.id]);
  
  return {
    ...row,
    synced: !!row.synced,
    links: links || [],
    collections: collections || [],
    assets: assets || [],
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
            INSERT INTO workflow_link (id, mtime, workflow_id, source_asset_type_id, target_asset_type_id, synced)
            VALUES (?, ?, ?, ?, ?, 1)
          `, [link.id, Date.now(), result.id, link.source_asset_type_id, link.target_asset_type_id]);
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
            INSERT INTO workflow_link (id, mtime, workflow_id, source_asset_type_id, target_asset_type_id, synced)
            VALUES (?, ?, ?, ?, ?, 1)
          `, [link.id, Date.now(), workflow.id, link.source_asset_type_id, link.target_asset_type_id]);
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
      execute(db, 'DELETE FROM workflow_collection WHERE workflow_id = ?', [workflowId]);
      execute(db, 'DELETE FROM workflow_asset WHERE workflow_id = ?', [workflowId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteWorkflow local update error:', error);
    }
  },

  // Adds a workflow to an collection - creates collections and assets based on workflow definition
  // This matches the Go implementation which creates actual collections and assets
  AddWorkflow: async (projectPath, workflowId, name, collectionTypeId, parentId) => {
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
      
      // Get workflow collections (child collection templates)
      const workflowCollections = query(db, 'SELECT * FROM workflow_collection WHERE workflow_id = ?', [workflowId]);
      
      // Get workflow assets (asset templates)
      const workflowAssets = query(db, 'SELECT * FROM workflow_asset WHERE workflow_id = ?', [workflowId]);
      
      // Get workflow links (linked workflows)
      const workflowLinks = query(db, 'SELECT * FROM workflow_link WHERE workflow_id = ?', [workflowId]);
      
      // 1. Create the parent collection with the workflow name
      const parentCollectionId = crypto.randomUUID();
      const parentCollectionPath = computeCollectionPath(db, parentId, name);
      execute(db, `
        INSERT INTO collection (id, created_at, mtime, name, description, collection_type_id, parent_id, collection_path, is_library, trashed, synced)
        VALUES (?, ?, ?, ?, '', ?, ?, ?, 0, 0, 0)
      `, [parentCollectionId, createdAt, now, name, collectionTypeId, parentId || '', parentCollectionPath]);
      
      // 2. Create child collections from workflow collection templates
      for (const we of workflowCollections) {
        const childCollectionId = crypto.randomUUID();
        const childCollectionPath = computeCollectionPath(db, parentCollectionId, we.name);
        execute(db, `
          INSERT INTO collection (id, created_at, mtime, name, description, collection_type_id, parent_id, collection_path, is_library, trashed, synced)
          VALUES (?, ?, ?, ?, '', ?, ?, ?, 0, 0, 0)
        `, [childCollectionId, createdAt, now, we.name, we.collection_type_id, parentCollectionId, childCollectionPath]);
      }
      
      // Get the default status (status with short_name = 'todo')
      const defaultStatus = queryOne(db, "SELECT id FROM status WHERE short_name = 'todo' LIMIT 1");
      const defaultStatusId = defaultStatus?.id || '';
      
      // 3. Create assets from workflow asset templates
      for (const wt of workflowAssets) {
        const assetId = crypto.randomUUID();
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
        
        // Create the asset
        execute(db, `
          INSERT INTO asset (id, mtime, created_at, name, description, extension, is_resource, 
                           status_id, asset_type_id, collection_id, assignee_id, assigner_id, 
                           is_link, pointer, preview_id, trashed, synced)
          VALUES (?, ?, ?, ?, '', ?, ?, ?, ?, ?, '', '', ?, ?, '', 0, 0)
        `, [
          assetId, now, createdAt, wt.name, extension, wt.is_resource ? 1 : 0,
          defaultStatusId, wt.asset_type_id, parentCollectionId, wt.is_link ? 1 : 0, wt.pointer || ''
        ]);
        
        // Create initial checkpoint for the asset
        execute(db, `
          INSERT INTO asset_checkpoint (id, created_at, mtime, asset_id, xxhash_checksum, 
                                       time_modified, file_size, chunks, comment, 
                                       author_id, group_id, preview_id, trashed, synced)
          VALUES (?, ?, ?, ?, '', ?, 0, ?, 'Asset created', ?, ?, '', 0, 0)
        `, [checkpointId, createdAt, now, assetId, now, chunks, userId, groupId]);
      }
      
      // 4. Recursively apply linked workflows
      for (const wl of workflowLinks) {
        if (wl.linked_workflow_id) {
          // Recursively call AddWorkflow for linked workflows
          await WorkflowService.AddWorkflow(
            projectPath, 
            wl.linked_workflow_id, 
            wl.name || workflow.name, 
            wl.collection_type_id || collectionTypeId, 
            parentCollectionId
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
