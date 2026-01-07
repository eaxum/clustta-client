import { globalApiCall } from './http-client.js';
import { setSetting, getSetting } from './storage.js';
import { StudioService } from './studioservice.js';

export const SettingsService = {
  // Returns whether to use alternate URL for studio connections
  GetUseAltUrl: async () => getSetting('useAltUrl', false),
  // Sets whether to use alternate URL for studio connections
  SetUseAltUrl: async (value) => setSetting('useAltUrl', value),
  // Returns whether project view is in grid mode
  IsProjectGridView: async () => getSetting('projectGridView', true),
  // Sets project view to grid or list mode
  SetProjectGridView: async (value) => setSetting('projectGridView', value),
  // Returns whether to use grid layout
  GetUseGrid: async () => getSetting('useGrid', true),
  // Sets grid layout preference
  SetUseGrid: async (value) => setSetting('useGrid', value),
  // Returns whether to show untracked projects
  IsShowUntrackedProjects: async () => getSetting('showUntrackedProjects', true),
  // Sets whether to show untracked projects
  SetShowUntrackedProjects: async (value) => setSetting('showUntrackedProjects', value),
  // Returns whether compact view is enabled
  IsCompactView: async () => getSetting('compactView', false),
  // Sets compact view preference
  SetCompactView: async (value) => setSetting('compactView', value),
  // Returns whether to show hidden files
  IsShowHiddenFiles: async () => getSetting('showHiddenFiles', false),
  // Sets whether to show hidden files
  SetShowHiddenFiles: async (value) => setSetting('showHiddenFiles', value),

  // Returns current theme setting
  GetTheme: async () => getSetting('theme', 'dark'),
  // Sets the application theme
  SetTheme: async (theme) => setSetting('theme', theme),
  // Returns current icon scheme
  GetIconScheme: async () => getSetting('iconScheme', 'solid'),
  // Sets the icon scheme
  SetIconScheme: async (scheme) => setSetting('iconScheme', scheme),

  // Returns the last accessed studio
  GetLastStudio: async () => getSetting('lastStudio', ''),
  // Sets the last accessed studio
  SetLastStudio: async (studio) => setSetting('lastStudio', studio),

  // Returns all studios for the current user
  GetStudios: async (path) => {
    try {
      const userStudios = await globalApiCall('/person/studios', 'GET');

      const personal = {
        id: 'personal',
        name: 'Personal',
        url: '/web/projects',
        alt_url: '',
        Users: [],
      };

      const studios = [personal];
      const studioArray = Array.isArray(userStudios) ? userStudios :
                          (userStudios && typeof userStudios === 'object') ? Object.values(userStudios) : [];

      for (const userStudio of studioArray) {
        if (userStudio && userStudio.name) {
          // Fetch users for each studio (matching Go behavior)
          let studioUsers = [];
          try {
            const studioId = userStudio.id || '';
            if (studioId) {
              studioUsers = await StudioService.GetStudioUsers(studioId);
            }
          } catch (err) {
            console.warn(`Failed to fetch users for studio ${userStudio.name}:`, err);
          }

          studios.push({
            id: userStudio.id || '',
            name: userStudio.name || '',
            url: userStudio.url || userStudio.URL || '',
            alt_url: userStudio.alt_url || userStudio.AltURL || '',
            Users: studioUsers,
          });
        }
      }

      setSetting('studios', studios);
      return studios;
    } catch (error) {
      console.warn('[GetStudios] Failed to fetch studios from API:', error);

      const cachedStudios = getSetting('studios', []);
      if (cachedStudios.length > 0) {
        return cachedStudios;
      }

      return [{
        id: 'personal',
        name: 'Personal',
        url: '/web/projects',
        alt_url: '',
        Users: [],
      }];
    }
  },

  // Returns pinned projects for a studio
  GetPinnedProjects: async (studioName) => getSetting(`pinnedProjects_${studioName}`, []),
  // Sets pinned projects for a studio
  SetPinnedProjects: async (studioName, projects) => setSetting(`pinnedProjects_${studioName}`, projects),
  // Adds a project to pinned list
  AddPinnedProject: async (studioName, projectId) => {
    const pinned = getSetting(`pinnedProjects_${studioName}`, []);
    if (!pinned.includes(projectId)) {
      pinned.push(projectId);
      setSetting(`pinnedProjects_${studioName}`, pinned);
    }
    return pinned;
  },
  // Removes a project from pinned list
  RemovePinnedProject: async (studioName, projectId) => {
    const pinned = getSetting(`pinnedProjects_${studioName}`, []).filter(id => id !== projectId);
    setSetting(`pinnedProjects_${studioName}`, pinned);
    return pinned;
  },
  // Returns recent projects for a studio
  GetRecentProjects: async (studioName) => getSetting(`recentProjects_${studioName}`, []),
  // Adds a project to recent list
  AddRecentProject: async (studioName, projectId) => {
    let recent = getSetting(`recentProjects_${studioName}`, []);
    recent = [projectId, ...recent.filter(id => id !== projectId)].slice(0, 10);
    setSetting(`recentProjects_${studioName}`, recent);
    return recent;
  },
  // Clears recent projects for a studio
  ClearRecentProject: async (studioName) => setSetting(`recentProjects_${studioName}`, []),

  // Returns the default storage location
  GetDefaultLocation: async () => ({ id: 'web', name: 'Web Storage', path: '/web' }),
  // Returns all available location paths
  GetAllLocationPaths: async () => [{ id: 'web', name: 'Web Storage', path: '/web' }],
  // Returns the project directory path
  GetProjectDirectory: async () => '/web/projects',
  // Returns the location for a specific project
  GetProjectLocation: async (projectId) => 'web',
  // Assigns a project to a location
  AssignProjectToLocation: async (projectId, locationId) => {},
  // Adds a new project location
  AddProjectLocation: async (name, path) => ({ id: 'web', name, path }),
  // Removes a project location
  RemoveProjectLocation: async (locationId) => {},
  // Sets the default location
  SetDefaultLocation: async (locationId) => {},
  // Returns storage usage for a location
  GetLocationUsage: async (locationId) => 0,
  // Checks if a location can be deleted
  CanDeleteLocation: async (locationId) => false,
  // Checks health status of a location
  CheckLocationHealth: async (locationId) => ({ healthy: true, locationId }),
  // Checks health status of all locations
  CheckAllLocationsHealth: async () => [{ healthy: true, locationId: 'web' }],

  // Returns workspaces for a project
  GetProjectWorkspaces: async (projectId) => getSetting(`workspaces_${projectId}`, []),
  // Adds a workspace to a project
  AddProjectWorkspace: async (projectId, workspace) => {
    const workspaces = getSetting(`workspaces_${projectId}`, []);
    workspaces.push(workspace);
    setSetting(`workspaces_${projectId}`, workspaces);
  },
  // Removes a workspace from a project
  RemoveProjectWorkspace: async (projectId, workspaceId) => {
    const workspaces = getSetting(`workspaces_${projectId}`, []).filter(w => w.id !== workspaceId);
    setSetting(`workspaces_${projectId}`, workspaces);
  },

  // Returns the current application version
  GetCurrentVersion: async () => '0.2.0-web',
  // Returns whether EULA has been accepted
  GetEulaAccepted: async () => getSetting('eulaAccepted', false),
  // Sets EULA acceptance status
  SetEulaAccepted: async (accepted) => setSetting('eulaAccepted', accepted),
};
