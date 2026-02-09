export const LogService = {
  // Logs an info message
  LogInfo: async (message) => console.log('[INFO]', message),

  // Logs an error message
  LogError: async (message) => console.error('[ERROR]', message),

  // Logs a warning message
  LogWarning: async (message) => console.warn('[WARN]', message),

  // Logs a debug message
  LogDebug: async (message) => console.debug('[DEBUG]', message),
};
