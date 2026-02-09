import { globalApiCall } from './http-client.js';

export const StudioService = {
  // Registers a new studio with the given name and URL
  RegisterStudio: async (name, studioUrl) => {
    return await globalApiCall('/studio', 'POST', { name, url: studioUrl });
  },

  // Returns all users in a studio
  GetStudioUsers: async (studioId) => {
    const users = await globalApiCall(`/studio/${studioId}/persons`, 'GET');
    return Array.isArray(users) ? users : [];
  },

  // Adds a collaborator to a studio
  AddCollaborator: async (email, studioId, roleName) => {
    return await globalApiCall('/studio/person', 'POST', {
      email,
      studio_id: studioId,
      role_name: roleName,
    });
  },

  // Removes a collaborator from a studio
  RemoveCollaborator: async (userId, studioId) => {
    return await globalApiCall(`/studio/person/${studioId}/${userId}`, 'DELETE');
  },

  // Changes a collaborator's role in a studio
  ChangeCollaboratorRole: async (userId, studioId, roleName) => {
    return await globalApiCall('/studio/person', 'PUT', {
      user_id: userId,
      studio_id: studioId,
      role_name: roleName,
    });
  },

  // Checks if a studio name is already taken
  CheckStudioNameExists: async (studioName) => {
    try {
      const response = await globalApiCall(`/check-studio-availability/${encodeURIComponent(studioName)}`, 'GET');
      return response.exists === true;
    } catch {
      return false;
    }
  },

  // Returns the online/offline status of a studio server
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

  // Returns the version of a studio server
  GetServerVersion: async (studioUrl) => {
    try {
      const response = await fetch(`${studioUrl}/version`, { method: 'GET' });
      if (response.ok) {
        const data = await response.json();
        return data.version || '';
      }
      return '';
    } catch {
      return '';
    }
  },

  // Updates studio connection details
  UpdateStudio: async (studioName, url, altUrl, port, key) => {
    return await globalApiCall(`/studio/${studioName}/url`, 'PUT', {
      url,
      alt_url: altUrl,
      port,
      key,
    });
  },

  // Verifies a deployment code and returns validity and studio URL
  VerifyDeploymentCode: async (code) => {
    const response = await globalApiCall('/studio/verify-deployment-code', 'POST', { code });
    return [response.valid === true, response.studio_url || ''];
  },

  // Races primary and alternative studio URLs, returning whichever responds first.
  // Falls back to the primary URL if no alternative is set.
  ResolveStudioUrl: async (url, altUrl) => {
    if (!altUrl) return url;

    const ping = (studioUrl) =>
      fetch(`${studioUrl}/ping`, { method: 'GET', signal: AbortSignal.timeout(5000) })
        .then((res) => {
          if (res.ok) return studioUrl;
          throw new Error('not ok');
        });

    try {
      return await Promise.any([ping(url), ping(altUrl)]);
    } catch {
      return url;
    }
  },
};
