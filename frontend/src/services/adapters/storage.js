import { STORAGE_KEYS } from './config.js';

// Returns all settings from localStorage
export function getSettings() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEYS.SETTINGS) || '{}');
  } catch {
    return {};
  }
}

// Sets a single setting value
export function setSetting(key, value) {
  const settings = getSettings();
  settings[key] = value;
  localStorage.setItem(STORAGE_KEYS.SETTINGS, JSON.stringify(settings));
}

// Returns a setting value with optional default
export function getSetting(key, defaultValue) {
  const settings = getSettings();
  return settings[key] !== undefined ? settings[key] : defaultValue;
}

// Returns the multi-account token structure from localStorage
export function getMultiAccountToken() {
  try {
    const data = localStorage.getItem(STORAGE_KEYS.ACCOUNTS);
    if (!data) {
      return {
        active_account_id: '',
        accounts: {},
      };
    }
    return JSON.parse(data);
  } catch {
    return {
      active_account_id: '',
      accounts: {},
    };
  }
}

// Saves the multi-account token structure to localStorage
export function setMultiAccountToken(multiToken) {
  localStorage.setItem(STORAGE_KEYS.ACCOUNTS, JSON.stringify(multiToken));
  
  // Also update the legacy USER key for backward compatibility
  if (multiToken.active_account_id && multiToken.accounts[multiToken.active_account_id]) {
    const activeAccount = multiToken.accounts[multiToken.active_account_id];
    localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(activeAccount.user));
    if (activeAccount.session_id) {
      localStorage.setItem(STORAGE_KEYS.SESSION_ID, activeAccount.session_id);
    }
  }
}

// Adds an account to the multi-account structure
export function addAccountToStorage(token) {
  const multiToken = getMultiAccountToken();
  
  multiToken.accounts[token.user.id] = token;
  
  if (!multiToken.active_account_id || Object.keys(multiToken.accounts).length === 1) {
    multiToken.active_account_id = token.user.id;
  }
  
  setMultiAccountToken(multiToken);
}

// Removes an account from storage
export function removeAccountFromStorage(userId) {
  const multiToken = getMultiAccountToken();
  
  delete multiToken.accounts[userId];
  
  if (multiToken.active_account_id === userId) {
    const remainingIds = Object.keys(multiToken.accounts);
    multiToken.active_account_id = remainingIds.length > 0 ? remainingIds[0] : '';
  }
  
  setMultiAccountToken(multiToken);
  
  if (Object.keys(multiToken.accounts).length === 0) {
    localStorage.removeItem(STORAGE_KEYS.USER);
    localStorage.removeItem(STORAGE_KEYS.SESSION_ID);
  }
}

// Switches the active account
export function switchActiveAccount(userId) {
  const multiToken = getMultiAccountToken();
  
  if (!multiToken.accounts[userId]) {
    throw new Error(`Account with id ${userId} not found`);
  }
  
  multiToken.active_account_id = userId;
  setMultiAccountToken(multiToken);
  
  return multiToken.accounts[userId];
}

// Clears user-specific cached data when switching accounts
export function clearUserSpecificData() {
  const settings = getSettings();
  
  const preservedSettings = {
    theme: settings.theme,
    language: settings.language,
    iconScheme: settings.iconScheme,
    useGrid: settings.useGrid,
    projectGridView: settings.projectGridView,
    compactView: settings.compactView,
    showHiddenFiles: settings.showHiddenFiles,
    showUntrackedProjects: settings.showUntrackedProjects,
    eulaAccepted: settings.eulaAccepted,
  };
  
  localStorage.setItem(STORAGE_KEYS.SETTINGS, JSON.stringify(preservedSettings));
  
  localStorage.removeItem(STORAGE_KEYS.STUDIO_URL);
  localStorage.removeItem(STORAGE_KEYS.ACTIVE_STUDIO);
}

// Clears all user data when fully logging out
export function clearAllUserData() {
  const settings = getSettings();
  
  const preservedSettings = {
    theme: settings.theme,
    iconScheme: settings.iconScheme,
    eulaAccepted: settings.eulaAccepted,
  };
  
  const keysToRemove = [];
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (key && key.startsWith('clustta_')) {
      keysToRemove.push(key);
    }
  }
  keysToRemove.forEach(key => localStorage.removeItem(key));
}

// Migrates from old single-account storage to multi-account
export function migrateToMultiAccount() {
  const multiToken = getMultiAccountToken();
  
  if (Object.keys(multiToken.accounts).length > 0) {
    return;
  }
  
  const oldUser = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || 'null');
  const oldSessionId = localStorage.getItem(STORAGE_KEYS.SESSION_ID);
  
  if (oldUser && oldUser.id) {
    const token = {
      user: oldUser,
      session_id: oldSessionId || '',
    };
    
    addAccountToStorage(token);
    console.log('[Migration] Migrated single account to multi-account storage');
  }
}

// Run migration on module load
migrateToMultiAccount();
