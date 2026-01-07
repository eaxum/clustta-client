export const SyncService = {
  // Pulls data from remote to local project
  PullData: async (projectPath, remoteURL, pullChunk, syncOptions) => {
    console.warn('PullData - full sync not yet implemented in web mode');
  },

  // Synchronizes data between local and remote
  SyncData: async (projectPath, remoteURL, pullChunk, syncOptions) => {
    console.warn('SyncData - full sync not yet implemented in web mode');
  },

  // Clones a project from remote
  CloneProject: async (projectUri, studioName, workingDir, syncOptions) => {
    console.warn('CloneProject not yet implemented in web mode');
  },

  // Pushes checkpoints to remote
  PushCheckpoints: async (projectPath, remoteURL, pullChunk, syncOptions) => {
    console.warn('PushCheckpoints not yet implemented in web mode');
  },

  // Pulls latest checkpoints from remote
  PullLatestCheckpoints: async (projectPath, remoteURL) => {
    console.warn('PullLatestCheckpoints not yet implemented in web mode');
  },

  // Downloads a specific checkpoint
  DownloadCheckpoint: async (projectPath, remoteURL, checkpointId) => {
    console.warn('DownloadCheckpoint not yet implemented in web mode');
  },

  // Checks if project has unsynced changes
  IsUnsynced: async (projectPath) => {
    return false;
  },

  // Cancels ongoing sync operation
  CancelSync: async () => {},
};
