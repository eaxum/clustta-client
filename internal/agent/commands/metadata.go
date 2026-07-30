package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

func init() {
	Register(metadataCommand(
		"batch_change_status",
		"Change the status of tracked assets in a structured scope. Changes are local-only and require manual sync.",
		"status_id", "Change Status", []scope.EntityType{scope.TypeAsset},
		func(tx txLike, entity scope.Entity, value string) error {
			return repository.UpdateStatus(tx.Tx(), entity.ID, value)
		},
	))
	Register(typeCommand())
}

// txLike keeps the metadata command helper focused while using the repository's sqlx transaction.
type txLike interface {
	Tx() *sqlx.Tx
}

type sqlxTx struct{ tx *sqlx.Tx }

func (s sqlxTx) Tx() *sqlx.Tx { return s.tx }

func metadataCommand(name, description, valueKey, action string, types []scope.EntityType, apply func(txLike, scope.Entity, string) error) Definition {
	enums := make([]string, len(types))
	for i, entityType := range types {
		enums[i] = string(entityType)
	}
	return Definition{
		Name: name, Description: description, Permission: permissionForMetadata(name), Risk: "destructive",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope":  ScopeSchema(enums),
				valueKey: map[string]interface{}{"type": "string", "description": "Canonical target ID."},
			},
			"required": []string{"scope", valueKey},
		},
		Plan: func(projectPath string, args map[string]interface{}) (planning.Plan, error) {
			req, err := ParseScope(args, types)
			if err != nil {
				return planning.Plan{}, err
			}
			resolved, err := scope.Resolve(projectPath, req)
			if err != nil {
				return planning.Plan{}, err
			}
			value := stringArg(args, valueKey)
			if value == "" {
				return planning.Plan{}, fmt.Errorf("%s is required", valueKey)
			}
			if err := validateMetadataTarget(projectPath, name, resolved.Entities, value); err != nil {
				return planning.Plan{}, err
			}
			targetLabel, err := metadataTargetLabel(projectPath, name, types[0], value)
			if err != nil {
				return planning.Plan{}, err
			}
			plan := newPlan(name, resolved)
			plan.Options = map[string]interface{}{valueKey: value}
			for _, entity := range resolved.Entities {
				change := planning.Change{Entity: entity, Action: action, Valid: true}
				requireTracked(&change)
				if name == "batch_change_status" && entity.Type != scope.TypeAsset {
					change.Valid = false
					change.Errors = append(change.Errors, "status is only supported for assets")
				}
				beforeKey := valueKey
				if name == "batch_change_type" {
					if entity.Type == scope.TypeAsset {
						beforeKey = "asset_type_id"
					} else {
						beforeKey = "collection_type_id"
					}
				}
				change.Before = map[string]interface{}{beforeKey: entity.Metadata[beforeKey]}
				change.After = map[string]interface{}{beforeKey: value}
				if name == "batch_change_status" {
					change.Before["status"] = entity.Metadata["status"]
					change.After["status"] = targetLabel
				}
				if entity.Metadata[beforeKey] == value {
					change.Valid = false
					change.Warnings = append(change.Warnings, "already has the requested value")
				}
				addChange(&plan, change)
			}
			if len(plan.Changes) == 0 {
				plan.Errors = append(plan.Errors, "scope resolved to no supported entities")
			}
			return plan, nil
		},
		Execute: func(projectPath string, plan planning.Plan) (planning.Result, error) {
			value, _ := plan.Options[valueKey].(string)
			db, err := utils.OpenDb(projectPath)
			if err != nil {
				return planning.Result{}, err
			}
			defer db.Close()
			tx, err := db.Beginx()
			if err != nil {
				return planning.Result{}, err
			}
			defer tx.Rollback()
			result := planning.Result{PlanID: plan.ID, Command: name, LocalOnly: true, RequiresSync: true}
			for _, change := range plan.Changes {
				if !change.Valid {
					result.Skipped++
					result.Items = append(result.Items, resultItem(change, "skipped"))
					continue
				}
				if err := apply(sqlxTx{tx}, change.Entity, value); err != nil {
					return planning.Result{}, fmt.Errorf("%s %s: %w", action, change.Entity.Name, err)
				}
				result.Applied++
				result.Items = append(result.Items, resultItem(change, "applied"))
			}
			if err := tx.Commit(); err != nil {
				return planning.Result{}, err
			}
			return result, nil
		},
	}
}

func typeCommand() Definition {
	return Definition{
		Name:        "batch_change_type",
		Description: "Change asset and/or collection types in one structured batch. Use asset_type_rules to map asset-name suffixes to different type IDs in a single call. Local-only; manual sync required.",
		Permission:  "update_asset", Risk: "destructive",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope":              ScopeSchema([]string{"asset", "collection"}),
				"asset_type_id":      map[string]interface{}{"type": "string"},
				"collection_type_id": map[string]interface{}{"type": "string"},
				"asset_type_rules": map[string]interface{}{
					"type":        "array",
					"description": "Optional suffix-to-type mappings. Resolve the full scope once and apply the matching target to each asset. Prefer this over separate calls per suffix.",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"suffix":        map[string]interface{}{"type": "string"},
							"asset_type_id": map[string]interface{}{"type": "string"},
						},
						"required": []string{"suffix", "asset_type_id"},
					},
				},
			},
			"required": []string{"scope"},
		},
		Plan: func(projectPath string, args map[string]interface{}) (planning.Plan, error) {
			assetTypeID := stringArg(args, "asset_type_id")
			collectionTypeID := stringArg(args, "collection_type_id")
			assetRules, err := parseAssetTypeRules(args["asset_type_rules"])
			if err != nil {
				return planning.Plan{}, err
			}
			if assetTypeID == "" && collectionTypeID == "" && len(assetRules) == 0 {
				return planning.Plan{}, fmt.Errorf("asset_type_id, collection_type_id, or asset_type_rules is required")
			}
			req, err := ParseScope(args, []scope.EntityType{scope.TypeAsset, scope.TypeCollection})
			if err != nil {
				return planning.Plan{}, err
			}
			// A single target type determines the only compatible entity kind.
			// Do not allow an omitted or malformed scope.types field to pull the
			// opposite kind into the plan and invalidate the whole operation.
			if (assetTypeID != "" || len(assetRules) > 0) && collectionTypeID == "" {
				req.Types = []scope.EntityType{scope.TypeAsset}
			} else if collectionTypeID != "" && assetTypeID == "" {
				req.Types = []scope.EntityType{scope.TypeCollection}
			}
			if err := validateScopeFilters(req); err != nil {
				return planning.Plan{}, err
			}
			resolved, err := scope.Resolve(projectPath, req)
			if err != nil {
				return planning.Plan{}, err
			}
			if assetTypeID != "" {
				if err := validateMetadataTarget(projectPath, "batch_change_type", []scope.Entity{{Type: scope.TypeAsset}}, assetTypeID); err != nil {
					return planning.Plan{}, err
				}
			}
			if collectionTypeID != "" {
				if err := validateMetadataTarget(projectPath, "batch_change_type", []scope.Entity{{Type: scope.TypeCollection}}, collectionTypeID); err != nil {
					return planning.Plan{}, err
				}
			}
			assetTypeLabel := ""
			collectionTypeLabel := ""
			assetTypeIcon := ""
			collectionTypeIcon := ""
			if assetTypeID != "" {
				assetTypeLabel, assetTypeIcon, err = metadataTargetType(projectPath, scope.TypeAsset, assetTypeID)
				if err != nil {
					return planning.Plan{}, err
				}
			}
			ruleTargets := make(map[string]typeTarget, len(assetRules))
			for _, rule := range assetRules {
				if _, ok := ruleTargets[rule.AssetTypeID]; ok {
					continue
				}
				label, icon, targetErr := metadataTargetType(projectPath, scope.TypeAsset, rule.AssetTypeID)
				if targetErr != nil {
					return planning.Plan{}, fmt.Errorf("target asset type for suffix %q not found", rule.Suffix)
				}
				ruleTargets[rule.AssetTypeID] = typeTarget{Label: label, Icon: icon}
			}
			if collectionTypeID != "" {
				collectionTypeLabel, collectionTypeIcon, err = metadataTargetType(projectPath, scope.TypeCollection, collectionTypeID)
				if err != nil {
					return planning.Plan{}, err
				}
			}
			plan := newPlan("batch_change_type", resolved)
			plan.Options = map[string]interface{}{"asset_type_id": assetTypeID, "collection_type_id": collectionTypeID}
			for _, entity := range resolved.Entities {
				key, target := "asset_type_id", assetTypeID
				targetLabel, targetIcon := assetTypeLabel, assetTypeIcon
				matchedRule := true
				if entity.Type == scope.TypeCollection {
					key, target = "collection_type_id", collectionTypeID
					targetLabel, targetIcon = collectionTypeLabel, collectionTypeIcon
				} else if target == "" && len(assetRules) > 0 {
					var matched assetTypeRule
					matched, matchedRule = matchAssetTypeRule(entity.Name, assetRules)
					if matchedRule {
						target = matched.AssetTypeID
						targetLabel = ruleTargets[target].Label
						targetIcon = ruleTargets[target].Icon
					}
				}
				labelKey := "asset_type"
				iconKey := "asset_type_icon"
				if entity.Type == scope.TypeCollection {
					labelKey = "collection_type"
					iconKey = "collection_type_icon"
				}
				change := planning.Change{
					Entity: entity, Action: "Change Type", Valid: true,
					Before: map[string]interface{}{key: entity.Metadata[key], labelKey: entity.Metadata[labelKey], iconKey: entity.Metadata[iconKey]},
					After:  map[string]interface{}{key: target, labelKey: targetLabel, iconKey: targetIcon},
				}
				if !matchedRule {
					change.Valid = false
					change.Warnings = append(change.Warnings, "no asset type rule matched the name suffix")
				} else if target == "" {
					change.Valid = false
					change.Errors = append(change.Errors, key+" is required for the resolved scope")
				} else if entity.Metadata[key] == target {
					change.Valid = false
					change.Warnings = append(change.Warnings, "already has the requested type")
				}
				addChange(&plan, change)
			}
			if len(plan.Changes) == 0 {
				plan.Errors = append(plan.Errors, "scope resolved to no supported entities")
			}
			return plan, nil
		},
		Execute: func(projectPath string, plan planning.Plan) (planning.Result, error) {
			db, err := utils.OpenDb(projectPath)
			if err != nil {
				return planning.Result{}, err
			}
			defer db.Close()
			tx, err := db.Beginx()
			if err != nil {
				return planning.Result{}, err
			}
			defer tx.Rollback()
			result := planning.Result{PlanID: plan.ID, Command: plan.Command, LocalOnly: true, RequiresSync: true}
			for _, change := range plan.Changes {
				if !change.Valid {
					result.Skipped++
					result.Items = append(result.Items, resultItem(change, "skipped"))
					continue
				}
				if change.Entity.Type == scope.TypeAsset {
					value, _ := change.After["asset_type_id"].(string)
					err = repository.ChangeAssetType(tx, change.Entity.ID, value)
				} else {
					value, _ := change.After["collection_type_id"].(string)
					err = repository.ChangeCollectionType(tx, change.Entity.ID, value)
				}
				if err != nil {
					return planning.Result{}, err
				}
				result.Applied++
				result.Items = append(result.Items, resultItem(change, "applied"))
			}
			if err := tx.Commit(); err != nil {
				return planning.Result{}, err
			}
			return result, nil
		},
	}
}

type assetTypeRule struct {
	Suffix      string
	AssetTypeID string
}

type typeTarget struct {
	Label string
	Icon  string
}

func parseAssetTypeRules(value interface{}) ([]assetTypeRule, error) {
	if value == nil {
		return nil, nil
	}
	items, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("asset_type_rules must be an array")
	}
	rules := make([]assetTypeRule, 0, len(items))
	for index, item := range items {
		raw, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("asset_type_rules[%d] must be an object", index)
		}
		rule := assetTypeRule{
			Suffix:      strings.TrimSpace(stringArg(raw, "suffix")),
			AssetTypeID: strings.TrimSpace(stringArg(raw, "asset_type_id")),
		}
		if rule.Suffix == "" || rule.AssetTypeID == "" {
			return nil, fmt.Errorf("asset_type_rules[%d] requires suffix and asset_type_id", index)
		}
		rules = append(rules, rule)
	}
	sort.SliceStable(rules, func(i, j int) bool {
		return len(rules[i].Suffix) > len(rules[j].Suffix)
	})
	return rules, nil
}

func matchAssetTypeRule(name string, rules []assetTypeRule) (assetTypeRule, bool) {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	for _, rule := range rules {
		if strings.HasSuffix(lowerName, strings.ToLower(rule.Suffix)) {
			return rule, true
		}
	}
	return assetTypeRule{}, false
}

func metadataTargetType(projectPath string, entityType scope.EntityType, value string) (string, string, error) {
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", "", err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback()
	if entityType == scope.TypeCollection {
		target, err := repository.GetCollectionType(tx, value)
		if err != nil {
			return "", "", err
		}
		return target.Name, target.Icon, nil
	}
	target, err := repository.GetAssetType(tx, value)
	if err != nil {
		return "", "", err
	}
	return target.Name, target.Icon, nil
}

func metadataTargetLabel(projectPath, command string, entityType scope.EntityType, value string) (string, error) {
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if command == "batch_change_status" {
		status, err := repository.GetStatus(tx, value)
		if err != nil {
			return "", err
		}
		if status.ShortName != "" {
			return status.ShortName, nil
		}
		return status.Name, nil
	}
	if entityType == scope.TypeCollection {
		target, err := repository.GetCollectionType(tx, value)
		if err != nil {
			return "", err
		}
		return target.Name, nil
	}
	target, err := repository.GetAssetType(tx, value)
	if err != nil {
		return "", err
	}
	return target.Name, nil
}

func validateMetadataTarget(projectPath, command string, entities []scope.Entity, value string) error {
	db, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if command == "batch_change_status" {
		if _, err := repository.GetStatus(tx, value); err != nil {
			return fmt.Errorf("target status not found")
		}
		return nil
	}
	needsAssetType := false
	needsCollectionType := false
	for _, entity := range entities {
		needsAssetType = needsAssetType || entity.Type == scope.TypeAsset
		needsCollectionType = needsCollectionType || entity.Type == scope.TypeCollection
	}
	if needsAssetType {
		if _, err := repository.GetAssetType(tx, value); err != nil {
			return fmt.Errorf("target asset type not found")
		}
	}
	if needsCollectionType {
		if _, err := repository.GetCollectionType(tx, value); err != nil {
			return fmt.Errorf("target collection type not found")
		}
	}
	return nil
}

func permissionForMetadata(name string) string {
	if name == "batch_change_status" {
		return "change_status"
	}
	return "update_asset"
}
