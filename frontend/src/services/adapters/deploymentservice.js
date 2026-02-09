import { globalApiCall } from './http-client.js';

export const DeploymentService = {
  // Creates a new deployment
  Deploy: async (options) => {
    return await globalApiCall('/api/deploy', 'POST', options);
  },

  // Returns the status of a deployment
  GetDeploymentStatus: async (deploymentId) => {
    return await globalApiCall(`/api/deploy/${deploymentId}/status`, 'GET');
  },

  // Destroys a deployment
  DestroyDeployment: async (deploymentId) => {
    return await globalApiCall(`/api/deploy/${deploymentId}`, 'DELETE');
  },

  // Returns deployment configuration
  GetDeploymentConfig: async () => {
    return await globalApiCall('/api/deploy/config', 'GET');
  },
};
