// =============================================================================
// FS SERVICE (Limited in web mode)
// =============================================================================

export const FSService = {
  WatchPath: async (path) => {},
  UnwatchPath: async (path) => {},
  GetFileInfo: async (path) => ({}),
  ReadDirectory: async (path) => [],
  OpenFile: async (path) => {
    console.warn('OpenFile not available in web mode');
  },
  OpenInExplorer: async (path) => {
    console.warn('OpenInExplorer not available in web mode');
  },
  OpenTerminal: async (path) => {
    console.warn('OpenTerminal not available in web mode');
  },
  Exists: async (path) => false,
  IsDirectory: async (path) => false,
  
  // File icon support - returns empty string in web mode
  // The icon store will fall back to fileIconIndex.json
  GetFileIcon: async (ext) => '',
  
  // Path utilities
  TempDir: async () => '/tmp',
  JoinPath: async (...paths) => paths.join('/'),
  WriteFile: async (path, data) => {
    console.warn('WriteFile not available in web mode');
  },
  ReadFile: async (path) => {
    console.warn('ReadFile not available in web mode');
    return '';
  },
  FileHash: async (path) => '',
  FileStat: async (path) => ({}),
  BaseName: async (path) => path.split('/').pop() || path,
};
