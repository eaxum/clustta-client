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
  // Authenticates user with email and password
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

    if (response.session_id) {
      localStorage.setItem(STORAGE_KEYS.SESSION_ID, response.session_id);
    }
    localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user));
    addAccountToStorage(token);

    return token;
  },

  // Logs out the current user and switches to next account if available
  Logout: async () => {
    try {
      await globalApiCall('/auth/logout', 'GET');
    } catch (error) {
      console.warn('Logout API call failed:', error);
    }

    const multiToken = getMultiAccountToken();
    const activeAccountId = multiToken.active_account_id;

    if (activeAccountId) {
      removeAccountFromStorage(activeAccountId);
    }

    const updatedMultiToken = getMultiAccountToken();
    const remainingAccounts = Object.keys(updatedMultiToken.accounts);

    if (remainingAccounts.length > 0) {
      const nextAccountId = remainingAccounts[0];
      const nextAccount = updatedMultiToken.accounts[nextAccountId];

      localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(nextAccount.user));
      if (nextAccount.session_id) {
        localStorage.setItem(STORAGE_KEYS.SESSION_ID, nextAccount.session_id);
      }

      clearUserSpecificData();
      return { hasRemainingAccounts: true, nextAccount: nextAccount };
    } else {
      clearAllUserData();
      return { hasRemainingAccounts: false, nextAccount: null };
    }
  },

  // Logs out a specific account without affecting current session
  LogoutAccount: async (userId) => {
    const multiToken = getMultiAccountToken();

    if (multiToken.active_account_id === userId) {
      return await AuthService.Logout();
    }

    removeAccountFromStorage(userId);
    return { hasRemainingAccounts: Object.keys(getMultiAccountToken().accounts).length > 0 };
  },

  // Registers a new user account
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

  // Checks if current session is authenticated
  IsAuthenticated: async () => {
    try {
      const localUser = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || 'null');

      if (!localUser || !localUser.id) {
        return [false, {}];
      }

      await globalApiCall('/auth/authenticated', 'GET');
      return [true, localUser];
    } catch (error) {
      return [false, {}];
    }
  },

  // Returns the authenticated user or throws if not authenticated
  AuthUser: async () => {
    const localUser = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || 'null');

    if (!localUser || !localUser.id) {
      throw new Error('Not authenticated');
    }

    try {
      await globalApiCall('/auth/authenticated', 'GET');
      return localUser;
    } catch (error) {
      clearAllUserData();
      throw new Error('Session expired');
    }
  },

  // Verifies OTP token for email verification
  VerifyOTP: async (email, token) => {
    await globalApiCall('/auth/verify-otp', 'POST', { email, otp: token });
  },

  // Resends verification token to email
  ResendToken: async (email) => {
    await globalApiCall('/auth/token/resend', 'POST', { email });
  },

  // Sends password reset email
  ResetPassword: async (email) => {
    await globalApiCall('/auth/reset-password', 'POST', { email });
  },

  // Changes password using reset token (from email link)
  ResetChangePassword: async (email, token, newPassword, confirmPassword) => {
    await globalApiCall('/auth/reset-password', 'PUT', {
      email,
      token,
      new_password: newPassword,
      confirm_password: confirmPassword,
    });
  },

  // Changes user password
  ChangePassword: async (currentPassword, newPassword, confirmPassword) => {
    await globalApiCall('/auth/change-password', 'POST', {
      password: currentPassword,
      new_password: newPassword,
      confirm_password: confirmPassword,
    });
  },

  // Checks if email is already registered
  CheckEmailExists: async (email) => {
    try {
      const response = await globalApiCall(`/auth/email-exists/${encodeURIComponent(email)}`, 'GET');
      return response.exists === true;
    } catch {
      return false;
    }
  },

  // Checks if username is already taken
  CheckUsernameExists: async (username) => {
    try {
      const response = await globalApiCall(`/auth/username-exists/${encodeURIComponent(username)}`, 'GET');
      return response.exists === true;
    } catch {
      return false;
    }
  },

  // Updates user profile information
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

  // Updates user profile photo
  UpdateUserPhoto: async (photoBase64) => {
    await globalApiCall('/person/photo', 'POST', { photo: photoBase64 });
  },

  // Deactivates the user account
  DeactivateUserAccount: async () => {
    await globalApiCall('/person/deactivate-account', 'POST');
    localStorage.removeItem(STORAGE_KEYS.SESSION_ID);
    localStorage.removeItem(STORAGE_KEYS.USER);
  },

  // Sends invitation email to collaborate on a project
  SendInvitationEmail: async (email, studioName, projectName) => {
    await globalApiCall('/auth/send-invitation', 'POST', { email, studio_name: studioName, project_name: projectName });
  },
};
