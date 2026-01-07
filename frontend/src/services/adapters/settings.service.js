// =============================================================================
// SETTINGS SERVICE (localStorage-based)
// =============================================================================

import { globalApiCall } from './http-client.js';
import { getSettings, setSetting, getSetting } from './storage.js';

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
