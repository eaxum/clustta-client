package agent

import (
	agentcommands "clustta/internal/agent/commands"
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToolSurfaceIsUniqueAndDeterministic(t *testing.T) {
	definitions := GetToolDefinitions()
	names := make([]string, 0, len(definitions))
	seen := map[string]bool{}
	for _, definition := range definitions {
		require.False(t, seen[definition.Name], definition.Name)
		seen[definition.Name] = true
		names = append(names, definition.Name)
	}
	require.True(t, sort.StringsAreSorted(names))
}

func TestBatchMutationsArePlannedAndRequireApproval(t *testing.T) {
	commands := []string{
		"batch_rename", "batch_change_status", "batch_change_type", "batch_assign", "batch_unassign",
		"batch_distribute", "batch_move", "batch_add_tags", "batch_remove_tags", "batch_toggle_task_resource",
		"batch_add_dependency", "batch_remove_dependency", "batch_delete",
		"batch_create_collections", "batch_create_assets", "batch_create_asset_types", "batch_create_collection_types",
		"batch_update_asset_types", "batch_update_collection_types", "setup_project_types",
		"setup_animation_production", "apply_workflow", "dcc_open", "dcc_render", "dcc_export",
		"dcc_run_script", "dcc_run_python", "dcc_set_settings", "dcc_link_dependencies",
	}

	for _, name := range commands {
		t.Run(name, func(t *testing.T) {
			definition, exists := agentcommands.DefinitionFor(name)
			require.True(t, exists)
			require.NotNil(t, definition.Plan)
			require.Nil(t, definition.Direct)
			require.Equal(t, RiskDestructive, definition.Risk)
			require.True(t, IsDestructive(name))
		})
	}
}

func TestToolSurfaceContainsOneCanonicalBatchTool(t *testing.T) {
	counts := map[string]int{}
	for _, definition := range GetToolDefinitions() {
		counts[definition.Name]++
	}
	canonical := []string{
		"batch_create_collections", "batch_create_assets", "batch_create_asset_types",
		"batch_create_collection_types", "batch_update_asset_types", "batch_update_collection_types",
		"setup_project_types", "setup_animation_production", "apply_workflow", "batch_distribute",
	}
	for _, name := range canonical {
		require.Equal(t, 1, counts[name], name)
	}
	retired := []string{
		"list_collections", "list_assets_in_collection",
		"create_asset_type", "create_collection_type", "create_collection", "create_asset",
		"rename_asset", "rename_collection", "change_asset_status", "change_asset_type", "change_collection_type",
		"assign_asset", "unassign_asset", "move_assets", "delete_asset", "delete_collection",
		"add_tag_to_asset", "remove_tag_from_asset", "add_dependency", "remove_dependency",
		"update_asset_type", "update_collection_type",
		"bulk_delete_assets", "bulk_change_status", "bulk_assign", "bulk_change_asset_type",
		"bulk_change_collection_type", "random_assign", "unassign_all_assets",
		"open_in_dcc", "blender_render", "blender_export", "blender_run_script", "blender_run_python",
		"blender_set_settings", "blender_link", "run_terminal_command",
	}
	for _, name := range retired {
		require.Zero(t, counts[name], name)
	}
}

func TestUnsupportedOperationsAreNotExposed(t *testing.T) {
	counts := map[string]int{}
	for _, definition := range GetToolDefinitions() {
		counts[definition.Name]++
	}
	unsupported := []string{
		"track_entities", "batch_track_entities", "inspect_entity", "squash_assets",
		"create_checkpoint", "batch_create_checkpoints", "revert_checkpoint", "delete_checkpoint",
	}
	for _, name := range unsupported {
		require.Zero(t, counts[name], name)
	}
}

func TestSelectBatchItemsKeepsOnlyApprovedInputs(t *testing.T) {
	args := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"name": "First"},
			map[string]interface{}{"name": "Second"},
		},
	}
	approved := planning.Plan{Changes: []planning.Change{
		{Entity: scope.Entity{Type: scope.TypeAsset, ID: "first", Metadata: map[string]interface{}{"input_index": 0}}, Valid: true},
		{Entity: scope.Entity{Type: scope.TypeAsset, ID: "second", Metadata: map[string]interface{}{"input_index": 1}}, Valid: true},
	}}

	err := selectBatchItems("items")(args, approved, []string{"asset:second"})

	require.NoError(t, err)
	items, err := batchObjectItems(args["items"])
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "Second", items[0]["name"])
}

func TestPreserveSelectedDistributionKeepsPreviewedAssignments(t *testing.T) {
	args := map[string]interface{}{"user_ids": []interface{}{"ada", "tobi"}}
	approved := planning.Plan{Changes: []planning.Change{
		{Entity: scope.Entity{Type: scope.TypeAsset, ID: "first", Name: "First"}, Valid: true, After: map[string]interface{}{"assignee_id": "ada"}},
		{Entity: scope.Entity{Type: scope.TypeAsset, ID: "second", Name: "Second"}, Valid: true, After: map[string]interface{}{"assignee_id": "tobi"}},
	}}

	err := preserveSelectedDistribution(args, approved, []string{"asset:second"})

	require.NoError(t, err)
	assignments, ok := args["_assignments"].(map[string]string)
	require.True(t, ok)
	require.Equal(t, map[string]string{"second": "tobi"}, assignments)
	request, ok := args["scope"].(scope.Request)
	require.True(t, ok)
	require.Len(t, request.Selection, 1)
	require.Equal(t, "second", request.Selection[0].ID)
}
