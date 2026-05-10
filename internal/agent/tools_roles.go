package agent

import (
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
)

// execListRoles returns all project roles with their permission attributes.
func execListRoles(projectPath string) ToolResult {
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

	roles, err := repository.GetRoles(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: roles}
}

// execChangeCollaboratorRole changes a project user's role by role name.
func execChangeCollaboratorRole(projectPath string, args map[string]interface{}) ToolResult {
	userID := getStringArg(args, "user_id", "")
	roleName := getStringArg(args, "role_name", "")
	if userID == "" || roleName == "" {
		return ToolResult{Success: false, Error: "user_id and role_name are required"}
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

	if err := repository.ChangeUserRoleByName(tx, userID, roleName); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"user_id": userID, "role_name": roleName}}
}

// roleAttributesFromMap converts an arguments map into a RoleAttributes struct.
// Missing keys default to false (the underlying API requires every flag set explicitly).
func roleAttributesFromMap(m map[string]interface{}) models.RoleAttributes {
	b := func(key string) bool {
		v, ok := m[key]
		if !ok {
			return false
		}
		bv, ok := v.(bool)
		return ok && bv
	}
	return models.RoleAttributes{
		ViewCollection:     b("view_collection"),
		CreateCollection:   b("create_collection"),
		UpdateCollection:   b("update_collection"),
		DeleteCollection:   b("delete_collection"),
		ViewAsset:          b("view_asset"),
		CreateAsset:        b("create_asset"),
		UpdateAsset:        b("update_asset"),
		DeleteAsset:        b("delete_asset"),
		ViewTemplate:       b("view_template"),
		CreateTemplate:     b("create_template"),
		UpdateTemplate:     b("update_template"),
		DeleteTemplate:     b("delete_template"),
		ViewCheckpoint:     b("view_checkpoint"),
		CreateCheckpoint:   b("create_checkpoint"),
		DeleteCheckpoint:   b("delete_checkpoint"),
		PullChunk:          b("pull_chunk"),
		AssignAsset:        b("assign_asset"),
		UnassignAsset:      b("unassign_asset"),
		AddUser:            b("add_user"),
		RemoveUser:         b("remove_user"),
		ChangeRole:         b("change_role"),
		ChangeStatus:       b("change_status"),
		SetDoneAsset:       b("set_done_asset"),
		SetRetakeAsset:     b("set_retake_asset"),
		ViewDoneAsset:      b("view_done_asset"),
		ManageDependencies: b("manage_dependencies"),
		ManageShareLinks:   b("manage_share_links"),
	}
}

// execUpdateRole updates a role's name and full permission set.
func execUpdateRole(projectPath string, args map[string]interface{}) ToolResult {
	id := getStringArg(args, "id", "")
	name := getStringArg(args, "name", "")
	if id == "" || name == "" {
		return ToolResult{Success: false, Error: "id and name are required"}
	}
	attrsRaw, _ := args["attributes"].(map[string]interface{})
	if attrsRaw == nil {
		return ToolResult{Success: false, Error: "attributes object is required"}
	}
	attrs := roleAttributesFromMap(attrsRaw)

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

	updated, err := repository.UpdateRole(tx, id, name, attrs)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"id": updated.Id, "name": updated.Name}}
}
