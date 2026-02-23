import { defineStore } from 'pinia';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
// IntegrationService will be available after bindings are generated

export const useIntegrationStore = defineStore('integrations', {
  state: () => ({
    availableIntegrations: [],    // All registered integrations
    linkedIntegration: null,      // Current project's linked integration
    tokens: {},                   // Stored tokens per integration: { kitsu: { token, apiUrl, userId }, ... }
    isLoading: false,
    isAuthenticating: false,
    syncPreview: null,            // Current sync preview data
    isSyncing: false,
    lastSyncAt: null,
  }),

  getters: {
    // Get token for a specific integration
    getToken: (state) => (integrationId) => {
      return state.tokens[integrationId]?.token || null;
    },

    // Get API URL for a specific integration
    getApiUrl: (state) => (integrationId) => {
      return state.tokens[integrationId]?.apiUrl || null;
    },

    // Check if authenticated with a specific integration
    isAuthenticated: (state) => (integrationId) => {
      return !!state.tokens[integrationId]?.token;
    },

    // Check if current project has an integration linked
    hasLinkedIntegration: (state) => {
      return !!state.linkedIntegration;
    },

    // Get the linked integration type
    linkedIntegrationId: (state) => {
      return state.linkedIntegration?.integration_id || null;
    },

    // Get integration info by ID
    getIntegration: (state) => (integrationId) => {
      return state.availableIntegrations.find(i => i.id === integrationId) || null;
    },

    // Collections to be synced (from preview)
    collectionsToSync: (state) => {
      if (!state.syncPreview) return [];
      return state.syncPreview.collections || [];
    },

    // Assets to be synced (from preview)
    assetsToSync: (state) => {
      if (!state.syncPreview) return [];
      return state.syncPreview.assets || [];
    },
  },

  actions: {
    // Load available integrations from backend
    async loadAvailableIntegrations() {
      try {
        const { IntegrationService } = await import('@/services');
        const integrations = await IntegrationService.GetAvailableIntegrations();
        this.availableIntegrations = integrations || [];
      } catch (error) {
        console.error('Failed to load integrations:', error);
        this.availableIntegrations = [];
      }
    },

    // Load linked integration for current project
    async loadLinkedIntegration() {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.path) {
        this.linkedIntegration = null;
        return;
      }

      try {
        const { IntegrationService } = await import('@/services');
        const integration = await IntegrationService.GetLinkedIntegration(projectStore.activeProject.path);
        this.linkedIntegration = integration;
      } catch (error) {
        // No integration linked - this is expected for most projects
        this.linkedIntegration = null;
      }
    },

    // Authenticate with an integration
    async authenticate(integrationId, credentials) {
      const notificationStore = useNotificationStore();
      this.isAuthenticating = true;

      try {
        const { IntegrationService } = await import('@/services');
        const result = await IntegrationService.Authenticate(integrationId, credentials);

        if (result.success) {
          // Store token locally
          this.tokens[integrationId] = {
            token: result.access_token,
            apiUrl: credentials.api_url || '',
            userId: result.user_id,
            userName: result.user_name,
            userEmail: result.user_email,
          };
          this.saveTokens();
          return { success: true, user: result };
        } else {
          notificationStore.sendNotification(result.error || 'Authentication failed', 'error');
          return { success: false, error: result.error };
        }
      } catch (error) {
        notificationStore.sendNotification(error.message || 'Authentication failed', 'error');
        return { success: false, error: error.message };
      } finally {
        this.isAuthenticating = false;
      }
    },

    // Validate stored token
    async validateToken(integrationId) {
      const tokenData = this.tokens[integrationId];
      if (!tokenData?.token) return false;

      try {
        const { IntegrationService } = await import('@/services');
        const valid = await IntegrationService.ValidateToken(integrationId, tokenData.token, tokenData.apiUrl);
        if (!valid) {
          // Token expired, remove it
          delete this.tokens[integrationId];
          this.saveTokens();
        }
        return valid;
      } catch (error) {
        return false;
      }
    },

    // Disconnect from an integration (remove stored token)
    disconnect(integrationId) {
      delete this.tokens[integrationId];
      this.saveTokens();
    },

    // Get external projects for an integration
    async getExternalProjects(integrationId) {
      const tokenData = this.tokens[integrationId];
      if (!tokenData?.token) {
        throw new Error('Not authenticated with ' + integrationId);
      }

      try {
        const { IntegrationService } = await import('@/services');
        return await IntegrationService.GetExternalProjects(
          integrationId,
          tokenData.token,
          tokenData.apiUrl
        );
      } catch (error) {
        console.error('Failed to get external projects:', error);
        throw error;
      }
    },

    // Link current project to an external project
    async linkProject(integrationId, externalProjectId, externalProjectName, syncOptions = {}) {
      const projectStore = useProjectStore();
      const notificationStore = useNotificationStore();
      const tokenData = this.tokens[integrationId];

      if (!projectStore.activeProject?.path) {
        throw new Error('No active project');
      }
      if (!tokenData?.token) {
        throw new Error('Not authenticated with ' + integrationId);
      }

      try {
        const { IntegrationService } = await import('@/services');
        const result = await IntegrationService.LinkProject(
          projectStore.activeProject.path,
          integrationId,
          externalProjectId,
          externalProjectName,
          tokenData.apiUrl,
          JSON.stringify(syncOptions),
          tokenData.userId
        );
        this.linkedIntegration = result;
        notificationStore.sendNotification('Project linked to ' + externalProjectName, 'success');
        return result;
      } catch (error) {
        notificationStore.sendNotification(error.message || 'Failed to link project', 'error');
        throw error;
      }
    },

    // Unlink current project from external integration
    async unlinkProject() {
      const projectStore = useProjectStore();
      const notificationStore = useNotificationStore();

      if (!projectStore.activeProject?.path) {
        throw new Error('No active project');
      }

      try {
        const { IntegrationService } = await import('@/services');
        await IntegrationService.UnlinkProject(projectStore.activeProject.path);
        this.linkedIntegration = null;
        this.syncPreview = null;
        notificationStore.sendNotification('Integration unlinked', 'success');
      } catch (error) {
        notificationStore.sendNotification(error.message || 'Failed to unlink project', 'error');
        throw error;
      }
    },

    // Get sync preview
    async getSyncPreview() {
      const projectStore = useProjectStore();
      const tokenData = this.tokens[this.linkedIntegration?.integration_id];

      if (!projectStore.activeProject?.path || !tokenData?.token) {
        throw new Error('Not ready to sync');
      }

      this.isLoading = true;
      try {
        const { IntegrationService } = await import('@/services');
        const preview = await IntegrationService.GetSyncPreview(
          projectStore.activeProject.path,
          tokenData.token
        );
        this.syncPreview = preview;
        return preview;
      } finally {
        this.isLoading = false;
      }
    },

    // Execute sync
    async executeSync(selectedCollections, selectedAssets) {
      const projectStore = useProjectStore();
      const notificationStore = useNotificationStore();
      const tokenData = this.tokens[this.linkedIntegration?.integration_id];

      if (!projectStore.activeProject?.path || !tokenData?.token) {
        throw new Error('Not ready to sync');
      }

      this.isSyncing = true;
      try {
        const { IntegrationService } = await import('@/services');
        await IntegrationService.ExecuteSync(
          projectStore.activeProject.path,
          tokenData.token,
          selectedCollections,
          selectedAssets
        );
        this.lastSyncAt = new Date().toISOString();
        notificationStore.sendNotification('Sync completed successfully', 'success');
      } catch (error) {
        notificationStore.sendNotification(error.message || 'Sync failed', 'error');
        throw error;
      } finally {
        this.isSyncing = false;
      }
    },

    // Save tokens to localStorage
    saveTokens() {
      try {
        localStorage.setItem('integration_tokens', JSON.stringify(this.tokens));
      } catch (error) {
        console.error('Failed to save integration tokens:', error);
      }
    },

    // Load tokens from localStorage
    loadTokens() {
      try {
        const stored = localStorage.getItem('integration_tokens');
        if (stored) {
          this.tokens = JSON.parse(stored);
        }
      } catch (error) {
        console.error('Failed to load integration tokens:', error);
        this.tokens = {};
      }
    },

    // Initialize store
    async initialize() {
      this.loadTokens();
      await this.loadAvailableIntegrations();
    },

    // Reset store state (called when project changes)
    reset() {
      this.linkedIntegration = null;
      this.syncPreview = null;
      this.isSyncing = false;
    },
  },
});
