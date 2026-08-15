package agent

import agentcommands "clustta/internal/agent/commands"

// Risk levels classify tool calls for approval gating.
const (
	RiskSafe        = "safe"        // pure read; no side effects
	RiskMutating    = "mutating"    // writes a single item or small state change
	RiskDestructive = "destructive" // bulk/structural change, deletion, role/membership change, or arbitrary code execution
)

// destructiveTools contains the remaining non-registry operations that require approval.
var destructiveTools = map[string]bool{
	"remove_user": true,

	"delete_asset_type":      true,
	"delete_collection_type": true,

	"add_project_collaborator":        true,
	"remove_project_collaborator":     true,
	"add_studio_collaborator":         true,
	"change_studio_collaborator_role": true,
	"remove_studio_collaborator":      true,
}

// GetToolRisk returns the risk level for the given tool name.
func GetToolRisk(toolName string) string {
	if definition, ok := agentcommands.DefinitionFor(toolName); ok && definition.Risk != "" {
		return definition.Risk
	}
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
	return GetToolRisk(toolName) == RiskDestructive
}
