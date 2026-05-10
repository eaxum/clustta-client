package agent

import (
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"
)

// execChangeAssetType updates a single asset's task type.
func execChangeAssetType(projectPath string, args map[string]interface{}) ToolResult {
	assetID := getStringArg(args, "asset_id", "")
	typeID := getStringArg(args, "task_type_id", "")
	if assetID == "" || typeID == "" {
		return ToolResult{Success: false, Error: "asset_id and task_type_id are required"}
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

	if err := repository.ChangeAssetType(tx, assetID, typeID); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"asset_id": assetID, "task_type_id": typeID}}
}

// execBulkChangeAssetType updates many assets to the same task type in one transaction.
func execBulkChangeAssetType(projectPath string, args map[string]interface{}) ToolResult {
	assetIDs := getStringSliceArg(args, "asset_ids")
	typeID := getStringArg(args, "task_type_id", "")
	if len(assetIDs) == 0 || typeID == "" {
		return ToolResult{Success: false, Error: "asset_ids and task_type_id are required"}
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

	for _, id := range assetIDs {
		if err := repository.ChangeAssetType(tx, id, typeID); err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed for %s: %s", id, err.Error())}
		}
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"updated": len(assetIDs)}}
}

// execUpdateAssetType updates an asset type's name/icon.
func execUpdateAssetType(projectPath string, args map[string]interface{}) ToolResult {
	id := getStringArg(args, "id", "")
	name := getStringArg(args, "name", "")
	icon := getStringArg(args, "icon", "")
	if id == "" {
		return ToolResult{Success: false, Error: "id is required"}
	}
	if icon != "" && !validAssetTypeIcons[icon] {
		return ToolResult{Success: false, Error: "invalid icon name"}
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

	updated, err := repository.UpdateAssetType(tx, id, name, icon)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"id": updated.Id, "name": updated.Name, "icon": updated.Icon}}
}

// execUpdateCollectionType updates a collection type's name/icon.
func execUpdateCollectionType(projectPath string, args map[string]interface{}) ToolResult {
	id := getStringArg(args, "id", "")
	name := getStringArg(args, "name", "")
	icon := getStringArg(args, "icon", "")
	if id == "" {
		return ToolResult{Success: false, Error: "id is required"}
	}
	if icon != "" && !validAssetTypeIcons[icon] {
		return ToolResult{Success: false, Error: "invalid icon name"}
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

	updated, err := repository.UpdateCollectionType(tx, id, name, icon)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"id": updated.Id, "name": updated.Name, "icon": updated.Icon}}
}

// execBatchUpdateAssetTypes updates multiple asset types in one transaction.
func execBatchUpdateAssetTypes(projectPath string, args map[string]interface{}) ToolResult {
	items := getObjSliceArg(args, "items")
	if len(items) == 0 {
		return ToolResult{Success: false, Error: "items is required"}
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

	for i, it := range items {
		id, _ := it["id"].(string)
		name, _ := it["name"].(string)
		icon, _ := it["icon"].(string)
		if id == "" {
			return ToolResult{Success: false, Error: fmt.Sprintf("item %d: id is required", i)}
		}
		if icon != "" && !validAssetTypeIcons[icon] {
			return ToolResult{Success: false, Error: fmt.Sprintf("item %d: invalid icon name '%s'", i, icon)}
		}
		if _, err := repository.UpdateAssetType(tx, id, name, icon); err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("item %d (%s): %s", i, id, err.Error())}
		}
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"updated": len(items)}}
}

// execBatchUpdateCollectionTypes updates multiple collection types in one transaction.
func execBatchUpdateCollectionTypes(projectPath string, args map[string]interface{}) ToolResult {
	items := getObjSliceArg(args, "items")
	if len(items) == 0 {
		return ToolResult{Success: false, Error: "items is required"}
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

	for i, it := range items {
		id, _ := it["id"].(string)
		name, _ := it["name"].(string)
		icon, _ := it["icon"].(string)
		if id == "" {
			return ToolResult{Success: false, Error: fmt.Sprintf("item %d: id is required", i)}
		}
		if icon != "" && !validAssetTypeIcons[icon] {
			return ToolResult{Success: false, Error: fmt.Sprintf("item %d: invalid icon name '%s'", i, icon)}
		}
		if _, err := repository.UpdateCollectionType(tx, id, name, icon); err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("item %d (%s): %s", i, id, err.Error())}
		}
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"updated": len(items)}}
}

// execChangeCollectionType updates a single collection's type.
func execChangeCollectionType(projectPath string, args map[string]interface{}) ToolResult {
	collectionID := getStringArg(args, "collection_id", "")
	typeID := getStringArg(args, "collection_type_id", "")
	if collectionID == "" || typeID == "" {
		return ToolResult{Success: false, Error: "collection_id and collection_type_id are required"}
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

	if err := repository.ChangeCollectionType(tx, collectionID, typeID); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"collection_id": collectionID, "collection_type_id": typeID}}
}

// execBulkChangeCollectionType updates many collections to the same type in one transaction.
func execBulkChangeCollectionType(projectPath string, args map[string]interface{}) ToolResult {
	collectionIDs := getStringSliceArg(args, "collection_ids")
	typeID := getStringArg(args, "collection_type_id", "")
	if len(collectionIDs) == 0 || typeID == "" {
		return ToolResult{Success: false, Error: "collection_ids and collection_type_id are required"}
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

	for _, id := range collectionIDs {
		if err := repository.ChangeCollectionType(tx, id, typeID); err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed for %s: %s", id, err.Error())}
		}
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"updated": len(collectionIDs)}}
}

// execBatchCreateAssetTypes creates multiple asset types in one transaction.
func execBatchCreateAssetTypes(projectPath string, args map[string]interface{}) ToolResult {
	items := getObjSliceArg(args, "items")
	if len(items) == 0 {
		return ToolResult{Success: false, Error: "items is required"}
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

	created := []map[string]string{}
	for i, it := range items {
		name, _ := it["name"].(string)
		icon, _ := it["icon"].(string)
		if name == "" || icon == "" {
			return ToolResult{Success: false, Error: fmt.Sprintf("item %d: name and icon are required", i)}
		}
		if !validAssetTypeIcons[icon] {
			return ToolResult{Success: false, Error: fmt.Sprintf("item %d: invalid icon name '%s'", i, icon)}
		}
		t, err := repository.CreateAssetType(tx, "", name, icon)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("item %d (%s): %s", i, name, err.Error())}
		}
		created = append(created, map[string]string{"id": t.Id, "name": t.Name, "icon": t.Icon})
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"created": created, "count": len(created)}}
}

// execBatchCreateCollectionTypes creates multiple collection types in one transaction.
func execBatchCreateCollectionTypes(projectPath string, args map[string]interface{}) ToolResult {
	items := getObjSliceArg(args, "items")
	if len(items) == 0 {
		return ToolResult{Success: false, Error: "items is required"}
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

	created := []map[string]string{}
	for i, it := range items {
		name, _ := it["name"].(string)
		icon, _ := it["icon"].(string)
		if name == "" || icon == "" {
			return ToolResult{Success: false, Error: fmt.Sprintf("item %d: name and icon are required", i)}
		}
		if !validAssetTypeIcons[icon] {
			return ToolResult{Success: false, Error: fmt.Sprintf("item %d: invalid icon name '%s'", i, icon)}
		}
		t, err := repository.CreateCollectionType(tx, "", name, icon)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("item %d (%s): %s", i, name, err.Error())}
		}
		created = append(created, map[string]string{"id": t.Id, "name": t.Name, "icon": t.Icon})
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"created": created, "count": len(created)}}
}
