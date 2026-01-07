export const AppService = {
  // Returns the current operating system
  GetOS: async () => 'web',

  // Returns the application version
  GetVersion: async () => '0.2.0-web',

  // Opens a URL in a new browser tab
  OpenURL: async (url) => window.open(url, '_blank'),

  // Returns the application data directory
  GetAppDataDir: async () => '/web/data',
};
