package agent

import "testing"

func TestPreviewPresentation(t *testing.T) {
	tests := []struct {
		tool      string
		operation string
		subject   string
	}{
		{tool: "batch_create_assets", operation: previewOperationCreate, subject: "asset"},
		{tool: "batch_create_collections", operation: previewOperationCreate, subject: "collection"},
		{tool: "batch_create_asset_types", operation: previewOperationCreate, subject: "asset type"},
		{tool: "batch_rename", operation: previewOperationUpdate, subject: "item"},
		{tool: "batch_update_collection_types", operation: previewOperationUpdate, subject: "collection type"},
		{tool: "batch_delete", operation: previewOperationDelete, subject: "item"},
		{tool: "delete_asset_type", operation: previewOperationDelete, subject: "asset type"},
		{tool: "dcc_render", operation: previewOperationExecute, subject: "asset"},
		{tool: "add_project_collaborator", operation: previewOperationMembership, subject: "access change"},
	}

	for _, test := range tests {
		t.Run(test.tool, func(t *testing.T) {
			operation, subject := previewPresentation(test.tool)
			if operation != test.operation || subject != test.subject {
				t.Fatalf("previewPresentation(%q) = (%q, %q), want (%q, %q)", test.tool, operation, subject, test.operation, test.subject)
			}
		})
	}
}
