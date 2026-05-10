package agent

import (
	"clustta/internal/repository"
	"clustta/internal/utils"
	"strings"
)

// execSearchProjectText searches across asset, collection, checkpoint, tag and role text fields.
func execSearchProjectText(projectPath string, args map[string]interface{}) ToolResult {
	query := strings.ToLower(strings.TrimSpace(getStringArg(args, "query", "")))
	if query == "" {
		return ToolResult{Success: false, Error: "query is required"}
	}
	limit := 25
	if v, ok := args["limit"]; ok {
		switch n := v.(type) {
		case float64:
			if int(n) > 0 {
				limit = int(n)
			}
		case int:
			if n > 0 {
				limit = n
			}
		}
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

	type assetHit struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Field       string `json:"matched_in"`
	}
	type collectionHit struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Field       string `json:"matched_in"`
	}
	type checkpointHit struct {
		Id      string `json:"id"`
		AssetId string `json:"asset_id"`
		Comment string `json:"comment"`
	}
	type tagHit struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	type roleHit struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}

	results := struct {
		Assets      []assetHit      `json:"assets"`
		Collections []collectionHit `json:"collections"`
		Checkpoints []checkpointHit `json:"checkpoints"`
		Tags        []tagHit        `json:"tags"`
		Roles       []roleHit       `json:"roles"`
	}{}

	contains := func(s string) bool { return strings.Contains(strings.ToLower(s), query) }

	if assets, err := repository.GetAssets(tx, false); err == nil {
		for _, a := range assets {
			if len(results.Assets) >= limit {
				break
			}
			matched := ""
			if contains(a.Name) {
				matched = "name"
			} else if contains(a.Description) {
				matched = "description"
			}
			if matched != "" {
				results.Assets = append(results.Assets, assetHit{Id: a.Id, Name: a.Name, Description: a.Description, Field: matched})
			}
		}
	}

	if collections, err := repository.GetCollections(tx, false); err == nil {
		for _, c := range collections {
			if len(results.Collections) >= limit {
				break
			}
			matched := ""
			if contains(c.Name) {
				matched = "name"
			} else if contains(c.Description) {
				matched = "description"
			}
			if matched != "" {
				results.Collections = append(results.Collections, collectionHit{Id: c.Id, Name: c.Name, Description: c.Description, Field: matched})
			}
		}
	}

	type checkpointRow struct {
		Id      string `db:"id"`
		AssetId string `db:"asset_id"`
		Comment string `db:"comment"`
	}
	rows := []checkpointRow{}
	likeArg := "%" + query + "%"
	_ = tx.Select(&rows, "SELECT id, asset_id, comment FROM asset_checkpoint WHERE LOWER(comment) LIKE ? ORDER BY created_at DESC LIMIT ?", likeArg, limit)
	for _, r := range rows {
		results.Checkpoints = append(results.Checkpoints, checkpointHit{Id: r.Id, AssetId: r.AssetId, Comment: r.Comment})
	}

	if tags, err := repository.GetTags(tx); err == nil {
		for _, t := range tags {
			if len(results.Tags) >= limit {
				break
			}
			if contains(t.Name) {
				results.Tags = append(results.Tags, tagHit{Id: t.Id, Name: t.Name})
			}
		}
	}

	if roles, err := repository.GetRoles(tx); err == nil {
		for _, r := range roles {
			if len(results.Roles) >= limit {
				break
			}
			if contains(r.Name) {
				results.Roles = append(results.Roles, roleHit{Id: r.Id, Name: r.Name})
			}
		}
	}

	return ToolResult{Success: true, Data: results}
}
