import { studioApiCall, getActiveStudioUrl, setActiveStudioUrl } from './http-client.js';
import { CLUSTTA_AGENT, STORAGE_KEYS, isDev } from './config.js';

// Fetches preview for a single project and updates it in place
const fetchProjectPreview = async (project, studioUrl) => {
  try {
    const user = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || '{}');
    const headers = {
      'Clustta-Agent': CLUSTTA_AGENT,
      'UserId': user.id || '',
      'UserData': JSON.stringify(user),
    };

    let fetchUrl;
    if (isDev) {
      fetchUrl = `/studio-proxy/${project.name}/preview`;
      headers['X-Studio-URL'] = studioUrl;
    } else {
      fetchUrl = `${studioUrl}/${project.name}/preview`;
    }

    const response = await fetch(fetchUrl, { headers });
    if (response.ok) {
      const blob = await response.blob();
      if (blob.size > 0) {
        project.preview = URL.createObjectURL(blob);
      }
    }
  } catch {
    // Silently fail - preview is optional
  }
};

export const ProjectService = {
  // Returns all projects for a studio
  GetStudioProjects: async (url, studioName) => {
    setActiveStudioUrl(url);
    const projects = await studioApiCall(url, '/projects', 'GET');
    
    if (!Array.isArray(projects)) {
      return [];
    }

    // Parse projects for web mode with empty preview initially
    const parsedProjects = projects.map(project => ({
      ...project,
      has_remote: true,
      uri: `${url}/${project.name}`,
      remote: `${url}/${project.name}`,
      working_directory: '',
      is_downloaded: false,
      is_tracked: true,
      sync_token: project.sync_token || '',
      preview: '',
    }));

    // Fire-and-forget: fetch previews async without blocking
    parsedProjects.forEach(project => {
      fetchProjectPreview(project, url);
    });

    return parsedProjects;
  },

  // Creates a new project in the studio
  CreateProject: async (projectUri, studioName, workingDir, templateName) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectUri.split('/').pop()?.replace('.clst', '') || projectUri;
    return await studioApiCall(studioUrl, `/${projectName}`, 'POST', {
      working_dir: workingDir,
      template: templateName,
    });
  },

  // Returns detailed information about a project
  ProjectInfo: async (projectPath) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
    return await studioApiCall(studioUrl, `/${projectName}`, 'GET');
  },

  // Returns information about multiple projects
  ProjectsInfo: async (projectPaths) => {
    const results = await Promise.all(
      projectPaths.map(path => ProjectService.ProjectInfo(path).catch(() => null))
    );
    return results.filter(Boolean);
  },

  // Renames a project
  Rename: async (projectUri, studioName, newName) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectUri.split('/').pop()?.replace('.clst', '') || projectUri;
    await studioApiCall(studioUrl, `/${projectName}`, 'PUT', { name: newName });
  },

  // Updates the project icon
  UpdateIcon: async (projectUri, studioName, iconValue) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectUri.split('/').pop()?.replace('.clst', '') || projectUri;
    await studioApiCall(studioUrl, `/${projectName}/icon`, 'PUT', { icon: iconValue });
  },

  // Toggles the closed state of a project
  ToggleCloseProject: async (projectUri, studioName) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectUri.split('/').pop()?.replace('.clst', '') || projectUri;
    await studioApiCall(studioUrl, `/${projectName}/toggle-close`, 'PUT');
  },

  // Closes a project
  CloseProject: async (projectPath) => {
    await ProjectService.ToggleCloseProject(projectPath, '');
  },

  // Returns the ignore list for a project
  GetIgnoreList: async (projectPath) => {
    return [];
  },

  // Sets the ignore list for a project
  SetIgnoreList: async (projectUri, studioName, ignoreList) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectUri.split('/').pop()?.replace('.clst', '') || projectUri;
    await studioApiCall(studioUrl, `/${projectName}/ignore-list`, 'PUT', ignoreList);
  },

  // Returns the sync token for a project
  GetSyncToken: async (projectUri) => {
    const studioUrl = getActiveStudioUrl();
    const projectName = projectUri.split('/').pop()?.replace('.clst', '') || projectUri;
    return await studioApiCall(studioUrl, `/${projectName}/sync-token`, 'GET');
  },

  // Returns the preview image URL for a project
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

  // Updates the project preview image
  UpdatePreview: async (projectPath, previewPath) => {
    console.warn('UpdatePreview not implemented in web mode');
  },

  // Returns whether a project is closed
  GetIsClose: async (projectPath) => {
    const info = await ProjectService.ProjectInfo(projectPath);
    return info?.is_closed || false;
  },

  // Adds a user to a project
  AddUser: async (projectPath, email, roleName) => {
    console.warn('AddUser not implemented in web mode');
    return {};
  },

  // Removes a user from a project
  RemoveUser: async (projectPath, userId) => {
    console.warn('RemoveUser not implemented in web mode');
  },

  // Changes a user's role in a project
  ChangeRole: async (projectPath, userId, roleName) => {
    console.warn('ChangeRole not implemented in web mode');
  },

  // Checks if a user is in a project
  UserInProject: async (projectPath, userId) => {
    return true;
  },

  // Returns available project templates
  GetTemplates: async () => {
    return [];
  },

  // Applies a template to a project
  ApplyTemplate: async (projectPath, templateName) => {
    console.warn('ApplyTemplate not implemented in web mode');
  },

  // Resets templates to defaults
  ResetDefaultTemplates: async () => {
    console.warn('ResetDefaultTemplates not implemented in web mode');
  },

  // Returns untracked items in a folder
  GetFolderUntrackedItems: async (projectWorkingDir, directory, ignoreList, tracked) => {
    return { tasks: [], entities: [] };
  },

  // Checks if an item is ignored
  IsIgnored: async (itemPath, ignoreList) => {
    return false;
  },

  // Updates the working directory for a project
  UpdateWorkingDirectory: async (projectUri, studioName, newWorkingDir) => {
    console.warn('UpdateWorkingDirectory not applicable in web mode');
  },

  // Purges project data
  Purge: async (projectPath) => {
    console.warn('Purge not implemented in web mode');
  },
};
