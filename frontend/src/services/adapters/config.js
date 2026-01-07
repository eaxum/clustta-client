// In development, use the Vite proxy to bypass CORS
// In production, use the actual API URL
export const isDev = import.meta.env.DEV;
export const GLOBAL_API = isDev ? '/api' : (import.meta.env.VITE_API_URL || 'https://api.clustta.com');
export const CLUSTTA_AGENT = 'Clustta/0.2';

// Storage keys
export const STORAGE_KEYS = {
  SESSION_ID: 'clustta_session_id',
  USER: 'clustta_user',
  ACTIVE_STUDIO: 'clustta_active_studio',
  STUDIO_URL: 'clustta_studio_url',
  SETTINGS: 'clustta_settings',
  ACCOUNTS: 'clustta_accounts',        // Multi-account storage
  ACTIVE_ACCOUNT_ID: 'clustta_active_account_id',
};
