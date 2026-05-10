package agent

import (
	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/utils"
)

// execListWorkflows returns all workflows in the project with their entity/asset counts.
func execListWorkflows(projectPath string) ToolResult {
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

	workflows, err := repository.GetWorkflows(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	type wfSummary struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Collections int    `json:"collection_count"`
		Assets      int    `json:"asset_count"`
		Links       int    `json:"link_count"`
	}
	results := make([]wfSummary, 0, len(workflows))
	for _, w := range workflows {
		results = append(results, wfSummary{
			Id:          w.Id,
			Name:        w.Name,
			Collections: len(w.Collections),
			Assets:      len(w.Assets),
			Links:       len(w.Links),
		})
	}
	return ToolResult{Success: true, Data: results}
}

// execApplyWorkflow instantiates a workflow under the given parent.
func execApplyWorkflow(projectPath string, args map[string]interface{}) ToolResult {
	workflowID := getStringArg(args, "workflow_id", "")
	name := getStringArg(args, "name", "")
	collectionTypeID := getStringArg(args, "collection_type_id", "")
	parentID := getStringArg(args, "parent_id", "")
	if workflowID == "" || name == "" || collectionTypeID == "" {
		return ToolResult{Success: false, Error: "workflow_id, name and collection_type_id are required"}
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return ToolResult{Success: false, Error: "could not determine active user: " + err.Error()}
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

	if err := repository.AddWorkflow(tx, workflowID, name, collectionTypeID, parentID, user); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{"workflow_id": workflowID, "root_collection_name": name}}
}
