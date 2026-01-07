import { studioApiCall, getActiveStudioUrl } from './http-client.js';
import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
}

/**
 * Convert database row to checkpoint object
 */
function rowToCheckpoint(row) {
  if (!row) return null;
  return {
    ...row,
    time_modified: Number(row.time_modified || 0),
    file_size: Number(row.file_size || 0),
    trashed: !!row.trashed,
    synced: !!row.synced,
  };
}

export const CheckpointService = {
  // Returns all checkpoints for a task
  GetCheckpoints: async (projectPath, taskId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const rows = query(db, 'SELECT * FROM task_checkpoint WHERE task_id = ? AND trashed = 0 ORDER BY created_at DESC', [taskId]);
      return rows.map(rowToCheckpoint);
    } catch (error) {
      console.error('GetCheckpoints error:', error);
      return [];
    }
  },

  // Returns a specific checkpoint by ID
  GetCheckpoint: async (projectPath, checkpointId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const row = queryOne(db, 'SELECT * FROM task_checkpoint WHERE id = ?', [checkpointId]);
      return rowToCheckpoint(row) || {};
    } catch (error) {
      console.error('GetCheckpoint error:', error);
      return {};
    }
  },

  // Creates a new checkpoint
  CreateCheckpoint: async (projectPath, checkpoint) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/checkpoint`, 'POST', checkpoint);
    
    try {
      const db = await getDatabase(projectName);
      const now = Date.now();
      execute(db, `
        INSERT INTO task_checkpoint (id, mtime, created_at, task_id, xxhash_checksum, time_modified, file_size, comment, chunks, author_uid, preview_id, group_id, trashed, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1)
      `, [
        result.id, now, result.created_at || new Date().toISOString(), result.task_id,
        result.xxhash_checksum || '', result.time_modified || now, result.file_size || 0,
        result.comment || '', result.chunks || '', result.author_uid || '',
        result.preview_id || '', result.group_id || ''
      ]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('CreateCheckpoint local insert error:', error);
    }
    
    return result;
  },

  // Restores a project to a checkpoint state
  RestoreCheckpoint: async (projectPath, checkpointId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    await studioApiCall(studioUrl, `/${projectName}/checkpoint/${checkpointId}/restore`, 'POST');
  },

  // Deletes a checkpoint
  DeleteCheckpoint: async (projectPath, checkpointId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    await studioApiCall(studioUrl, `/${projectName}/checkpoint/${checkpointId}`, 'DELETE');
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM task_checkpoint WHERE id = ?', [checkpointId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteCheckpoint local update error:', error);
    }
  },

  // Returns the latest checkpoint for a task
  GetLatestCheckpoint: async (projectPath, taskId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const row = queryOne(db, 'SELECT * FROM task_checkpoint WHERE task_id = ? AND trashed = 0 ORDER BY created_at DESC LIMIT 1', [taskId]);
      return rowToCheckpoint(row) || {};
    } catch (error) {
      console.error('GetLatestCheckpoint error:', error);
      return {};
    }
  },
};
