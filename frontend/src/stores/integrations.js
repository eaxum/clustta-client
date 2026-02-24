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
      if (!projectStore.activeProject?.uri) {
        this.linkedIntegration = null;
        return;
      }

      try {
        const { IntegrationService } = await import('@/services');
        const integration = await IntegrationService.GetLinkedIntegration(projectStore.activeProject.uri);
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
          await this.saveTokens();
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
          await this.saveTokens();
        }
        return valid;
      } catch (error) {
        return false;
      }
    },

    // Disconnect from an integration (remove stored token)
    async disconnect(integrationId) {
      delete this.tokens[integrationId];
      const { SettingsService } = await import('@/services');
      await SettingsService.DeleteIntegrationCredential(integrationId);
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
      console.log(integrationId)
      console.log(externalProjectId)
      console.log(externalProjectName)
      const projectStore = useProjectStore();
      const notificationStore = useNotificationStore();
      const tokenData = this.tokens[integrationId];

      if (!projectStore.activeProject?.uri) {
        throw new Error('No active project');
      }
      if (!tokenData?.token) {
        throw new Error('Not authenticated with ' + integrationId);
      }

      try {
        const { IntegrationService } = await import('@/services');
        const result = await IntegrationService.LinkProject(
          String(projectStore.activeProject.uri),
          String(integrationId),
          String(externalProjectId || ''),
          String(externalProjectName || ''),
          String(tokenData.apiUrl || ''),
          JSON.stringify(syncOptions),
          String(tokenData.userId || '')
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

      if (!projectStore.activeProject?.uri) {
        throw new Error('No active project');
      }

      try {
        const { IntegrationService } = await import('@/services');
        await IntegrationService.UnlinkProject(projectStore.activeProject.uri);
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

      if (!projectStore.activeProject?.uri || !tokenData?.token) {
        throw new Error('Not ready to sync');
      }

      this.isLoading = true;
      try {
        const { IntegrationService } = await import('@/services');
        const preview = await IntegrationService.GetSyncPreview(
          projectStore.activeProject.uri,
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

      if (!projectStore.activeProject?.uri || !tokenData?.token) {
        throw new Error('Not ready to sync');
      }

      this.isSyncing = true;
      try {
        const { IntegrationService } = await import('@/services');
        await IntegrationService.ExecuteSync(
          projectStore.activeProject.uri,
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

    // Save tokens to user settings 
    async saveTokens() {
      try {
        const { SettingsService } = await import('@/services');
        for (const [integrationId, tokenData] of Object.entries(this.tokens)) {
          await SettingsService.SaveIntegrationCredential({
            integration_id: integrationId,
            user_id: tokenData.userId || '',
            user_name: tokenData.userName || '',
            user_email: tokenData.userEmail || '',
            access_token: tokenData.token || '',
            refresh_token: tokenData.refreshToken || '',
            expires_at: tokenData.expiresAt || 0,
            api_url: tokenData.apiUrl || '',
            created_at: tokenData.createdAt || new Date().toISOString(),
            updated_at: new Date().toISOString(),
          });
        }
      } catch (error) {
        console.error('Failed to save integration tokens:', error);
      }
    },

    // Load tokens from user settings 
    async loadTokens() {
      try {
        const { SettingsService } = await import('@/services');
        // Load for each known integration
        for (const integration of this.availableIntegrations) {
          try {
            const cred = await SettingsService.GetIntegrationCredential(integration.id);
            if (cred && cred.access_token) {
              this.tokens[integration.id] = {
                token: cred.access_token,
                refreshToken: cred.refresh_token,
                apiUrl: cred.api_url,
                userId: cred.user_id,
                userName: cred.user_name,
                userEmail: cred.user_email,
                expiresAt: cred.expires_at,
                createdAt: cred.created_at,
              };
            }
          } catch (e) {
            // No credentials for this integration - expected
          }
        }
      } catch (error) {
        console.error('Failed to load integration tokens:', error);
        this.tokens = {};
      }
    },

    // Initialize store and load credentials
    async initialize() {
      await this.loadAvailableIntegrations();
      await this.loadTokens();
    },

    // Reset store state (called when project changes)
    // Note: tokens are NOT reset because they are user-level, not project-level
    reset() {
      this.linkedIntegration = null;
      this.syncPreview = null;
      this.isSyncing = false;
    },
  },
});
