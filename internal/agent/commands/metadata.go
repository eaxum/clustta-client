package commands

import (
	"clustta/internal/agent/planning"
	"clustta/internal/agent/scope"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"

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
		Description: "Change asset and/or collection types in a structured scope. Provide the target ID for every included entity type. Local-only; manual sync required.",
		Permission:  "update_asset", Risk: "destructive",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope":              ScopeSchema([]string{"asset", "collection"}),
				"asset_type_id":      map[string]interface{}{"type": "string"},
				"collection_type_id": map[string]interface{}{"type": "string"},
			},
			"required": []string{"scope"},
		},
		Plan: func(projectPath string, args map[string]interface{}) (planning.Plan, error) {
			req, err := ParseScope(args, []scope.EntityType{scope.TypeAsset, scope.TypeCollection})
			if err != nil {
				return planning.Plan{}, err
			}
			resolved, err := scope.Resolve(projectPath, req)
			if err != nil {
				return planning.Plan{}, err
			}
			assetTypeID := stringArg(args, "asset_type_id")
			collectionTypeID := stringArg(args, "collection_type_id")
			if assetTypeID == "" && collectionTypeID == "" {
				return planning.Plan{}, fmt.Errorf("asset_type_id or collection_type_id is required")
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
				if entity.Type == scope.TypeCollection {
					key, target = "collection_type_id", collectionTypeID
				}
				labelKey, targetLabel := "asset_type", assetTypeLabel
				iconKey, targetIcon := "asset_type_icon", assetTypeIcon
				if entity.Type == scope.TypeCollection {
					labelKey, targetLabel = "collection_type", collectionTypeLabel
					iconKey, targetIcon = "collection_type_icon", collectionTypeIcon
				}
				change := planning.Change{
					Entity: entity, Action: "Change Type", Valid: true,
					Before: map[string]interface{}{key: entity.Metadata[key], labelKey: entity.Metadata[labelKey], iconKey: entity.Metadata[iconKey]},
					After:  map[string]interface{}{key: target, labelKey: targetLabel, iconKey: targetIcon},
				}
				if target == "" {
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
					value, _ := plan.Options["asset_type_id"].(string)
					err = repository.ChangeAssetType(tx, change.Entity.ID, value)
				} else {
					value, _ := plan.Options["collection_type_id"].(string)
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
