export const FSService = {
  // Starts watching a path for changes
  WatchPath: async (path) => {},

  // Stops watching a path
  UnwatchPath: async (path) => {},

  // Returns file metadata
  GetFileInfo: async (path) => ({}),

  // Returns directory contents
  ReadDirectory: async (path) => [],

  // Opens a file in the default application
  OpenFile: async (path) => {
    console.warn('OpenFile not available in web mode');
  },

  // Opens path in system file explorer
  OpenInExplorer: async (path) => {
    console.warn('OpenInExplorer not available in web mode');
  },

  // Opens terminal at path
  OpenTerminal: async (path) => {
    console.warn('OpenTerminal not available in web mode');
  },

  // Checks if path exists
  // In web mode, all project database paths (.clst files) are considered to exist
  // since we access them through the API/SQLite adapter
  Exists: async (path) => {
    // Project database files (.clst) are always available in web mode
    if (path && path.endsWith('.clst')) {
      return true;
    }
    return false;
  },

  // Checks if path is a directory
  IsDirectory: async (path) => false,

  // Returns system icon for file extension
  GetFileIcon: async (ext) => '',

  // Returns system temporary directory
  TempDir: async () => '/tmp',

  // Joins path segments
  JoinPath: async (...paths) => paths.join('/'),

  // Writes data to file
  WriteFile: async (path, data) => {
    console.warn('WriteFile not available in web mode');
  },

  // Reads file contents
  ReadFile: async (path) => {
    console.warn('ReadFile not available in web mode');
    return '';
  },

  // Returns hash of file contents
  FileHash: async (path) => '',

  // Returns file statistics
  FileStat: async (path) => ({}),

  // Returns base name of path
  BaseName: async (path) => path.split('/').pop() || path,
};
