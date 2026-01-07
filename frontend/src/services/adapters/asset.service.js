// =============================================================================
// ASSET SERVICE
// =============================================================================

export const AssetService = {
  // These would need studio server endpoints
  GetAssets: async (projectPath, entityPath) => [],
  GetAsset: async (projectPath, assetId) => ({}),
  CreateAsset: async (projectPath, asset) => ({}),
  UpdateAsset: async (projectPath, asset) => ({}),
  DeleteAsset: async (projectPath, assetId) => {},
  MoveAsset: async (projectPath, assetId, newPath) => {},
  CopyAsset: async (projectPath, assetId, newPath) => {},
  GetAssetTypes: async (projectPath) => [],
  CreateAssetType: async (projectPath, type) => ({}),
  UpdateAssetType: async (projectPath, type) => ({}),
  DeleteAssetType: async (projectPath, typeId) => {},
};
