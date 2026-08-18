package planning

import (
	"clustta/internal/agent/scope"
	"time"
)

type Change struct {
	Key      string                 `json:"key,omitempty"`
	Entity   scope.Entity           `json:"entity"`
	Action   string                 `json:"action"`
	Before   map[string]interface{} `json:"before,omitempty"`
	After    map[string]interface{} `json:"after,omitempty"`
	Valid    bool                   `json:"valid"`
	Warnings []string               `json:"warnings,omitempty"`
	Errors   []string               `json:"errors,omitempty"`
}

type Plan struct {
	ID           string                 `json:"plan_id"`
	Command      string                 `json:"command"`
	Scope        scope.Result           `json:"scope"`
	Changes      []Change               `json:"changes"`
	Counts       map[string]int         `json:"counts"`
	Warnings     []string               `json:"warnings,omitempty"`
	Errors       []string               `json:"errors,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	Fingerprint  string                 `json:"fingerprint"`
	LocalOnly    bool                   `json:"local_only"`
	RequiresSync bool                   `json:"requires_sync"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

func (p Plan) Executable() bool {
	if len(p.Errors) > 0 {
		return false
	}
	executable := 0
	for _, change := range p.Changes {
		if len(change.Errors) > 0 {
			return false
		}
		if change.Valid {
			executable++
		}
	}
	return executable > 0
}

type Result struct {
	PlanID       string                   `json:"plan_id"`
	Command      string                   `json:"command"`
	Applied      int                      `json:"applied"`
	Skipped      int                      `json:"skipped"`
	Failed       int                      `json:"failed"`
	LocalOnly    bool                     `json:"local_only"`
	RequiresSync bool                     `json:"requires_sync"`
	Items        []map[string]interface{} `json:"items,omitempty"`
	Errors       []string                 `json:"errors,omitempty"`
}
