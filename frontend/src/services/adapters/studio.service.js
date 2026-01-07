// =============================================================================
// STUDIO SERVICE
// =============================================================================

import { globalApiCall } from './http-client.js';

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
