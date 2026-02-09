import { STORAGE_KEYS } from './config.js';
import { globalApiCall } from './http-client.js';
import {
  getMultiAccountToken,
  switchActiveAccount,
  removeAccountFromStorage,
  addAccountToStorage,
} from './storage.js';

export const AccountService = {
  // Returns all studios for the current user
  GetStudios: async () => {
    const studios = await globalApiCall('/person/studios', 'GET');
    return Array.isArray(studios) ? studios : [];
  },

  // Returns all studios including those the user has access to
  GetAllStudios: async () => {
    const studios = await globalApiCall('/person/all-studios', 'GET');
    return Array.isArray(studios) ? studios : [];
  },

  // Returns all stored user accounts
  GetAllAccounts: async () => {
    const multiToken = getMultiAccountToken();
    return multiToken.accounts;
  },

  // Returns the currently active account
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

  // Returns the number of stored accounts
  GetAccountCount: async () => {
    const multiToken = getMultiAccountToken();
    return Object.keys(multiToken.accounts).length;
  },

  // Changes the active account to the specified user
  SwitchAccount: async (userId) => {
    const token = switchActiveAccount(userId);

    // Verify session is still valid with server
    try {
      await globalApiCall('/auth/authenticated', 'GET');
    } catch (error) {
      console.warn('Session expired for switched account, re-authentication may be required');
    }

    return token;
  },

  // Removes an account from storage
  RemoveAccount: async (userId) => {
    removeAccountFromStorage(userId);
  },

  // Adds a new account to storage
  AddAccount: async (token) => {
    const accountToken = {
      user: token.user || token,
      session_id: token.session_id || localStorage.getItem(STORAGE_KEYS.SESSION_ID) || '',
    };

    addAccountToStorage(accountToken);
  },
};