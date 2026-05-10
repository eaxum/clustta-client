package agent

// Risk levels classify tool calls for approval gating.
const (
	RiskSafe        = "safe"        // pure read; no side effects
	RiskMutating    = "mutating"    // writes a single item or small state change
	RiskDestructive = "destructive" // bulk/structural change, deletion, role/membership change, or arbitrary code execution
)

// destructiveTools enumerates tool names that require explicit user approval
// before execution unless auto-approve is enabled.
var destructiveTools = map[string]bool{
	// Single-item destructive
	"delete_asset":      true,
	"delete_collection": true,
	"remove_user":       true,

	// Type/workflow management (mass impact)
	"delete_asset_type":            true,
	"delete_collection_type":       true,
	"apply_workflow":               true,
	"setup_project_types":          true,
	"batch_update_asset_types":     true,
	"batch_update_collection_types": true,

	// Bulk asset operations
	"bulk_delete_assets":     true,
	"bulk_change_asset_type": true,
	"bulk_change_collection_type": true,
	"bulk_change_status":     true,
	"bulk_assign":            true,
	"random_assign":          true,
	"unassign_all_assets":    true,
	"batch_create_assets":    true,
	"batch_create_collections": true,

	// Server-side membership changes
	"add_project_collaborator":       true,
	"remove_project_collaborator":    true,
	"add_studio_collaborator":        true,
	"change_studio_collaborator_role": true,
	"remove_studio_collaborator":     true,

	// Arbitrary code execution
	"run_terminal_command": true,
	"blender_run_script":   true,
	"blender_run_python":   true,
}

// GetToolRisk returns the risk level for the given tool name.
func GetToolRisk(toolName string) string {
	if destructiveTools[toolName] {
		return RiskDestructive
	}
	if readOnlyTools[toolName] {
		return RiskSafe
	}
	return RiskMutating
}

// IsDestructive reports whether a tool requires approval.
func IsDestructive(toolName string) bool {
	return destructiveTools[toolName]
}
