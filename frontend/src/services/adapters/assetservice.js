export const AssetService = {
  // Returns all assets for an entity
  GetAssets: async (projectPath, entityPath) => [],

  // Returns a specific asset by ID
  GetAsset: async (projectPath, assetId) => ({}),

  // Creates a new asset
  CreateAsset: async (projectPath, asset) => ({}),

  // Updates an existing asset
  UpdateAsset: async (projectPath, asset) => ({}),

  // Deletes an asset
  DeleteAsset: async (projectPath, assetId) => {},

  // Moves an asset to a new path
  MoveAsset: async (projectPath, assetId, newPath) => {},

  // Copies an asset to a new path
  CopyAsset: async (projectPath, assetId, newPath) => {},

  // Returns all asset types for a project
  GetAssetTypes: async (projectPath) => [],

  // Creates a new asset type
  CreateAssetType: async (projectPath, type) => ({}),

  // Updates an existing asset type
  UpdateAssetType: async (projectPath, type) => ({}),

  // Deletes an asset type
  DeleteAssetType: async (projectPath, typeId) => {},
};
