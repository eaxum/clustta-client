import { defineStore } from 'pinia';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { IntegrationService, SettingsService } from '@/services';

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
    // Type mapping state
    typeMappings: null,           // Current sync options with type mappings
    missingTypes: [],             // Types that need to be created
    externalEntityTypes: [],      // Entity types from external system
    externalTaskTypes: [],        // Task types from external system
    localEntityTypes: [],         // Clustta entity types
    localTaskTypes: [],           // Clustta task types
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

    // Builds a hierarchical tree from collection_path values (from DirectoryMapping).
    // Creates folder nodes for path segments and places items at their correct locations.
    syncPreviewTree: (state) => {
      if (!state.syncPreview) return [];

      const collections = state.syncPreview.collections || [];
      const assets = state.syncPreview.assets || [];

      // Root node to hold everything
      const root = { children: new Map() };

      // Helper to get or create a folder node at a path
      const getOrCreateFolder = (pathSegments) => {
        let current = root;
        for (const segment of pathSegments) {
          if (!current.children.has(segment)) {
            current.children.set(segment, {
              id: `folder-${pathSegments.slice(0, pathSegments.indexOf(segment) + 1).join('/')}`,
              type: 'entity',
              name: segment,
              entity_type_name: 'Folder',
              entity_type_icon: 'folder',
              external_type: 'folder',
              action: 'virtual',
              children: new Map(),
            });
          }
          current = current.children.get(segment);
        }
        return current;
      };

      // Place collections into tree based on collection_path
      collections.forEach(c => {
        const path = c.collection_path || c.external_name;
        const segments = path.split('/').filter(s => s);
        
        if (segments.length === 0) return;

        // Parent path is all segments except the last one
        const parentSegments = segments.slice(0, -1);
        const parent = parentSegments.length > 0 ? getOrCreateFolder(parentSegments) : root;

        // Use last segment as key (matches getOrCreateFolder's segment-based lookup)
        const key = segments[segments.length - 1];

        // Check if a virtual folder already exists at this key (from child paths processed earlier)
        const existingNode = parent.children.get(key);
        
        // Create the collection node, preserving children from any existing virtual folder
        const collectionNode = {
          id: c.external_id,
          type: 'entity',
          name: c.external_name,
          entity_type_name: c.entity_type_name,
          entity_type_icon: c.entity_type_icon || 'folder',
          external_id: c.external_id,
          external_type: c.external_type,
          collection_path: c.collection_path,
          action: c.action,
          children: existingNode?.children instanceof Map ? existingNode.children : new Map(),
        };

        parent.children.set(key, collectionNode);
      });

      // Place assets into tree based on their collection_path (parent folder)
      assets.forEach(a => {
        const parentPath = a.collection_path || '';
        const parentSegments = parentPath.split('/').filter(s => s);
        const parent = parentSegments.length > 0 ? getOrCreateFolder(parentSegments) : root;

        const assetNode = {
          id: a.external_id,
          type: 'task',
          name: a.external_name,
          task_type_name: a.task_type_name,
          task_type_icon: a.task_type_icon || 'generic',
          external_id: a.external_id,
          external_type: a.external_type,
          external_type_id: a.external_type_id,
          collection_path: a.collection_path,
          action: a.action,
          children: new Map(),
        };

        // Use external_id as key for assets
        parent.children.set(`asset-${a.external_id}`, assetNode);
      });

      // Convert Map children to arrays recursively
      const convertToArray = (node) => {
        if (node.children instanceof Map) {
          node.children = Array.from(node.children.values()).map(convertToArray);
        }
        return node;
      };

      return Array.from(root.children.values()).map(convertToArray);
    },

    // Missing types from sync preview
    missingTypesFromPreview: (state) => {
      if (!state.syncPreview) return [];
      return state.syncPreview.missing_types || [];
    },

    // Check if there are missing types
    hasMissingTypes: (state) => {
      if (!state.syncPreview) return false;
      return (state.syncPreview.missing_types || []).length > 0;
    },

    // Count of missing entity types
    missingEntityTypesCount: (state) => {
      if (!state.syncPreview?.missing_types) return 0;
      return state.syncPreview.missing_types.filter(t => t.type_category === 'entity').length;
    },

    // Count of missing task types
    missingTaskTypesCount: (state) => {
      if (!state.syncPreview?.missing_types) return 0;
      return state.syncPreview.missing_types.filter(t => t.type_category === 'task').length;
    },

    // Entity type mappings from sync options
    entityTypeMappings: (state) => {
      return state.typeMappings?.entity_type_mappings || {};
    },

    // Task type mappings from sync options
    taskTypeMappings: (state) => {
      return state.typeMappings?.task_type_mappings || {};
    },
  },

  actions: {
    // Load available integrations from backend
    async loadAvailableIntegrations() {
      try {
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
          notificationStore.addNotification(result.error || 'Authentication failed', 'error');
          return { success: false, error: result.error };
        }
      } catch (error) {
        notificationStore.addNotification(error.message || 'Authentication failed', 'error');
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
      await SettingsService.DeleteIntegrationCredential(integrationId);
    },

    // Get external projects for an integration
    async getExternalProjects(integrationId) {
      const tokenData = this.tokens[integrationId];
      if (!tokenData?.token) {
        throw new Error('Not authenticated with ' + integrationId);
      }

      try {
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
        notificationStore.addNotification('Project linked to ' + externalProjectName, 'success');
        return result;
      } catch (error) {
        notificationStore.addNotification(error.message || 'Failed to link project', 'error');
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
        await IntegrationService.UnlinkProject(projectStore.activeProject.uri);
        this.linkedIntegration = null;
        this.syncPreview = null;
        notificationStore.addNotification('Integration unlinked', 'success');
      } catch (error) {
        notificationStore.addNotification(error.message || 'Failed to unlink project', 'error');
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
        await IntegrationService.ExecuteSync(
          projectStore.activeProject.uri,
          tokenData.token,
          selectedCollections,
          selectedAssets
        );
        this.lastSyncAt = new Date().toISOString();
        notificationStore.addNotification('Sync completed successfully', 'success');
      } catch (error) {
        notificationStore.addNotification(error.message || 'Sync failed', 'error');
        throw error;
      } finally {
        this.isSyncing = false;
      }
    },

    // Save tokens to user settings 
    async saveTokens() {
      try {
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

    // Load type mappings from backend
    async loadTypeMappings() {
      const projectStore = useProjectStore();
      if (!projectStore.activeProject?.uri) {
        this.typeMappings = null;
        return;
      }

      try {
        const mappings = await IntegrationService.GetTypeMappings(projectStore.activeProject.uri);
        this.typeMappings = mappings;
        return mappings;
      } catch (error) {
        console.error('Failed to load type mappings:', error);
        this.typeMappings = null;
      }
    },

    // Save type mappings to backend
    async saveTypeMappings(syncOptions) {
      const projectStore = useProjectStore();
      const notificationStore = useNotificationStore();

      if (!projectStore.activeProject?.uri) {
        throw new Error('No active project');
      }

      try {
        await IntegrationService.SaveTypeMappings(projectStore.activeProject.uri, syncOptions);
        this.typeMappings = syncOptions;
        notificationStore.addNotification('Type mappings saved', 'success');
      } catch (error) {
        notificationStore.addNotification(error.message || 'Failed to save type mappings', 'error');
        throw error;
      }
    },

    // Save directory structure to backend (merges with existing sync options)
    async saveDirectoryStructure(directoryStructure) {
      const projectStore = useProjectStore();

      if (!projectStore.activeProject?.uri) {
        throw new Error('No active project');
      }

      // Merge with existing type mappings
      const syncOptions = {
        ...this.typeMappings,
        directory_structure: directoryStructure,
      };

      try {
        await IntegrationService.SaveTypeMappings(projectStore.activeProject.uri, syncOptions);
        this.typeMappings = syncOptions;
      } catch (error) {
        throw error;
      }
    },

    // Save task type templates to backend (merges with existing sync options)
    async saveTaskTypeTemplates(taskTypeTemplates) {
      const projectStore = useProjectStore();

      if (!projectStore.activeProject?.uri) {
        throw new Error('No active project');
      }

      // Merge with existing type mappings
      const syncOptions = {
        ...this.typeMappings,
        task_type_templates: taskTypeTemplates,
      };

      try {
        await IntegrationService.SaveTypeMappings(projectStore.activeProject.uri, syncOptions);
        this.typeMappings = syncOptions;
      } catch (error) {
        throw error;
      }
    },

    // Fetch external types from integration
    async getExternalTypes() {
      const projectStore = useProjectStore();
      const tokenData = this.tokens[this.linkedIntegration?.integration_id];

      if (!projectStore.activeProject?.uri || !tokenData?.token) {
        throw new Error('Not ready to fetch types');
      }

      try {
        const result = await IntegrationService.GetExternalTypes(
          projectStore.activeProject.uri,
          tokenData.token
        );
        // Result is [entityTypes, taskTypes]
        this.externalEntityTypes = result[0] || [];
        this.externalTaskTypes = result[1] || [];
        return { entityTypes: this.externalEntityTypes, taskTypes: this.externalTaskTypes };
      } catch (error) {
        console.error('Failed to get external types:', error);
        throw error;
      }
    },

    // Get missing types that need to be created
    async getMissingTypes() {
      const projectStore = useProjectStore();
      const tokenData = this.tokens[this.linkedIntegration?.integration_id];

      if (!projectStore.activeProject?.uri || !tokenData?.token) {
        throw new Error('Not ready to check types');
      }

      this.isLoading = true;
      try {
        const missing = await IntegrationService.GetMissingTypes(
          projectStore.activeProject.uri,
          tokenData.token
        );
        this.missingTypes = missing || [];
        return this.missingTypes;
      } catch (error) {
        console.error('Failed to get missing types:', error);
        throw error;
      } finally {
        this.isLoading = false;
      }
    },

    // Create missing types in Clustta
    async createMissingTypes(typesToCreate) {
      const projectStore = useProjectStore();
      const notificationStore = useNotificationStore();

      if (!projectStore.activeProject?.uri) {
        throw new Error('No active project');
      }

      this.isLoading = true;
      try {
        await IntegrationService.CreateMissingTypes(projectStore.activeProject.uri, typesToCreate);
        // Reload type mappings after creation
        await this.loadTypeMappings();
        // Reload local types
        await this.getLocalTypes();
        this.missingTypes = [];
        notificationStore.addNotification(`Created ${typesToCreate.length} type(s)`, 'success');
      } catch (error) {
        notificationStore.addNotification(error.message || 'Failed to create types', 'error');
        throw error;
      } finally {
        this.isLoading = false;
      }
    },

    // Load local Clustta types
    async getLocalTypes() {
      const projectStore = useProjectStore();

      if (!projectStore.activeProject?.uri) {
        this.localEntityTypes = [];
        this.localTaskTypes = [];
        return;
      }

      try {
        const result = await IntegrationService.GetLocalTypes(projectStore.activeProject.uri);
        // Result is [entityTypes, taskTypes]
        this.localEntityTypes = result[0] || [];
        this.localTaskTypes = result[1] || [];
        return { entityTypes: this.localEntityTypes, taskTypes: this.localTaskTypes };
      } catch (error) {
        console.error('Failed to get local types:', error);
        this.localEntityTypes = [];
        this.localTaskTypes = [];
      }
    },

    // Reset store state (called when project changes)
    // Note: tokens are NOT reset because they are user-level, not project-level
    reset() {
      this.linkedIntegration = null;
      this.syncPreview = null;
      this.isSyncing = false;
      this.typeMappings = null;
      this.missingTypes = [];
      this.externalEntityTypes = [];
      this.externalTaskTypes = [];
      this.localEntityTypes = [];
      this.localTaskTypes = [];
    },
  },
});
