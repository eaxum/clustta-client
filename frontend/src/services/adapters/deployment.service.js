// =============================================================================
// DEPLOYMENT SERVICE
// =============================================================================

import { globalApiCall } from './http-client.js';

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
