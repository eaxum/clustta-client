import { studioApiCall, studioDataFetch, getActiveStudioUrl } from './http-client.js';
import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';
import { decompress } from 'fzstd';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

/**
 * Convert bytes to human-readable format (like Go's BytesToHumanReadable)
 * @param {number} bytes - The number of bytes
 * @returns {string} - Human readable string (e.g., "1.5 MB")
 */
function bytesToHumanReadable(bytes) {
  const KB = 1024;
  const MB = KB * 1024;
  const GB = MB * 1024;
  
  if (bytes >= GB) {
    return `${(bytes / GB).toFixed(2)} GB`;
  } else if (bytes >= MB) {
    return `${(bytes / MB).toFixed(2)} MB`;
  } else if (bytes >= KB) {
    return `${(bytes / KB).toFixed(2)} KB`;
  } else {
    return `${bytes} B`;
  }
}

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

  // Returns the project timeline showing checkpoint history
  GetTimeline: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      
      // Get all checkpoints with task info, grouped by date
      const checkpointRows = query(db, `
        SELECT tc.*, t.name as task_name, t.entity_id
        FROM task_checkpoint tc
        LEFT JOIN task t ON tc.task_id = t.id
        WHERE tc.trashed = 0
        ORDER BY tc.created_at DESC
      `);
      
      // Build task type map
      const taskTypeRows = query(db, 'SELECT * FROM task_type');
      const taskTypeMap = {};
      for (const tt of taskTypeRows) {
        taskTypeMap[tt.id] = tt;
      }
      
      // Build entity map for path info
      const entityRows = query(db, 'SELECT * FROM entity');
      const entityMap = {};
      for (const e of entityRows) {
        entityMap[e.id] = e;
      }
      
      // Transform checkpoints to timeline format
      const timeline = checkpointRows.map(row => {
        const entity = entityMap[row.entity_id] || {};
        return {
          id: row.id,
          task_id: row.task_id,
          task_name: row.task_name || '',
          entity_id: row.entity_id || '',
          entity_name: entity.name || '',
          entity_path: entity.entity_path || '',
          created_at: row.created_at,
          time_modified: Number(row.time_modified || 0),
          file_size: Number(row.file_size || 0),
          comment: row.comment || '',
          author_uid: row.author_uid || '',
          preview_id: row.preview_id || '',
          group_id: row.group_id || '',
          xxhash_checksum: row.xxhash_checksum || '',
        };
      });
      
      return timeline;
    } catch (error) {
      console.error('GetTimeline error:', error);
      return [];
    }
  },

  /**
   * Downloads an asset file by fetching chunks, reassembling, decompressing, and triggering browser download.
   * @param {string} projectPath - The project path/uri
   * @param {string} taskId - The task ID to download
   * @param {string} [checkpointId] - Optional specific checkpoint ID (defaults to latest)
   * @param {function} [onProgress] - Optional callback for progress updates (current, total)
   * @returns {Promise<void>}
   */
  DownloadAsset: async (projectPath, taskId, checkpointId = null, onProgress = null) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    try {
      const db = await getDatabase(projectName);
      const task = queryOne(db, 'SELECT * FROM task WHERE id = ?', [taskId]);
      if (!task) {
        throw new Error('Task not found');
      }
      
      // Get checkpoint (either specified or latest)
      let checkpoint;
      if (checkpointId) {
        checkpoint = queryOne(db, 'SELECT * FROM task_checkpoint WHERE id = ?', [checkpointId]);
      } else {
        checkpoint = queryOne(db, 'SELECT * FROM task_checkpoint WHERE task_id = ? AND trashed = 0 ORDER BY created_at DESC LIMIT 1', [taskId]);
      }
      
      if (!checkpoint || !checkpoint.chunks) {
        throw new Error('No checkpoint found for this task');
      }
      
      // Parse chunk hashes from comma-separated string
      const chunkHashes = checkpoint.chunks.split(',').filter(h => h.length > 0);
      if (chunkHashes.length === 0) {
        throw new Error('Checkpoint has no chunks');
      }
      
      const filename = `${task.name}${task.extension}`;
      const displayName = `${utils.capitalizeStr(task.name)}${task.extension}`;
      const totalChunks = chunkHashes.length;
      const totalSize = checkpoint.file_size || 0;  // Get from checkpoint record
      
      const downloadProgressCallback = (receivedBytes, contentLength) => {
        // Use checkpoint.file_size as total if Content-Length not available
        const total = contentLength > 0 ? contentLength : totalSize;
        const percentage = total > 0 ? Math.round((receivedBytes / total) * 95) : 0;
        const currentSize = bytesToHumanReadable(receivedBytes);
        const totalSizeStr = total > 0 ? bytesToHumanReadable(total) : '?';
        
        emitter.emit('progress-update', {
          current: percentage,
          total: 100,
          percentage,
          title: 'Downloading',
          message: `Receiving ${currentSize}/${totalSizeStr}`,
          extra_message: displayName,
          operation_type: 'read'
        });
      };
      
      // Fetch chunks from studio server using stream-chunks endpoint (GET with body)
      const chunksData = await studioDataFetch(
        studioUrl,
        `/${projectName}/stream-chunks`,
        'GET',
        { chunks: chunkHashes },
        {},
        downloadProgressCallback
      );
      
      const totalBytes = chunksData.byteLength;
      
      // Decode TLV-encoded chunks and reassemble (no separate progress needed, already shown during download)
      const reassembledData = decodeTLVChunks(new Uint8Array(chunksData), totalChunks, null);
      const decompressedData = decompress(reassembledData);
      
      const blob = new Blob([decompressedData], { type: 'application/octet-stream' });
      triggerBrowserDownload(blob, filename);
      
      // Emit 100% progress to close the modal
      emitter.emit('progress-update', {
        current: 100,
        total: 100,
        percentage: 100,
        title: 'Downloading',
        message: `Downloaded ${bytesToHumanReadable(decompressedData.length)}`,
        extra_message: displayName,
        operation_type: 'read'
      });
      
      // Show success notification
      emitter.emit('add_message', {
        message: 'Download Complete',
        longMessage: `${displayName} (${bytesToHumanReadable(decompressedData.length)}) downloaded successfully`,
        type: 'success',
        hasUndo: false,
        read: false,
      });
      
    } catch (error) {
      console.error('[DownloadAsset] Error:', error);
      throw error;
    }
  },

  /**
   * Downloads a specific checkpoint file
   * @param {string} projectPath - The project path/uri
   * @param {string} checkpointId - The checkpoint ID to download
   * @param {function} [onProgress] - Optional callback for progress updates
   * @returns {Promise<void>}
   */
  DownloadCheckpoint: async (projectPath, checkpointId, onProgress = null) => {
    const projectName = getProjectName(projectPath);
    const db = await getDatabase(projectName);
    
    // Get checkpoint to find the associated task
    const checkpoint = queryOne(db, 'SELECT * FROM task_checkpoint WHERE id = ?', [checkpointId]);
    if (!checkpoint) {
      throw new Error('Checkpoint not found');
    }
    
    // Use DownloadAsset with the specific checkpoint
    return CheckpointService.DownloadAsset(projectPath, checkpoint.task_id, checkpointId, onProgress);
  },

};

/**
 * Decode TLV (Tag-Length-Value) encoded chunks from server response.
 * Format: [32 bytes hash][4 bytes length (big-endian)][data...]
 * @param {Uint8Array} data - The raw TLV-encoded data
 * @param {number} expectedCount - Expected number of chunks
 * @param {function} [onProgress] - Optional progress callback (processedCount, expectedCount, processedBytes, totalBytes)
 * @returns {Uint8Array} - Concatenated chunk data (still compressed)
 */
function decodeTLVChunks(data, expectedCount, onProgress = null) {
  const chunks = [];
  let offset = 0;
  let processedCount = 0;
  let processedBytes = 0;
  const totalBytes = data.length;
  
  while (offset < data.length) {
    // Read 32-byte hash (tag)
    if (offset + 32 > data.length) {
      console.warn('[decodeTLVChunks] Incomplete hash at end of data');
      break;
    }
    const hashBytes = data.slice(offset, offset + 32);
    offset += 32;
    processedBytes += 32;
    
    // Read 4-byte length (big-endian)
    if (offset + 4 > data.length) {
      console.warn('[decodeTLVChunks] Incomplete length at end of data');
      break;
    }
    const length = (data[offset] << 24) | (data[offset + 1] << 16) | (data[offset + 2] << 8) | data[offset + 3];
    offset += 4;
    processedBytes += 4;
    
    // Read chunk data
    if (offset + length > data.length) {
      console.warn(`[decodeTLVChunks] Incomplete chunk data: expected ${length} bytes, have ${data.length - offset}`);
      break;
    }
    const chunkData = data.slice(offset, offset + length);
    offset += length;
    processedBytes += length;
    
    chunks.push(chunkData);
    processedCount++;
    
    if (onProgress) {
      onProgress(processedCount, expectedCount, processedBytes, totalBytes);
    }
  }
  
  // Concatenate all chunks
  const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
  const result = new Uint8Array(totalLength);
  let writeOffset = 0;
  for (const chunk of chunks) {
    result.set(chunk, writeOffset);
    writeOffset += chunk.length;
  }
  
  return result;
}

/**
 * Trigger a browser file download from a Blob
 * @param {Blob} blob - The file data
 * @param {string} filename - The filename to save as
 */
function triggerBrowserDownload(blob, filename) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
