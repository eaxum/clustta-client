// =============================================================================
// APP SERVICE
// =============================================================================

export const AppService = {
  GetOS: async () => 'web',
  GetVersion: async () => '0.2.0-web',
  OpenURL: async (url) => window.open(url, '_blank'),
  GetAppDataDir: async () => '/web/data',
};
