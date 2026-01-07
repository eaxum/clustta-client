// =============================================================================
// LOG SERVICE
// =============================================================================

export const LogService = {
  LogInfo: async (message) => console.log('[INFO]', message),
  LogError: async (message) => console.error('[ERROR]', message),
  LogWarning: async (message) => console.warn('[WARN]', message),
  LogDebug: async (message) => console.debug('[DEBUG]', message),
};
