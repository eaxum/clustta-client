package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"fmt"
	"strings"
	"time"
)

func newPlan(command string, resolved scope.Result) planning.Plan {
	return planning.Plan{
		Command: command, Scope: resolved, Counts: map[string]int{},
		CreatedAt: time.Now().UTC(), LocalOnly: true, RequiresSync: true,
	}
}

func addChange(plan *planning.Plan, change planning.Change) {
	plan.Changes = append(plan.Changes, change)
	if change.Valid && len(change.Errors) == 0 {
		plan.Counts["changes"]++
	} else {
		plan.Counts["invalid"]++
	}
	if len(change.Warnings) > 0 {
		plan.Counts["warnings"] += len(change.Warnings)
	}
}

func stringArg(args map[string]interface{}, key string) string {
	value, _ := args[key].(string)
	return strings.TrimSpace(value)
}

func resultItem(change planning.Change, status string) map[string]interface{} {
	entity := change.Entity
	if name, ok := change.After["name"].(string); ok && name != "" {
		entity.Name = name
	}
	if path, ok := change.After["path"].(string); ok && path != "" {
		entity.Path = path
	}
	if parentID, ok := change.After["parent_id"].(string); ok {
		entity.ParentID = parentID
		if entity.Type == scope.TypeAsset || entity.Type == scope.TypeUntrackedAsset {
			entity.CollectionID = parentID
		}
	}
	return map[string]interface{}{
		"type": entity.Type, "id": entity.ID, "name": entity.Name,
		"entity": entity, "action": change.Action, "before": change.Before, "after": change.After, "status": status,
	}
}

func requireTracked(change *planning.Change) {
	if !change.Entity.Type.Tracked() {
		change.Valid = false
		change.Errors = append(change.Errors, fmt.Sprintf("%s does not support this operation", change.Entity.Type))
	}
}
