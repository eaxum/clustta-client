import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';
import { globalApiCall } from './http-client.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
}

/**
 * Generate a simple unique ID
 */
function generateId() {
  return crypto.randomUUID ? crypto.randomUUID() : Date.now().toString(36) + Math.random().toString(36).substr(2);
}

export const IntegrationService = {
  // Returns all registered integrations with their info
  GetAvailableIntegrations: async () => {
    try {
      return await globalApiCall('/api/integrations', 'GET') || [];
    } catch {
      return [];
    }
  },

  // Returns info for a specific integration by ID
  GetIntegration: async (integrationId) => {
    try {
      return await globalApiCall(`/api/integrations/${integrationId}`, 'GET') || {};
    } catch {
      return {};
    }
  },

  // Authenticates with an external integration
  Authenticate: async (integrationId, credentials) => {
    return await globalApiCall(`/api/integrations/${integrationId}/auth`, 'POST', credentials);
  },

  // Validates an existing token is still valid
  ValidateToken: async (integrationId, token, apiUrl) => {
    try {
      const result = await globalApiCall(`/api/integrations/${integrationId}/validate`, 'POST', { token, api_url: apiUrl });
      return result?.valid ?? false;
    } catch {
      return false;
    }
  },

  // Returns the integration linked to a project
  GetLinkedIntegration: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    const db = await getDatabase(projectName);
    const row = queryOne(db, 'SELECT * FROM integration_project LIMIT 1');
    return row || {};
  },

  // Fetches available projects from an external integration
  GetExternalProjects: async (integrationId, token, apiUrl) => {
    try {
      return await globalApiCall(`/api/integrations/${integrationId}/projects`, 'POST', { token, api_url: apiUrl }) || [];
    } catch {
      return [];
    }
  },

  // Links a Clustta project to an external project
  LinkProject: async (projectPath, integrationId, externalProjectId, externalProjectName, apiUrl, syncOptions, userId) => {
    const projectName = getProjectName(projectPath);
    const db = await getDatabase(projectName);

    const id = generateId();
    const mtime = Date.now();
    const linkedAt = new Date().toISOString();

    execute(db, `
      INSERT INTO integration_project (id, mtime, integration_id, external_project_id, external_project_name, api_url, sync_options, linked_by_user_id, linked_at, enabled, synced)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 0)
    `, [id, mtime, integrationId, externalProjectId, externalProjectName, apiUrl || '', syncOptions || '{}', userId || '', linkedAt]);

    await persistDatabase(projectName);
    return queryOne(db, 'SELECT * FROM integration_project WHERE id = ?', [id]) || {};
  },

  // Removes the integration link from a project and deletes all mappings
  UnlinkProject: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    const db = await getDatabase(projectName);

    execute(db, 'DELETE FROM integration_asset_mapping');
    execute(db, 'DELETE FROM integration_collection_mapping');
    execute(db, 'DELETE FROM integration_project');

    await persistDatabase(projectName);
  },

  // Fetches external hierarchy and compares with local state
  GetSyncPreview: async (projectPath, token) => {
    const projectName = getProjectName(projectPath);
    const db = await getDatabase(projectName);

    const integrationProject = queryOne(db, 'SELECT * FROM integration_project LIMIT 1');
    if (!integrationProject) {
      throw new Error('No integration linked to this project');
    }

    try {
      const preview = await globalApiCall(`/api/integrations/${integrationProject.integration_id}/sync-preview`, 'POST', {
        token,
        api_url: integrationProject.api_url,
        external_project_id: integrationProject.external_project_id,
      });

      if (!preview) {
        return { Collections: [], Assets: [] };
      }

      // Enrich preview with local mapping state
      const existingCollections = query(db, 'SELECT * FROM integration_collection_mapping WHERE integration_id = ?', [integrationProject.integration_id]);
      const existingAssets = query(db, 'SELECT * FROM integration_asset_mapping WHERE integration_id = ?', [integrationProject.integration_id]);

      const collectionMap = {};
      for (const m of existingCollections) {
        collectionMap[m.external_id] = m;
      }
      const assetMap = {};
      for (const m of existingAssets) {
        assetMap[m.external_id] = m;
      }

      // Update actions based on local state
      if (preview.Collections) {
        for (const c of preview.Collections) {
          if (collectionMap[c.ExternalID]) {
            c.CollectionID = collectionMap[c.ExternalID].collection_id;
            c.Action = collectionMap[c.ExternalID].external_name !== c.ExternalName ? 'update' : 'unchanged';
          } else {
            c.Action = 'create';
          }
        }
      }

      if (preview.Assets) {
        for (const a of preview.Assets) {
          if (assetMap[a.ExternalID]) {
            a.AssetID = assetMap[a.ExternalID].asset_id;
            a.Action = assetMap[a.ExternalID].external_name !== a.ExternalName ? 'update' : 'unchanged';
          } else {
            a.Action = 'create';
          }
        }
      }

      return preview;
    } catch {
      return { Collections: [], Assets: [] };
    }
  },

  // Stores mappings for selected external items
  ExecuteSync: async (projectPath, token, selectedCollections, selectedAssets) => {
    const projectName = getProjectName(projectPath);
    const db = await getDatabase(projectName);

    const integrationProject = queryOne(db, 'SELECT * FROM integration_project LIMIT 1');
    if (!integrationProject) {
      throw new Error('No integration linked to this project');
    }

    const syncedAt = new Date().toISOString();
    const mtime = Date.now();

    // Create collection mappings for selected items
    if (selectedCollections?.length) {
      for (const externalId of selectedCollections) {
        const existing = queryOne(db, 'SELECT id FROM integration_collection_mapping WHERE integration_id = ? AND external_id = ?',
          [integrationProject.integration_id, externalId]);
        if (existing) continue;

        const id = generateId();
        execute(db, `
          INSERT INTO integration_collection_mapping (id, mtime, integration_id, external_id, external_type, external_name, external_parent_id, external_path, external_metadata, collection_id, synced_at, synced)
          VALUES (?, ?, ?, ?, '', '', '', '', '{}', '', ?, 0)
        `, [id, mtime, integrationProject.integration_id, externalId, syncedAt]);
      }
    }

    // Create asset mappings for selected items
    if (selectedAssets?.length) {
      for (const externalId of selectedAssets) {
        const existing = queryOne(db, 'SELECT id FROM integration_asset_mapping WHERE integration_id = ? AND external_id = ?',
          [integrationProject.integration_id, externalId]);
        if (existing) continue;

        const id = generateId();
        execute(db, `
          INSERT INTO integration_asset_mapping (id, mtime, integration_id, external_id, external_name, external_parent_id, external_type, external_status, external_assignees, external_metadata, asset_id, synced_at, synced)
          VALUES (?, ?, ?, ?, '', '', '', '', '[]', '{}', '', ?, 0)
        `, [id, mtime, integrationProject.integration_id, externalId, syncedAt]);
      }
    }

    await persistDatabase(projectName);
  },

  // Returns all collection mappings for the project
  GetCollectionMappings: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    const db = await getDatabase(projectName);

    const integrationProject = queryOne(db, 'SELECT * FROM integration_project LIMIT 1');
    if (!integrationProject) return [];

    return query(db, 'SELECT * FROM integration_collection_mapping WHERE integration_id = ?', [integrationProject.integration_id]);
  },

  // Returns all asset mappings for the project
  GetAssetMappings: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    const db = await getDatabase(projectName);

    const integrationProject = queryOne(db, 'SELECT * FROM integration_project LIMIT 1');
    if (!integrationProject) return [];

    return query(db, 'SELECT * FROM integration_asset_mapping WHERE integration_id = ?', [integrationProject.integration_id]);
  },

  // Returns the external info for a synced asset
  GetAssetExternalInfo: async (projectPath, assetId) => {
    const projectName = getProjectName(projectPath);
    const db = await getDatabase(projectName);
    return queryOne(db, 'SELECT * FROM integration_asset_mapping WHERE asset_id = ?', [assetId]) || {};
  },

  // Returns the external info for a synced collection
  GetCollectionExternalInfo: async (projectPath, collectionId) => {
    const projectName = getProjectName(projectPath);
    const db = await getDatabase(projectName);
    return queryOne(db, 'SELECT * FROM integration_collection_mapping WHERE collection_id = ?', [collectionId]) || {};
  },
};
