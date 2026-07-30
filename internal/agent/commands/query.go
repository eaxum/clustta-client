package commands

import (
	"clustta/internal/agent/scope"
	"fmt"
	"os"
	"strings"
)

func init() {
	allTypes := []string{"asset", "collection", "untracked_asset", "untracked_collection"}
	Register(Definition{
		Name:        "query_entities",
		Description: "Resolve and return entities using the common structured scope and filters. Use this for concrete project queries instead of collecting IDs through multiple list calls.",
		Risk:        "safe",
		Parameters: map[string]interface{}{
			"type": "object", "properties": map[string]interface{}{"scope": ScopeSchema(allTypes)},
			"required": []string{"scope"},
		},
		Direct: func(projectPath string, args map[string]interface{}) (interface{}, error) {
			req, err := ParseScope(args, nil)
			if err != nil {
				return nil, err
			}
			return scope.Resolve(projectPath, req)
		},
	})
	Register(Definition{
		Name:        "audit_entities",
		Description: "Audit entities in a structured scope. Supports naming format, missing tracked paths, unassigned assets, and empty collections.",
		Risk:        "safe",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope": ScopeSchema(allTypes),
				"checks": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string", "enum": []string{"naming", "missing", "unassigned", "empty_collection"}},
				},
				"format": map[string]interface{}{"type": "string", "enum": []string{"camelCase", "PascalCase", "snake_case", "kebab-case", "lowercase", "UPPERCASE"}},
			},
			"required": []string{"scope", "checks"},
		},
		Direct: executeAudit,
	})
}

func executeAudit(projectPath string, args map[string]interface{}) (interface{}, error) {
	req, err := ParseScope(args, nil)
	if err != nil {
		return nil, err
	}
	resolved, err := scope.Resolve(projectPath, req)
	if err != nil {
		return nil, err
	}
	checks := stringSliceArg(args["checks"])
	if len(checks) == 0 {
		return nil, fmt.Errorf("checks is required")
	}
	format := stringArg(args, "format")
	type finding struct {
		Entity   scope.Entity `json:"entity"`
		Check    string       `json:"check"`
		Message  string       `json:"message"`
		Expected interface{}  `json:"expected,omitempty"`
	}
	findings := []finding{}
	childCounts := map[string]int{}
	for _, entity := range resolved.Entities {
		if entity.ParentID != "" {
			childCounts[entity.ParentID]++
		}
	}
	for _, entity := range resolved.Entities {
		for _, check := range checks {
			switch check {
			case "naming":
				if format == "" {
					return nil, fmt.Errorf("format is required for naming audit")
				}
				expected, err := formatEntityName(entity.Name, format)
				if err != nil {
					return nil, err
				}
				if expected != entity.Name {
					findings = append(findings, finding{Entity: entity, Check: check, Message: "name does not match " + format, Expected: expected})
				}
			case "missing":
				if entity.Type.Tracked() && entity.Path != "" {
					if _, err := os.Stat(entity.Path); os.IsNotExist(err) {
						findings = append(findings, finding{Entity: entity, Check: check, Message: "tracked path is missing"})
					}
				}
			case "unassigned":
				if entity.Type == scope.TypeAsset && metadataText(entity, "assignee_id") == "" {
					findings = append(findings, finding{Entity: entity, Check: check, Message: "asset is unassigned"})
				}
			case "empty_collection":
				if (entity.Type == scope.TypeCollection || entity.Type == scope.TypeUntrackedCollection) && childCounts[entity.ID] == 0 {
					findings = append(findings, finding{Entity: entity, Check: check, Message: "collection is empty"})
				}
			}
		}
	}
	return map[string]interface{}{
		"scope": resolved.Request, "checked": len(resolved.Entities), "finding_count": len(findings),
		"findings": findings,
	}, nil
}

func stringSliceArg(value interface{}) []string {
	raw, _ := value.([]interface{})
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
			out = append(out, strings.TrimSpace(text))
		}
	}
	return out
}

func metadataText(entity scope.Entity, key string) string {
	value, _ := entity.Metadata[key].(string)
	return value
}
