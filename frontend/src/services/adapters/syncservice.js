import { studioDataFetch, getActiveStudioUrl } from './http-client.js';
import { initializeDatabase, getDatabase, persistDatabase, repository } from './project-database.js';
import { STORAGE_KEYS } from './config.js';

// We use fzstd for zstd decompression
import { decompress } from 'fzstd';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  // Handle both URL and file path formats
  const name = projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
  return name;
}

/**
 * Fetch project data from studio server, decompress, decode protobuf, and populate SQLite
 */
async function fetchAndPopulateProjectData(projectName, studioUrl) {
  const user = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || '{}');
  
  console.log('[SyncData] Fetching project data:', { projectName, studioUrl, userId: user.id });
  
  // Fetch compressed protobuf data from /{project}/data
  // The endpoint expects POST with user_id in body
  const compressedData = await studioDataFetch(
    studioUrl,
    `/${projectName}/data`,
    'POST',
    { user_id: user.id }
  );
  
  console.log('[SyncData] Received data, size:', compressedData.byteLength);
  
  // Decompress zstd data
  const compressedBytes = new Uint8Array(compressedData);
  const decompressedBytes = decompress(compressedBytes);
  
  // Decode protobuf
  const projectData = repository.ProjectData.decode(decompressedBytes);
  
  // Convert protobuf objects to plain objects for easier handling
  const plainData = repository.ProjectData.toObject(projectData, {
    longs: Number,
    enums: String,
    bytes: String,
    defaults: true,
    arrays: true,
    objects: true,
  });
  
  // Initialize/populate the SQLite database
  await initializeDatabase(projectName, plainData);
  
  return plainData;
}

export const SyncService = {
  /**
   * Pulls data from remote to local project
   * In web mode, this fetches from studio server and populates browser SQLite
   */
  PullData: async (projectPath, remoteURL, pullChunk, syncOptions) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = remoteURL || getActiveStudioUrl();
    
    try {
      await fetchAndPopulateProjectData(projectName, studioUrl);
      console.log(`PullData completed for ${projectName}`);
    } catch (error) {
      console.error('PullData error:', error);
      throw error;
    }
  },

  /**
   * Synchronizes data between local and remote
   * In web mode: pull only (no push), pullChunk is always false
   */
  SyncData: async (projectPath, remoteURL, pullChunk = false, syncOptions = {}) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = remoteURL || getActiveStudioUrl();
    
    try {
      // In web mode, we only pull data (read-only for now)
      // pullChunk is ignored - we never pull chunks in web mode
      await fetchAndPopulateProjectData(projectName, studioUrl);
      console.log(`SyncData completed for ${projectName}`);
    } catch (error) {
      console.error('SyncData error:', error);
      throw error;
    }
  },

  /**
   * Clones a project from remote
   * In web mode, this is essentially the same as SyncData
   */
  CloneProject: async (projectUri, studioName, workingDir, syncOptions) => {
    const projectName = getProjectName(projectUri);
    const studioUrl = getActiveStudioUrl();
    
    try {
      await fetchAndPopulateProjectData(projectName, studioUrl);
      console.log(`CloneProject completed for ${projectName}`);
    } catch (error) {
      console.error('CloneProject error:', error);
      throw error;
    }
  },

  /**
   * Pushes checkpoints to remote
   * In web mode, this is a no-op (read-only)
   */
  PushCheckpoints: async (projectPath, remoteURL, pullChunk, syncOptions) => {
    console.warn('PushCheckpoints not available in web mode (read-only)');
  },

  /**
   * Pulls latest checkpoints from remote
   * In web mode, this refreshes the local data
   */
  PullLatestCheckpoints: async (projectPath, remoteURL) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = remoteURL || getActiveStudioUrl();
    
    try {
      await fetchAndPopulateProjectData(projectName, studioUrl);
      console.log(`PullLatestCheckpoints completed for ${projectName}`);
    } catch (error) {
      console.error('PullLatestCheckpoints error:', error);
      throw error;
    }
  },

  /**
   * Downloads a specific checkpoint
   * In web mode, this is not supported (no file system)
   */
  DownloadCheckpoint: async (projectPath, remoteURL, checkpointId) => {
    console.warn('DownloadCheckpoint not available in web mode (no file system)');
  },

  /**
   * Checks if project has unsynced changes
   * In web mode, always returns false (read-only)
   */
  IsUnsynced: async (projectPath) => {
    return false;
  },

  /**
   * Cancels ongoing sync operation
   */
  CancelSync: async () => {
    // In web mode, fetch operations can't be easily cancelled
    // This is a no-op for now
  },
};

