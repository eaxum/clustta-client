// =============================================================================
// PROJECT SERVICE
// =============================================================================

import { studioApiCall, getActiveStudioUrl, setActiveStudioUrl } from './http-client.js';

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
