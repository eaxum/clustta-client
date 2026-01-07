// =============================================================================
// ACCOUNT SERVICE
// =============================================================================

import { STORAGE_KEYS } from './config.js';
import { globalApiCall } from './http-client.js';
import {
  getMultiAccountToken,
  switchActiveAccount,
  removeAccountFromStorage,
  addAccountToStorage,
} from './storage.js';

export const AccountService = {
  GetStudios: async () => {
    const studios = await globalApiCall('/person/studios', 'GET');
    return Array.isArray(studios) ? studios : [];
  },
  
  GetAllStudios: async () => {
    const studios = await globalApiCall('/person/all-studios', 'GET');
    return Array.isArray(studios) ? studios : [];
  },
  
  // Multi-account support
  GetAllAccounts: async () => {
    const multiToken = getMultiAccountToken();
    return multiToken.accounts;
  },
  
  GetActiveAccount: async () => {
    const multiToken = getMultiAccountToken();
    
    if (!multiToken.active_account_id) {
      throw new Error('No active account set');
    }
    
    const token = multiToken.accounts[multiToken.active_account_id];
    if (!token) {
      throw new Error('Active account not found in accounts list');
    }
    
    return token;
  },
  
  GetAccountCount: async () => {
    const multiToken = getMultiAccountToken();
    return Object.keys(multiToken.accounts).length;
  },
  
  SwitchAccount: async (userId) => {
    // Switch in local storage
    const token = switchActiveAccount(userId);
    
    // For web mode, switching accounts requires re-establishing the session
    // The stored session_id might be expired, so we need to verify with server
    try {
      // Try to verify the session is still valid by calling authenticated endpoint
      await globalApiCall('/auth/authenticated', 'GET');
    } catch (error) {
      // Session expired - user needs to re-login
      // We keep the account in storage but mark that re-auth is needed
      console.warn('Session expired for switched account, re-authentication may be required');
    }
    
    return token;
  },
  
  RemoveAccount: async (userId) => {
    removeAccountFromStorage(userId);
  },
  
  AddAccount: async (token) => {
    // Ensure token has the expected structure
    const accountToken = {
      user: token.user || token,
      session_id: token.session_id || localStorage.getItem(STORAGE_KEYS.SESSION_ID) || '',
    };
    
    addAccountToStorage(accountToken);
  },
};
