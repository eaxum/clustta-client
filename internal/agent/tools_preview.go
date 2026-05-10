package agent

import (
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"
	"strings"
)

// ToolPreview is a structured description of what a destructive tool would do,
// shown to the user in the approval modal.
type ToolPreview struct {
	Summary string                 `json:"summary"`          // one-line human-readable description
	Items   []map[string]string    `json:"items,omitempty"`  // affected items (name + optional context)
	Counts  map[string]int         `json:"counts,omitempty"` // category → count (e.g. "assets": 12)
	Notes   []string               `json:"notes,omitempty"`  // extra warnings or context lines
	Args    map[string]interface{} `json:"args,omitempty"`   // raw arguments for transparency
}

// buildToolPreview returns a ToolPreview describing the side effects of the given tool call.
// It performs read-only DB queries when useful; on any error it falls back to a generic preview.
func buildToolPreview(projectPath, toolName string, args map[string]interface{}) ToolPreview {
	preview := ToolPreview{Args: args}

	switch toolName {
	case "delete_asset":
		id, _ := args["asset_id"].(string)
		preview.Summary = "Permanently delete one asset and its checkpoints."
		if name := lookupAssetName(projectPath, id); name != "" {
			preview.Items = []map[string]string{{"name": name, "id": id}}
		}
	case "delete_collection":
		id, _ := args["collection_id"].(string)
		preview.Summary = "Permanently delete this collection. Child collections and assets inside it will also be removed."
		if name := lookupCollectionName(projectPath, id); name != "" {
			preview.Items = []map[string]string{{"name": name, "id": id}}
		}
	case "bulk_delete_assets":
		matched := previewBulkDelete(projectPath, args)
		preview.Summary = fmt.Sprintf("Permanently delete %d matching asset(s).", len(matched))
		preview.Counts = map[string]int{"assets": len(matched)}
		preview.Items = matched
		if len(matched) == 0 {
			preview.Notes = append(preview.Notes, "No assets currently match these filters.")
		}
	case "apply_workflow":
		wfId, _ := args["workflow_id"].(string)
		target, _ := args["entity_id"].(string)
		preview.Summary = "Instantiate a workflow under the target entity. This may create many collections and assets at once."
		preview.Notes = append(preview.Notes, fmt.Sprintf("workflow_id=%s, entity_id=%s", wfId, target))
	case "setup_project_types":
		preview.Summary = "Create project-wide asset/collection types in bulk."
	case "batch_update_asset_types", "batch_update_collection_types":
		preview.Summary = "Update many type definitions in a single batch. This affects all items currently assigned to those types."
	case "bulk_change_asset_type", "bulk_change_status", "bulk_assign", "random_assign", "unassign_all_assets":
		preview.Summary = "Modify many assets in one operation."
	case "batch_create_assets", "batch_create_collections":
		preview.Summary = "Create many items at once."
	case "delete_asset_type", "delete_collection_type":
		preview.Summary = "Delete a type definition. Items currently using this type may be affected."
	case "remove_user":
		uid, _ := args["user_id"].(string)
		preview.Summary = "Remove a user from the local project database."
		preview.Items = []map[string]string{{"user_id": uid}}
	case "add_project_collaborator":
		preview.Summary = "Invite a user to the project's remote (server-side membership change)."
	case "remove_project_collaborator":
		preview.Summary = "Revoke a user's server-side access to this project."
	case "add_studio_collaborator", "change_studio_collaborator_role":
		preview.Summary = "Modify studio-level membership."
	case "remove_studio_collaborator":
		preview.Summary = "Remove a user from the studio. They lose access to all studio projects."
	case "run_terminal_command":
		cmd, _ := args["command"].(string)
		preview.Summary = "Execute an arbitrary shell command on this machine."
		preview.Notes = append(preview.Notes, cmd)
	case "blender_run_script", "blender_run_python":
		preview.Summary = "Execute arbitrary Blender Python code. This can modify or overwrite files."
	default:
		preview.Summary = "Run destructive tool: " + toolName
	}

	return preview
}

// lookupAssetName returns the asset's name or empty string on any error.
func lookupAssetName(projectPath, id string) string {
	if id == "" {
		return ""
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ""
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ""
	}
	defer tx.Rollback()
	asset, err := repository.GetAsset(tx, id)
	if err != nil {
		return ""
	}
	return asset.Name
}

// lookupCollectionName returns the collection's name or empty string on any error.
func lookupCollectionName(projectPath, id string) string {
	if id == "" {
		return ""
	}
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return ""
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return ""
	}
	defer tx.Rollback()
	col, err := repository.GetCollection(tx, id)
	if err != nil {
		return ""
	}
	return col.Name
}

// previewBulkDelete returns the names of assets that would match the bulk_delete_assets filters.
// On any error it returns an empty slice.
func previewBulkDelete(projectPath string, args map[string]interface{}) []map[string]string {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return nil
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return nil
	}
	defer tx.Rollback()

	assets, err := repository.GetAssets(tx, false)
	if err != nil {
		return nil
	}

	nameFilter, _ := args["name_contains"].(string)
	statusFilter, _ := args["status"].(string)
	typeFilter, _ := args["asset_type"].(string)
	if typeFilter == "" {
		typeFilter, _ = args["task_type"].(string)
	}
	collectionId, _ := args["collection_id"].(string)
	nameFilter = strings.ToLower(strings.TrimSpace(nameFilter))

	out := []map[string]string{}
	for _, a := range assets {
		if nameFilter != "" && !strings.Contains(strings.ToLower(a.Name), nameFilter) {
			continue
		}
		if statusFilter != "" && !strings.EqualFold(a.StatusShortName, statusFilter) {
			continue
		}
		if typeFilter != "" && !strings.EqualFold(a.AssetTypeName, typeFilter) {
			continue
		}
		if collectionId != "" && a.CollectionId != collectionId {
			continue
		}
		out = append(out, map[string]string{"name": a.Name, "id": a.Id})
		if len(out) >= 100 { // cap preview size
			break
		}
	}
	return out
}

// verifyToolPreview re-checks that the target items still match the preview the
// user was shown in the approval modal. If a name has changed, an item has been
// deleted, or the matched set differs, returns a non-nil error and the agent
// loop treats this as a denial. Closes the TOCTOU window between approval and
// execution for the highest-blast-radius tools.
func verifyToolPreview(projectPath, toolName string, args map[string]interface{}, preview ToolPreview) error {
	switch toolName {
	case "delete_asset":
		id, _ := args["asset_id"].(string)
		if id == "" || len(preview.Items) == 0 {
			return nil
		}
		current := lookupAssetName(projectPath, id)
		if current == "" {
			return fmt.Errorf("asset no longer exists; aborting to keep approval honest")
		}
		expected := preview.Items[0]["name"]
		if expected != "" && current != expected {
			return fmt.Errorf("asset was renamed since you approved (%q → %q); please retry", expected, current)
		}
	case "delete_collection":
		id, _ := args["collection_id"].(string)
		if id == "" || len(preview.Items) == 0 {
			return nil
		}
		current := lookupCollectionName(projectPath, id)
		if current == "" {
			return fmt.Errorf("collection no longer exists; aborting to keep approval honest")
		}
		expected := preview.Items[0]["name"]
		if expected != "" && current != expected {
			return fmt.Errorf("collection was renamed since you approved (%q → %q); please retry", expected, current)
		}
	case "bulk_delete_assets":
		previewIDs := map[string]bool{}
		for _, it := range preview.Items {
			if id := it["id"]; id != "" {
				previewIDs[id] = true
			}
		}
		current := previewBulkDelete(projectPath, args)
		currentIDs := map[string]bool{}
		for _, it := range current {
			if id := it["id"]; id != "" {
				currentIDs[id] = true
			}
		}
		// Any change to the matched set invalidates the user's approval.
		for id := range currentIDs {
			if !previewIDs[id] {
				return fmt.Errorf("asset set changed since you approved (new asset matches the filters); please retry")
			}
		}
		for id := range previewIDs {
			if !currentIDs[id] {
				return fmt.Errorf("asset set changed since you approved (an asset was modified or removed); please retry")
			}
		}
	}
	return nil
}
