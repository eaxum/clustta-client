// =============================================================================
// SYNC SERVICE
// =============================================================================

export const SyncService = {
  PullData: async (projectPath, remoteURL, pullChunk, syncOptions) => {
    // TODO: Implement full sync with protobuf support
    console.warn('PullData - full sync not yet implemented in web mode');
  },
  
  SyncData: async (projectPath, remoteURL, pullChunk, syncOptions) => {
    // TODO: Implement full sync
    console.warn('SyncData - full sync not yet implemented in web mode');
  },
  
  CloneProject: async (projectUri, studioName, workingDir, syncOptions) => {
    // TODO: Implement clone
    console.warn('CloneProject not yet implemented in web mode');
  },
  
  PushCheckpoints: async (projectPath, remoteURL, pullChunk, syncOptions) => {
    // TODO: Implement checkpoint push
    console.warn('PushCheckpoints not yet implemented in web mode');
  },
  
  PullLatestCheckpoints: async (projectPath, remoteURL) => {
    // TODO: Implement checkpoint pull
    console.warn('PullLatestCheckpoints not yet implemented in web mode');
  },
  
  DownloadCheckpoint: async (projectPath, remoteURL, checkpointId) => {
    // TODO: Implement checkpoint download
    console.warn('DownloadCheckpoint not yet implemented in web mode');
  },
  
  IsUnsynced: async (projectPath) => {
    // TODO: Compare local vs remote sync tokens
    return false;
  },
  
  CancelSync: async () => {
    // Nothing to cancel in web mode currently
  },
};
