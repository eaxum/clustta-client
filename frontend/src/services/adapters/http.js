// HTTP adapter - REST API implementations for web mode
// This file provides HTTP-based implementations for all services
// that mirror the Wails bindings interface

// =============================================================================
// CONFIGURATION
// =============================================================================

// In development, use the Vite proxy to bypass CORS
// In production, use the actual API URL
const isDev = import.meta.env.DEV;
const GLOBAL_API = isDev ? '/api' : (import.meta.env.VITE_API_URL || 'https://api.clustta.com');
const CLUSTTA_AGENT = 'Clustta/0.2';

// Storage keys
const STORAGE_KEYS = {
  SESSION_ID: 'clustta_session_id',
  USER: 'clustta_user',
  ACTIVE_STUDIO: 'clustta_active_studio',
  STUDIO_URL: 'clustta_studio_url',
  SETTINGS: 'clustta_settings',
  ACCOUNTS: 'clustta_accounts',        // Multi-account storage
  ACTIVE_ACCOUNT_ID: 'clustta_active_account_id',
};

// =============================================================================
// MULTI-ACCOUNT HELPERS
// =============================================================================

/**
 * Get the multi-account token structure from localStorage
 * Mirrors Go's MultiAccountToken structure
 */
function getMultiAccountToken() {
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
function setMultiAccountToken(multiToken) {
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
function addAccountToStorage(token) {
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
function removeAccountFromStorage(userId) {
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
function switchActiveAccount(userId) {
  const multiToken = getMultiAccountToken();
  
  if (!multiToken.accounts[userId]) {
    throw new Error(`Account with id ${userId} not found`);
  }
  
  multiToken.active_account_id = userId;
  setMultiAccountToken(multiToken);
  
  return multiToken.accounts[userId];
}

/**
 * Migrate from old single-account storage to multi-account
 * Call this on app startup to ensure backward compatibility
 */
function migrateToMultiAccount() {
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

/**
 * Clear user-specific cached data (when switching accounts)
 * Keeps some settings but clears account-specific caches
 */
function clearUserSpecificData() {
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
function clearAllUserData() {
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
// HTTP CLIENT
// =============================================================================

/**
 * Makes an API call to the global Clustta server
 */
async function globalApiCall(endpoint, method = 'GET', body = null, options = {}) {
  // Note: We use credentials: 'include' to let the browser handle cookies automatically
  // Do NOT manually set Cookie header - browsers forbid it and it won't work
  
  const headers = {
    'Content-Type': 'application/json',
    'Clustta-Agent': CLUSTTA_AGENT,
    ...options.headers,
  };

  const response = await fetch(`${GLOBAL_API}${endpoint}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : null,
    credentials: 'include',  // This tells browser to send and receive cookies
  });

  if (!response.ok) {
    let errorMessage;
    try {
      const errorData = await response.json();
      errorMessage = errorData.message || errorData.error || response.statusText;
    } catch {
      errorMessage = await response.text() || response.statusText;
    }
    throw new Error(errorMessage);
  }

  // Handle empty responses
  const contentType = response.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    return response.json();
  }
  
  // Try to parse as JSON even if content-type is not set correctly
  const text = await response.text();
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

/**
 * Makes an API call to a studio server
 */
async function studioApiCall(studioUrl, endpoint, method = 'GET', body = null, options = {}) {
  const user = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || '{}');
  
  const headers = {
    'Content-Type': 'application/json',
    'Clustta-Agent': CLUSTTA_AGENT,
    'UserId': user.id || '',
    'UserData': JSON.stringify(user),
    ...options.headers,
  };

  const response = await fetch(`${studioUrl}${endpoint}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : null,
  });

  if (!response.ok) {
    let errorMessage;
    try {
      const errorData = await response.json();
      errorMessage = errorData.message || errorData.error || response.statusText;
    } catch {
      errorMessage = await response.text() || response.statusText;
    }
    throw new Error(errorMessage);
  }

  const contentType = response.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    return response.json();
  }
  return response.text();
}

/**
 * Get the active studio URL from storage or settings
 */
function getActiveStudioUrl() {
  return localStorage.getItem(STORAGE_KEYS.STUDIO_URL) || '';
}

/**
 * Set the active studio URL
 */
function setActiveStudioUrl(url) {
  localStorage.setItem(STORAGE_KEYS.STUDIO_URL, url);
}

// =============================================================================
// SETTINGS SERVICE (localStorage-based)
// =============================================================================

function getSettings() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEYS.SETTINGS) || '{}');
  } catch {
    return {};
  }
}

function setSetting(key, value) {
  const settings = getSettings();
  settings[key] = value;
  localStorage.setItem(STORAGE_KEYS.SETTINGS, JSON.stringify(settings));
}

function getSetting(key, defaultValue) {
  const settings = getSettings();
  return settings[key] !== undefined ? settings[key] : defaultValue;
}

export const SettingsService = {
  // View settings
  GetUseAltUrl: async () => getSetting('useAltUrl', false),
  SetUseAltUrl: async (value) => setSetting('useAltUrl', value),
  IsProjectGridView: async () => getSetting('projectGridView', true),
  SetProjectGridView: async (value) => setSetting('projectGridView', value),
  GetUseGrid: async () => getSetting('useGrid', true),
  SetUseGrid: async (value) => setSetting('useGrid', value),
  IsShowUntrackedProjects: async () => getSetting('showUntrackedProjects', true),
  SetShowUntrackedProjects: async (value) => setSetting('showUntrackedProjects', value),
  IsCompactView: async () => getSetting('compactView', false),
  SetCompactView: async (value) => setSetting('compactView', value),
  IsShowHiddenFiles: async () => getSetting('showHiddenFiles', false),
  SetShowHiddenFiles: async (value) => setSetting('showHiddenFiles', value),
  
  // Theme
  GetTheme: async () => getSetting('theme', 'dark'),
  SetTheme: async (theme) => setSetting('theme', theme),
  GetIconScheme: async () => getSetting('iconScheme', 'solid'),
  SetIconScheme: async (scheme) => setSetting('iconScheme', scheme),
  
  // Studio/Project
  GetLastStudio: async () => getSetting('lastStudio', ''),
  SetLastStudio: async (studio) => setSetting('lastStudio', studio),
  
  // GetStudios - mirrors Go behavior: check localStorage first, then fetch from API
  GetStudios: async (path) => {
    // Always fetch fresh data from API (like the Go version does)
    try {
      const userStudios = await globalApiCall('/person/studios', 'GET');
      
      // Build studios list similar to Go version
      // Start with "Personal" studio for web storage
      const personal = {
        id: 'personal',
        name: 'Personal',
        url: '/web/projects',
        alt_url: '',
        users: [],
      };
      
      const studios = [personal];
      
      // Handle both array and object responses
      const studioArray = Array.isArray(userStudios) ? userStudios : 
                          (userStudios && typeof userStudios === 'object') ? Object.values(userStudios) : [];
      
      for (const userStudio of studioArray) {
        if (userStudio && userStudio.name) {
          studios.push({
            id: userStudio.id || '',
            name: userStudio.name || '',
            url: userStudio.url || userStudio.URL || '',
            alt_url: userStudio.alt_url || userStudio.AltURL || '',
            users: userStudio.users || userStudio.Users || [],
          });
        }
      }
      
      // Cache in localStorage
      setSetting('studios', studios);
      
      return studios;
    } catch (error) {
      console.warn('[GetStudios] Failed to fetch studios from API:', error);
      
      // Try to get cached studios
      const cachedStudios = getSetting('studios', []);
      if (cachedStudios.length > 0) {
        return cachedStudios;
      }
      
      // Return at least the personal studio
      return [{
        id: 'personal',
        name: 'Personal',
        url: '/web/projects',
        alt_url: '',
        users: [],
      }];
    }
  },
  
  GetPinnedProjects: async (studioName) => getSetting(`pinnedProjects_${studioName}`, []),
  SetPinnedProjects: async (studioName, projects) => setSetting(`pinnedProjects_${studioName}`, projects),
  AddPinnedProject: async (studioName, projectId) => {
    const pinned = getSetting(`pinnedProjects_${studioName}`, []);
    if (!pinned.includes(projectId)) {
      pinned.push(projectId);
      setSetting(`pinnedProjects_${studioName}`, pinned);
    }
    return pinned;
  },
  RemovePinnedProject: async (studioName, projectId) => {
    const pinned = getSetting(`pinnedProjects_${studioName}`, []).filter(id => id !== projectId);
    setSetting(`pinnedProjects_${studioName}`, pinned);
    return pinned;
  },
  GetRecentProjects: async (studioName) => getSetting(`recentProjects_${studioName}`, []),
  AddRecentProject: async (studioName, projectId) => {
    let recent = getSetting(`recentProjects_${studioName}`, []);
    recent = [projectId, ...recent.filter(id => id !== projectId)].slice(0, 10);
    setSetting(`recentProjects_${studioName}`, recent);
    return recent;
  },
  ClearRecentProject: async (studioName) => setSetting(`recentProjects_${studioName}`, []),
  
  // Locations (limited in web mode)
  GetDefaultLocation: async () => ({ id: 'web', name: 'Web Storage', path: '/web' }),
  GetAllLocationPaths: async () => [{ id: 'web', name: 'Web Storage', path: '/web' }],
  GetProjectDirectory: async () => '/web/projects',
  GetProjectLocation: async (projectId) => 'web',
  AssignProjectToLocation: async (projectId, locationId) => {},
  AddProjectLocation: async (name, path) => ({ id: 'web', name, path }),
  RemoveProjectLocation: async (locationId) => {},
  SetDefaultLocation: async (locationId) => {},
  GetLocationUsage: async (locationId) => 0,
  CanDeleteLocation: async (locationId) => false,
  CheckLocationHealth: async (locationId) => ({ healthy: true, locationId }),
  CheckAllLocationsHealth: async () => [{ healthy: true, locationId: 'web' }],
  
  // Workspaces
  GetProjectWorkspaces: async (projectId) => getSetting(`workspaces_${projectId}`, []),
  AddProjectWorkspace: async (projectId, workspace) => {
    const workspaces = getSetting(`workspaces_${projectId}`, []);
    workspaces.push(workspace);
    setSetting(`workspaces_${projectId}`, workspaces);
  },
  RemoveProjectWorkspace: async (projectId, workspaceId) => {
    const workspaces = getSetting(`workspaces_${projectId}`, []).filter(w => w.id !== workspaceId);
    setSetting(`workspaces_${projectId}`, workspaces);
  },
  
  // Version & EULA
  GetCurrentVersion: async () => '0.2.0-web',
  GetEulaAccepted: async () => getSetting('eulaAccepted', false),
  SetEulaAccepted: async (accepted) => setSetting('eulaAccepted', accepted),
};

// =============================================================================
// AUTH SERVICE
// =============================================================================

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

// =============================================================================
// STUDIO SERVICE
// =============================================================================

export const StudioService = {
  RegisterStudio: async (name, studioUrl) => {
    return await globalApiCall('/studio', 'POST', { name, url: studioUrl });
  },
  
  GetStudioUsers: async (studioId) => {
    const users = await globalApiCall(`/studio/${studioId}/persons`, 'GET');
    return Array.isArray(users) ? users : [];
  },
  
  AddCollaborator: async (email, studioId, roleName) => {
    return await globalApiCall('/studio/person', 'POST', {
      email,
      studio_id: studioId,
      role_name: roleName,
    });
  },
  
  RemoveCollaborator: async (userId, studioId) => {
    return await globalApiCall(`/studio/person/${studioId}/${userId}`, 'DELETE');
  },
  
  ChangeCollaboratorRole: async (userId, studioId, roleName) => {
    return await globalApiCall('/studio/person', 'PUT', {
      user_id: userId,
      studio_id: studioId,
      role_name: roleName,
    });
  },
  
  CheckStudioNameExists: async (studioName) => {
    try {
      const response = await globalApiCall(`/check-studio-availability/${encodeURIComponent(studioName)}`, 'GET');
      return response.exists === true;
    } catch {
      return false;
    }
  },
  
  GetStudioStatus: async (studioUrl) => {
    try {
      const response = await fetch(`${studioUrl}/ping`, { method: 'GET' });
      if (response.ok) {
        return 'online';
      }
      return 'offline';
    } catch {
      return 'offline';
    }
  },
  
  UpdateStudio: async (studioName, url, altUrl, port, key) => {
    return await globalApiCall(`/studio/${studioName}/url`, 'PUT', {
      url,
      alt_url: altUrl,
      port,
      key,
    });
  },
  
  VerifyDeploymentCode: async (code) => {
    const response = await globalApiCall('/studio/verify-deployment-code', 'POST', { code });
    return [response.valid === true, response.studio_url || ''];
  },
};

// =============================================================================
// PROJECT SERVICE
// =============================================================================

export const ProjectService = {
  GetStudioProjects: async (url, studioName) => {
    setActiveStudioUrl(url);
    const projects = await studioApiCall(url, '/projects', 'GET');
    return Array.isArray(projects) ? projects : [];
  },
  
  CreateProject: async (projectUri, studioName, workingDir, templateName) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectUri.split('/').pop()?.replace('.clst', '') || projectUri;
    return await studioApiCall(studioUrl, `/${projectName}`, 'POST', {
      working_dir: workingDir,
      template: templateName,
    });
  },
  
  ProjectInfo: async (projectPath) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
    return await studioApiCall(studioUrl, `/${projectName}`, 'GET');
  },
  
  ProjectsInfo: async (projectPaths) => {
    const results = await Promise.all(
      projectPaths.map(path => ProjectService.ProjectInfo(path).catch(() => null))
    );
    return results.filter(Boolean);
  },
  
  Rename: async (projectUri, studioName, newName) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectUri.split('/').pop()?.replace('.clst', '') || projectUri;
    await studioApiCall(studioUrl, `/${projectName}`, 'PUT', { name: newName });
  },
  
  UpdateIcon: async (projectUri, studioName, iconValue) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectUri.split('/').pop()?.replace('.clst', '') || projectUri;
    await studioApiCall(studioUrl, `/${projectName}/icon`, 'PUT', { icon: iconValue });
  },
  
  ToggleCloseProject: async (projectUri, studioName) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectUri.split('/').pop()?.replace('.clst', '') || projectUri;
    await studioApiCall(studioUrl, `/${projectName}/toggle-close`, 'PUT');
  },
  
  CloseProject: async (projectPath) => {
    await ProjectService.ToggleCloseProject(projectPath, '');
  },
  
  GetIgnoreList: async (projectPath) => {
    // TODO: Implement via studio server or return default
    return [];
  },
  
  SetIgnoreList: async (projectUri, studioName, ignoreList) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectUri.split('/').pop()?.replace('.clst', '') || projectUri;
    await studioApiCall(studioUrl, `/${projectName}/ignore-list`, 'PUT', ignoreList);
  },
  
  GetSyncToken: async (projectUri) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectUri.split('/').pop()?.replace('.clst', '') || projectUri;
    return await studioApiCall(studioUrl, `/${projectName}/sync-token`, 'GET');
  },
  
  GetPreview: async (projectPath) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
    try {
      const response = await fetch(`${studioUrl}/${projectName}/preview`);
      if (response.ok) {
        const blob = await response.blob();
        return URL.createObjectURL(blob);
      }
    } catch {}
    return '';
  },
  
  UpdatePreview: async (projectPath, previewPath) => {
    // TODO: Implement file upload to studio server
    console.warn('UpdatePreview not implemented in web mode');
  },
  
  GetIsClose: async (projectPath) => {
    const info = await ProjectService.ProjectInfo(projectPath);
    return info?.is_closed || false;
  },
  
  // User management within projects
  AddUser: async (projectPath, email, roleName) => {
    // TODO: Implement via studio server
    console.warn('AddUser not implemented in web mode');
    return {};
  },
  
  RemoveUser: async (projectPath, userId) => {
    // TODO: Implement via studio server
    console.warn('RemoveUser not implemented in web mode');
  },
  
  ChangeRole: async (projectPath, userId, roleName) => {
    // TODO: Implement via studio server
    console.warn('ChangeRole not implemented in web mode');
  },
  
  UserInProject: async (projectPath, userId) => {
    // TODO: Implement via studio server
    return true;
  },
  
  // Templates
  GetTemplates: async () => {
    // TODO: Fetch from studio server or return defaults
    return [];
  },
  
  ApplyTemplate: async (projectPath, templateName) => {
    // TODO: Implement via studio server
    console.warn('ApplyTemplate not implemented in web mode');
  },
  
  ResetDefaultTemplates: async () => {
    // TODO: Implement via studio server
    console.warn('ResetDefaultTemplates not implemented in web mode');
  },
  
  // Untracked items (local file system - limited in web)
  GetFolderUntrackedItems: async (projectWorkingDir, directory, ignoreList, tracked) => {
    return { tasks: [], entities: [] };
  },
  
  IsIgnored: async (itemPath, ignoreList) => {
    return false;
  },
  
  // Working directory
  UpdateWorkingDirectory: async (projectUri, studioName, newWorkingDir) => {
    // Not applicable in web mode
    console.warn('UpdateWorkingDirectory not applicable in web mode');
  },
  
  Purge: async (projectPath) => {
    // TODO: Implement via studio server
    console.warn('Purge not implemented in web mode');
  },
};

// =============================================================================
// SYNC SERVICE
// =============================================================================

export const SyncService = {
  PullData: async (projectPath, remoteURL, pullChunk, syncOptions) => {
    // TODO: Implement full sync with protobuf support
    console.warn('PullData - full sync not yet implemented in web mode');
  },
  
  SyncData: async (projectPath, remoteURL, pullChunk, syncOptions) => {
    // TODO: Implement full sync
    console.warn('SyncData - full sync not yet implemented in web mode');
  },
  
  CloneProject: async (projectUri, studioName, workingDir, syncOptions) => {
    // TODO: Implement clone
    console.warn('CloneProject not yet implemented in web mode');
  },
  
  PushCheckpoints: async (projectPath, remoteURL, pullChunk, syncOptions) => {
    // TODO: Implement checkpoint push
    console.warn('PushCheckpoints not yet implemented in web mode');
  },
  
  PullLatestCheckpoints: async (projectPath, remoteURL) => {
    // TODO: Implement checkpoint pull
    console.warn('PullLatestCheckpoints not yet implemented in web mode');
  },
  
  DownloadCheckpoint: async (projectPath, remoteURL, checkpointId) => {
    // TODO: Implement checkpoint download
    console.warn('DownloadCheckpoint not yet implemented in web mode');
  },
  
  IsUnsynced: async (projectPath) => {
    // TODO: Compare local vs remote sync tokens
    return false;
  },
  
  CancelSync: async () => {
    // Nothing to cancel in web mode currently
  },
};

// =============================================================================
// APP SERVICE
// =============================================================================

export const AppService = {
  GetOS: async () => 'web',
  GetVersion: async () => '0.2.0-web',
  OpenURL: async (url) => window.open(url, '_blank'),
  GetAppDataDir: async () => '/web/data',
};

// =============================================================================
// LOG SERVICE
// =============================================================================

export const LogService = {
  LogInfo: async (message) => console.log('[INFO]', message),
  LogError: async (message) => console.error('[ERROR]', message),
  LogWarning: async (message) => console.warn('[WARN]', message),
  LogDebug: async (message) => console.debug('[DEBUG]', message),
};

// =============================================================================
// PROFILE SERVICE
// =============================================================================

export const ProfileService = {
  // Get complete user profile
  GetUserProfile: async (userId) => {
    const response = await globalApiCall(`/api/users/${userId}/profile`, 'GET');
    return response || {};
  },
  
  // Update user profile fields
  UpdateUserProfile: async (userId, profileData) => {
    return await globalApiCall(`/api/users/${userId}/profile`, 'PUT', profileData);
  },
  
  // Update user photo (base64)
  UpdateUserPhoto: async (photoBase64) => {
    return await globalApiCall('/person/photo', 'POST', { photo: photoBase64 });
  },
  
  // User Tools
  GetUserTools: async (userId) => {
    const tools = await globalApiCall(`/api/users/${userId}/tools`, 'GET');
    return Array.isArray(tools) ? tools : [];
  },
  
  AddUserTool: async (userId, toolData) => {
    return await globalApiCall(`/api/users/${userId}/tools`, 'POST', toolData);
  },
  
  UpdateUserTool: async (userId, toolId, proficiencyLevel) => {
    return await globalApiCall(`/api/users/${userId}/tools/${toolId}`, 'PUT', { proficiency_level: proficiencyLevel });
  },
  
  RemoveUserTool: async (userId, toolId) => {
    return await globalApiCall(`/api/users/${userId}/tools/${toolId}`, 'DELETE');
  },
  
  // User Skills
  GetUserSkills: async (userId) => {
    const skills = await globalApiCall(`/api/users/${userId}/skills`, 'GET');
    return Array.isArray(skills) ? skills : [];
  },
  
  AddUserSkill: async (userId, skillData) => {
    return await globalApiCall(`/api/users/${userId}/skills`, 'POST', skillData);
  },
  
  UpdateUserSkill: async (userId, skillId, proficiencyLevel) => {
    return await globalApiCall(`/api/users/${userId}/skills/${skillId}`, 'PUT', { proficiency_level: proficiencyLevel });
  },
  
  RemoveUserSkill: async (userId, skillId) => {
    return await globalApiCall(`/api/users/${userId}/skills/${skillId}`, 'DELETE');
  },
  
  // Reference data - all available tools and skills
  GetAllTools: async () => {
    const tools = await globalApiCall('/api/tools', 'GET');
    return Array.isArray(tools) ? tools : [];
  },
  
  GetToolsByCategory: async (category) => {
    const tools = await globalApiCall(`/api/tools/category/${encodeURIComponent(category)}`, 'GET');
    return Array.isArray(tools) ? tools : [];
  },
  
  GetAllSkills: async () => {
    const skills = await globalApiCall('/api/skills', 'GET');
    return Array.isArray(skills) ? skills : [];
  },
  
  GetSkillsByCategory: async (category) => {
    const skills = await globalApiCall(`/api/skills/category/${encodeURIComponent(category)}`, 'GET');
    return Array.isArray(skills) ? skills : [];
  },
  
  // Countries and genders for profile
  GetAllCountries: async () => {
    const countries = await globalApiCall('/api/countries', 'GET');
    return Array.isArray(countries) ? countries : [];
  },
  
  GetAllGenders: async () => {
    const genders = await globalApiCall('/api/genders', 'GET');
    return Array.isArray(genders) ? genders : [];
  },
  
  // Legacy aliases for backward compatibility
  GetProfile: async (userId) => ProfileService.GetUserProfile(userId),
  UpdateProfile: async (userId, profileData) => ProfileService.UpdateUserProfile(userId, profileData),
  DeleteUserTool: async (userId, toolId) => ProfileService.RemoveUserTool(userId, toolId),
  DeleteUserSkill: async (userId, skillId) => ProfileService.RemoveUserSkill(userId, skillId),
};

// =============================================================================
// ACCOUNT SERVICE
// =============================================================================

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

// =============================================================================
// USER SERVICE (Project-level users)
// =============================================================================

export const UserService = {
  GetUsers: async (projectPath) => {
    // TODO: Implement via studio server
    return [];
  },
  
  GetUser: async (projectPath, userId) => {
    // TODO: Implement via studio server
    return {};
  },
  
  UpdateUser: async (projectPath, user) => {
    // TODO: Implement via studio server
    return {};
  },
  
  GetRoles: async (projectPath) => {
    // TODO: Implement via studio server
    return [];
  },
  
  CreateRole: async (projectPath, role) => {
    // TODO: Implement via studio server
    return {};
  },
  
  UpdateRole: async (projectPath, role) => {
    // TODO: Implement via studio server
    return {};
  },
  
  DeleteRole: async (projectPath, roleId) => {
    // TODO: Implement via studio server
  },
};

// =============================================================================
// ASSET SERVICE
// =============================================================================

export const AssetService = {
  // These would need studio server endpoints
  GetAssets: async (projectPath, entityPath) => [],
  GetAsset: async (projectPath, assetId) => ({}),
  CreateAsset: async (projectPath, asset) => ({}),
  UpdateAsset: async (projectPath, asset) => ({}),
  DeleteAsset: async (projectPath, assetId) => {},
  MoveAsset: async (projectPath, assetId, newPath) => {},
  CopyAsset: async (projectPath, assetId, newPath) => {},
  GetAssetTypes: async (projectPath) => [],
  CreateAssetType: async (projectPath, type) => ({}),
  UpdateAssetType: async (projectPath, type) => ({}),
  DeleteAssetType: async (projectPath, typeId) => {},
};

// =============================================================================
// COLLECTION SERVICE
// =============================================================================

export const CollectionService = {
  GetCollections: async (projectPath) => [],
  GetCollection: async (projectPath, collectionId) => ({}),
  CreateCollection: async (projectPath, collection) => ({}),
  UpdateCollection: async (projectPath, collection) => ({}),
  DeleteCollection: async (projectPath, collectionId) => {},
  GetCollectionTypes: async (projectPath) => [],
  CreateCollectionType: async (projectPath, type) => ({}),
  UpdateCollectionType: async (projectPath, type) => ({}),
  DeleteCollectionType: async (projectPath, typeId) => {},
};

// =============================================================================
// CHECKPOINT SERVICE
// =============================================================================

export const CheckpointService = {
  GetCheckpoints: async (projectPath, taskId) => [],
  GetCheckpoint: async (projectPath, checkpointId) => ({}),
  CreateCheckpoint: async (projectPath, checkpoint) => ({}),
  RestoreCheckpoint: async (projectPath, checkpointId) => {},
  DeleteCheckpoint: async (projectPath, checkpointId) => {},
};

// =============================================================================
// TAG SERVICE
// =============================================================================

export const TagService = {
  GetTags: async (projectPath) => [],
  CreateTag: async (projectPath, tag) => ({}),
  UpdateTag: async (projectPath, tag) => ({}),
  DeleteTag: async (projectPath, tagId) => {},
};

// =============================================================================
// STATUS SERVICE
// =============================================================================

export const StatusService = {
  GetStatuses: async (projectPath) => [],
  CreateStatus: async (projectPath, status) => ({}),
  UpdateStatus: async (projectPath, status) => ({}),
  DeleteStatus: async (projectPath, statusId) => {},
};

// =============================================================================
// TEMPLATE SERVICE
// =============================================================================

export const TemplateService = {
  GetTemplates: async (projectPath) => [],
  CreateTemplate: async (projectPath, template) => ({}),
  UpdateTemplate: async (projectPath, template) => ({}),
  DeleteTemplate: async (projectPath, templateId) => {},
};

// =============================================================================
// WORKFLOW SERVICE
// =============================================================================

export const WorkflowService = {
  GetWorkflows: async (projectPath) => [],
  CreateWorkflow: async (projectPath, workflow) => ({}),
  UpdateWorkflow: async (projectPath, workflow) => ({}),
  DeleteWorkflow: async (projectPath, workflowId) => {},
};

// =============================================================================
// DEPENDENCY TYPE SERVICE
// =============================================================================

export const DependencyTypeService = {
  GetDependencyTypes: async (projectPath) => [],
  CreateDependencyType: async (projectPath, type) => ({}),
  UpdateDependencyType: async (projectPath, type) => ({}),
  DeleteDependencyType: async (projectPath, typeId) => {},
};

// =============================================================================
// FS SERVICE (Limited in web mode)
// =============================================================================

export const FSService = {
  WatchPath: async (path) => {},
  UnwatchPath: async (path) => {},
  GetFileInfo: async (path) => ({}),
  ReadDirectory: async (path) => [],
  OpenFile: async (path) => {
    console.warn('OpenFile not available in web mode');
  },
  OpenInExplorer: async (path) => {
    console.warn('OpenInExplorer not available in web mode');
  },
  OpenTerminal: async (path) => {
    console.warn('OpenTerminal not available in web mode');
  },
  Exists: async (path) => false,
  IsDirectory: async (path) => false,
  
  // File icon support - returns empty string in web mode
  // The icon store will fall back to fileIconIndex.json
  GetFileIcon: async (ext) => '',
  
  // Path utilities
  TempDir: async () => '/tmp',
  JoinPath: async (...paths) => paths.join('/'),
  WriteFile: async (path, data) => {
    console.warn('WriteFile not available in web mode');
  },
  ReadFile: async (path) => {
    console.warn('ReadFile not available in web mode');
    return '';
  },
  FileHash: async (path) => '',
  FileStat: async (path) => ({}),
  BaseName: async (path) => path.split('/').pop() || path,
};

// =============================================================================
// DIALOG SERVICE (Browser alternatives)
// =============================================================================

export const DialogService = {
  OpenFile: async (options) => {
    return new Promise((resolve) => {
      const input = document.createElement('input');
      input.type = 'file';
      if (options?.filters) {
        const extensions = options.filters.flatMap(f => f.extensions || []);
        input.accept = extensions.map(ext => `.${ext}`).join(',');
      }
      input.onchange = (e) => {
        const file = e.target.files?.[0];
        resolve(file ? file.name : null);
      };
      input.click();
    });
  },
  
  OpenDirectory: async (options) => {
    console.warn('OpenDirectory has limited support in web browsers');
    return null;
  },
  
  SaveFile: async (options) => {
    console.warn('SaveFile has limited support in web browsers');
    return null;
  },
  
  ShowMessage: async (options) => {
    alert(options?.message || options?.title || '');
  },
  
  ShowError: async (title, message) => {
    alert(`Error: ${title}\n${message}`);
  },
  
  ShowWarning: async (title, message) => {
    alert(`Warning: ${title}\n${message}`);
  },
  
  ShowInfo: async (title, message) => {
    alert(`${title}\n${message}`);
  },
  
  AskConfirmation: async (title, message) => {
    return confirm(`${title}\n${message}`);
  },
};

// =============================================================================
// CLIPBOARD SERVICE
// =============================================================================

export const ClipboardService = {
  WriteText: async (text) => {
    try {
      await navigator.clipboard.writeText(text);
    } catch (err) {
      console.error('Clipboard write failed:', err);
    }
  },
  
  ReadText: async () => {
    try {
      return await navigator.clipboard.readText();
    } catch (err) {
      console.error('Clipboard read failed:', err);
      return '';
    }
  },
};

// =============================================================================
// TRASH SERVICE
// =============================================================================

export const TrashService = {
  GetTrashedItems: async (projectPath) => [],
  RestoreItem: async (projectPath, itemId) => {},
  DeletePermanently: async (projectPath, itemId) => {},
  EmptyTrash: async (projectPath) => {},
};

// =============================================================================
// IMPORT SERVICE
// =============================================================================

export const ImportService = {
  ImportFiles: async (projectPath, files) => [],
  ImportFolder: async (projectPath, folderPath) => [],
};

// =============================================================================
// DEPLOYMENT SERVICE
// =============================================================================

export const DeploymentService = {
  Deploy: async (options) => {
    return await globalApiCall('/api/deploy', 'POST', options);
  },
  
  GetDeploymentStatus: async (deploymentId) => {
    return await globalApiCall(`/api/deploy/${deploymentId}/status`, 'GET');
  },
  
  DestroyDeployment: async (deploymentId) => {
    return await globalApiCall(`/api/deploy/${deploymentId}`, 'DELETE');
  },
  
  GetDeploymentConfig: async () => {
    return await globalApiCall('/api/deploy/config', 'GET');
  },
};
