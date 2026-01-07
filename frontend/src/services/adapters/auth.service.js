// =============================================================================
// AUTH SERVICE
// =============================================================================

import { STORAGE_KEYS } from './config.js';
import { globalApiCall } from './http-client.js';
import {
  addAccountToStorage,
  removeAccountFromStorage,
  getMultiAccountToken,
  clearUserSpecificData,
  clearAllUserData,
} from './storage.js';

export const AuthService = {
  Login: async (username, password) => {
    const response = await globalApiCall('/auth/login', 'POST', { email: username, password });
    
    const user = {
      id: response.user?.id || '',
      username: response.user?.username || response.user?.user_name || '',
      email: response.user?.email || '',
      first_name: response.user?.first_name || '',
      last_name: response.user?.last_name || '',
      photo: response.user?.photo || '',
    };
    
    const token = {
      session_id: response.session_id || '',
      user: user,
    };
    
    // Store in legacy keys for backward compatibility
    if (response.session_id) {
      localStorage.setItem(STORAGE_KEYS.SESSION_ID, response.session_id);
    }
    localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user));
    
    // Add to multi-account storage
    addAccountToStorage(token);
    
    return token;
  },
  
  Logout: async () => {
    try {
      await globalApiCall('/auth/logout', 'GET');
    } catch (error) {
      // Ignore logout errors - we'll clear local state anyway
      console.warn('Logout API call failed:', error);
    }
    
    // Get the current active account ID before clearing
    const multiToken = getMultiAccountToken();
    const activeAccountId = multiToken.active_account_id;
    
    // Remove only the current account from multi-account storage
    if (activeAccountId) {
      removeAccountFromStorage(activeAccountId);
    }
    
    // Check if there are remaining accounts
    const updatedMultiToken = getMultiAccountToken();
    const remainingAccounts = Object.keys(updatedMultiToken.accounts);
    
    if (remainingAccounts.length > 0) {
      // Switch to another account automatically
      const nextAccountId = remainingAccounts[0];
      const nextAccount = updatedMultiToken.accounts[nextAccountId];
      
      // Update legacy storage with next account
      localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(nextAccount.user));
      if (nextAccount.session_id) {
        localStorage.setItem(STORAGE_KEYS.SESSION_ID, nextAccount.session_id);
      }
      
      // Clear user-specific cached data since we're switching accounts
      clearUserSpecificData();
      
      return { hasRemainingAccounts: true, nextAccount: nextAccount };
    } else {
      // No more accounts, clear everything including all cached data
      clearAllUserData();
      
      return { hasRemainingAccounts: false, nextAccount: null };
    }
  },
  
  // Logout specific account (for removing from multi-account without logging out current)
  LogoutAccount: async (userId) => {
    const multiToken = getMultiAccountToken();
    
    // If this is the active account, call regular logout
    if (multiToken.active_account_id === userId) {
      return await AuthService.Logout();
    }
    
    // Otherwise just remove from storage
    removeAccountFromStorage(userId);
    return { hasRemainingAccounts: Object.keys(getMultiAccountToken().accounts).length > 0 };
  },
  
  Register: async (firstName, lastName, username, email, password, confirmPassword) => {
    const response = await globalApiCall('/auth/register', 'POST', {
      first_name: firstName,
      last_name: lastName,
      username,
      email,
      password,
      confirm_password: confirmPassword,
    });
    return {
      id: response.id || '',
      username: response.username || response.user_name || '',
      email: response.email || '',
      first_name: response.first_name || '',
      last_name: response.last_name || '',
      photo: '',
    };
  },
  
  IsAuthenticated: async () => {
    try {
      // First check if we have a user stored locally
      const localUser = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || 'null');
      
      // If no local user, not authenticated
      if (!localUser || !localUser.id) {
        return [false, {}];
      }
      
      // Verify with the server that the session is still valid
      const response = await globalApiCall('/auth/authenticated', 'GET');
      
      // If server says authenticated, return the local user data
      return [true, localUser];
    } catch (error) {
      // Server returned error or not authenticated
      return [false, {}];
    }
  },
  
  AuthUser: async () => {
    // First check localStorage for the user
    const localUser = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || 'null');
    
    // If no local user stored, not authenticated
    if (!localUser || !localUser.id) {
      throw new Error('Not authenticated');
    }
    
    // Verify with server that the session is still valid
    try {
      await globalApiCall('/auth/authenticated', 'GET');
      return localUser;
    } catch (error) {
      // Session is invalid - clear local storage
      clearAllUserData();
      throw new Error('Session expired');
    }
  },
  
  VerifyOTP: async (email, token) => {
    await globalApiCall('/auth/verify-otp', 'POST', { email, otp: token });
  },
  
  ResendToken: async (email) => {
    await globalApiCall('/auth/token/resend', 'POST', { email });
  },
  
  ResetPassword: async (email) => {
    await globalApiCall('/auth/reset-password', 'POST', { email });
  },
  
  ChangePassword: async (currentPassword, newPassword, confirmPassword) => {
    await globalApiCall('/auth/change-password', 'POST', {
      password: currentPassword,
      new_password: newPassword,
      confirm_password: confirmPassword,
    });
  },
  
  CheckEmailExists: async (email) => {
    try {
      const response = await globalApiCall(`/auth/email-exists/${encodeURIComponent(email)}`, 'GET');
      return response.exists === true;
    } catch {
      return false;
    }
  },
  
  CheckUsernameExists: async (username) => {
    try {
      const response = await globalApiCall(`/auth/username-exists/${encodeURIComponent(username)}`, 'GET');
      return response.exists === true;
    } catch {
      return false;
    }
  },
  
  UpdateUser: async (firstName, lastName, username, email) => {
    const response = await globalApiCall('/person/update', 'PUT', {
      first_name: firstName,
      last_name: lastName,
      username,
      email,
    });
    const user = {
      id: response.id || '',
      username: response.username || response.user_name || '',
      email: response.email || '',
      first_name: response.first_name || '',
      last_name: response.last_name || '',
      photo: response.photo || '',
    };
    localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user));
    return user;
  },
  
  UpdateUserPhoto: async (photoBase64) => {
    await globalApiCall('/person/photo', 'POST', { photo: photoBase64 });
  },
  
  DeactivateUserAccount: async () => {
    await globalApiCall('/person/deactivate-account', 'POST');
    localStorage.removeItem(STORAGE_KEYS.SESSION_ID);
    localStorage.removeItem(STORAGE_KEYS.USER);
  },
  
  SendInvitationEmail: async (email, studioName, projectName) => {
    await globalApiCall('/auth/send-invitation', 'POST', { email, studio_name: studioName, project_name: projectName });
  },
};
