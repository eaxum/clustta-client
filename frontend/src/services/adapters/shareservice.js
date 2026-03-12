import { studioApiCall } from './http-client.js';

export const ShareService = {
  // Creates a shareable download link for one or more checkpoints.
  CreateShareLink: async (studioUrl, projectName, checkpointIds, label, expiresInHours) => {
    return await studioApiCall(`/${projectName}/share`, 'POST', {
      checkpoint_ids: checkpointIds,
      label,
      expires_in_hours: expiresInHours,
    });
  },
};
