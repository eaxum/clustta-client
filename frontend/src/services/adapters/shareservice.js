import { globalApiCall } from './http-client.js';

export const ShareService = {
  // Creates a shareable download link for one or more checkpoints.
  CreateShareLink: async (studioId, projectName, checkpointIds, label, expiresInHours) => {
    return await globalApiCall('/share', 'POST', {
      studio_id: studioId,
      project_name: projectName,
      checkpoint_ids: checkpointIds,
      label,
      expires_in_hours: expiresInHours,
    });
  },

  // Revokes a share link.
  RevokeShareLink: async (token) => {
    return await globalApiCall(`/share/${token}`, 'DELETE');
  },
};
