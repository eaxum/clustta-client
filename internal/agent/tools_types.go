package agent

import (
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"
)

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

// execBulkChangeCollectionType updates many collections to the same type in one transaction.

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
