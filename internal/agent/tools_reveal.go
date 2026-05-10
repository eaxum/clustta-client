package agent

import (
	"clustta/internal/repository"
	"clustta/internal/utils"
)

// execRevealAssetOnDisk opens the OS file explorer at the asset's path.
func execRevealAssetOnDisk(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	if assetID == "" {
		return ToolResult{Success: false, Error: "asset_id is required"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	asset, err := repository.GetAsset(tx, assetID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	safePath, err := validateAssetPath(projectPath, asset.GetFilePath())
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	utils.RevealInExplorer(safePath)
	return ToolResult{Success: true, Data: map[string]interface{}{"asset_id": asset.Id, "name": asset.Name, "path": safePath}}
}

// execRevealInBrowser returns navigation data for the frontend to focus on an item.
// The agent loop emits the data via the `agent-reveal-in-browser` event.
func execRevealInBrowser(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	collectionID := getStringArg(args, "collection_id", "")
	if assetID == "" && collectionID == "" {
		return ToolResult{Success: false, Error: "either asset_id or collection_id is required"}
	}
	if assetID != "" && collectionID != "" {
		return ToolResult{Success: false, Error: "provide only one of asset_id or collection_id"}
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer tx.Rollback()

	if assetID != "" {
		asset, err := repository.GetAsset(tx, assetID)
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		return ToolResult{Success: true, Data: map[string]interface{}{
			"kind":          "asset",
			"asset_id":      asset.Id,
			"collection_id": asset.CollectionId,
			"name":          asset.Name,
		}}
	}

	collection, err := repository.GetCollection(tx, collectionID)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{
		"kind":          "collection",
		"collection_id": collection.Id,
		"name":          collection.Name,
	}}
}
