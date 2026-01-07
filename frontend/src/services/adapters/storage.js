// =============================================================================
// STORAGE UTILITIES
// =============================================================================
// localStorage helpers, multi-account management, and migration logic

import { STORAGE_KEYS } from './config.js';

// =============================================================================
// SETTINGS HELPERS
// =============================================================================

export function getSettings() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEYS.SETTINGS) || '{}');
  } catch {
    return {};
  }
}

export function setSetting(key, value) {
  const settings = getSettings();
  settings[key] = value;
  localStorage.setItem(STORAGE_KEYS.SETTINGS, JSON.stringify(settings));
}

export function getSetting(key, defaultValue) {
  const settings = getSettings();
  return settings[key] !== undefined ? settings[key] : defaultValue;
}

// =============================================================================
// MULTI-ACCOUNT HELPERS
// =============================================================================

/**
 * Get the multi-account token structure from localStorage
 * Mirrors Go's MultiAccountToken structure
 */
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

/**
 * Save the multi-account token structure to localStorage
 */
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

/**
 * Add an account to the multi-account structure
 */
export function addAccountToStorage(token) {
  const multiToken = getMultiAccountToken();
  
  // Add the new account
  multiToken.accounts[token.user.id] = token;
  
  // If this is the first account or no active account is set, make it active
  if (!multiToken.active_account_id || Object.keys(multiToken.accounts).length === 1) {
    multiToken.active_account_id = token.user.id;
  }
  
  setMultiAccountToken(multiToken);
}

/**
 * Remove an account from storage
 */
export function removeAccountFromStorage(userId) {
  const multiToken = getMultiAccountToken();
  
  // Remove the account
  delete multiToken.accounts[userId];
  
  // If we removed the active account, set a new active account
  if (multiToken.active_account_id === userId) {
    const remainingIds = Object.keys(multiToken.accounts);
    multiToken.active_account_id = remainingIds.length > 0 ? remainingIds[0] : '';
  }
  
  setMultiAccountToken(multiToken);
  
  // If no accounts left, clear legacy storage too
  if (Object.keys(multiToken.accounts).length === 0) {
    localStorage.removeItem(STORAGE_KEYS.USER);
    localStorage.removeItem(STORAGE_KEYS.SESSION_ID);
  }
}

/**
 * Switch the active account
 */
export function switchActiveAccount(userId) {
  const multiToken = getMultiAccountToken();
  
  if (!multiToken.accounts[userId]) {
    throw new Error(`Account with id ${userId} not found`);
  }
  
  multiToken.active_account_id = userId;
  setMultiAccountToken(multiToken);
  
  return multiToken.accounts[userId];
}

/**
 * Clear user-specific cached data (when switching accounts)
 * Keeps some settings but clears account-specific caches
 */
export function clearUserSpecificData() {
  // Get current settings to preserve non-user-specific ones
  const settings = getSettings();
  
  // Keep only UI preferences, clear everything else
  const preservedSettings = {
    theme: settings.theme,
    iconScheme: settings.iconScheme,
    useGrid: settings.useGrid,
    projectGridView: settings.projectGridView,
    compactView: settings.compactView,
    showHiddenFiles: settings.showHiddenFiles,
    showUntrackedProjects: settings.showUntrackedProjects,
    eulaAccepted: settings.eulaAccepted,
  };
  
  localStorage.setItem(STORAGE_KEYS.SETTINGS, JSON.stringify(preservedSettings));
  
  // Clear studio URL
  localStorage.removeItem(STORAGE_KEYS.STUDIO_URL);
  localStorage.removeItem(STORAGE_KEYS.ACTIVE_STUDIO);
}

/**
 * Clear ALL user data (when fully logging out)
 */
export function clearAllUserData() {
  // Get current settings to preserve only non-user settings BEFORE clearing
  const settings = getSettings();
  
  // Keep only app-level preferences (theme, icon scheme, EULA)
  const preservedSettings = {
    theme: settings.theme,
    iconScheme: settings.iconScheme,
    eulaAccepted: settings.eulaAccepted,
  };
  
  // Clear ALL clustta-prefixed keys from localStorage
  const keysToRemove = [];
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (key && key.startsWith('clustta_')) {
      keysToRemove.push(key);
    }
  }
  keysToRemove.forEach(key => localStorage.removeItem(key));
  
  // Only restore settings if we have something to preserve
  // if (preservedSettings.theme || preservedSettings.iconScheme || preservedSettings.eulaAccepted) {
  //   localStorage.setItem(STORAGE_KEYS.SETTINGS, JSON.stringify(preservedSettings));
  // }
}

// =============================================================================
// MIGRATION
// =============================================================================

/**
 * Migrate from old single-account storage to multi-account
 * Call this on app startup to ensure backward compatibility
 */
export function migrateToMultiAccount() {
  const multiToken = getMultiAccountToken();
  
  // If we already have accounts in multi-account storage, no migration needed
  if (Object.keys(multiToken.accounts).length > 0) {
    return;
  }
  
  // Check if we have old single-account data
  const oldUser = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || 'null');
  const oldSessionId = localStorage.getItem(STORAGE_KEYS.SESSION_ID);
  
  if (oldUser && oldUser.id) {
    // Migrate to multi-account structure
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
