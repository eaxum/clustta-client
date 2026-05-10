package agent

import (
	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"encoding/json"
	"fmt"
	"strings"
)

// filterDimensions is the in-memory representation of the project filter vocabulary.
type filterDimensions struct {
	Statuses        []map[string]string
	AssetTypes      []map[string]string
	CollectionTypes []map[string]string
	Tags            []map[string]string
	Users           []map[string]string
	Extensions      []string
	States          []string
	CurrentUserID   string
	Assets          []models.Asset
}

// loadFilterDimensions loads the project's filter vocabulary from SQLite.
func loadFilterDimensions(projectPath string) (*filterDimensions, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	statuses, err := repository.GetStatuses(tx)
	if err != nil {
		return nil, fmt.Errorf("statuses: %w", err)
	}
	assetTypes, err := repository.GetAssetTypes(tx)
	if err != nil {
		return nil, fmt.Errorf("asset types: %w", err)
	}
	collectionTypes, err := repository.GetCollectionTypes(tx)
	if err != nil {
		return nil, fmt.Errorf("collection types: %w", err)
	}
	tags, err := repository.GetTags(tx)
	if err != nil {
		return nil, fmt.Errorf("tags: %w", err)
	}
	users, err := repository.GetUsers(tx)
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}
	assets, err := repository.GetAssets(tx, false)
	if err != nil {
		return nil, fmt.Errorf("assets: %w", err)
	}

	dim := &filterDimensions{
		States: []string{"normal", "modified", "outdated", "rebuildable", "missing"},
	}
	for _, s := range statuses {
		dim.Statuses = append(dim.Statuses, map[string]string{
			"id": s.Id, "name": s.Name, "short_name": s.ShortName, "color": s.Color,
		})
	}
	for _, a := range assetTypes {
		dim.AssetTypes = append(dim.AssetTypes, map[string]string{"id": a.Id, "name": a.Name})
	}
	for _, c := range collectionTypes {
		dim.CollectionTypes = append(dim.CollectionTypes, map[string]string{"id": c.Id, "name": c.Name})
	}
	for _, t := range tags {
		dim.Tags = append(dim.Tags, map[string]string{"id": t.Id, "name": t.Name})
	}
	for _, u := range users {
		dim.Users = append(dim.Users, map[string]string{
			"id":         u.Id,
			"username":   u.Username,
			"email":      u.Email,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"name":       strings.TrimSpace(u.FirstName + " " + u.LastName),
		})
	}
	extSet := map[string]struct{}{}
	for _, a := range assets {
		ext := strings.TrimPrefix(strings.ToLower(a.Extension), ".")
		if ext == "" {
			continue
		}
		extSet[ext] = struct{}{}
	}
	for e := range extSet {
		dim.Extensions = append(dim.Extensions, e)
	}
	if u, err := auth_service.GetActiveUser(); err == nil {
		dim.CurrentUserID = u.Id
	}
	dim.Assets = assets
	return dim, nil
}

// execListFilterDimensions returns the project's filter vocabulary so the LLM
// can resolve natural-language filter requests against real values.
func execListFilterDimensions(projectPath string) ToolResult {
	dim, err := loadFilterDimensions(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	return ToolResult{Success: true, Data: map[string]interface{}{
		"statuses":         dim.Statuses,
		"asset_types":      dim.AssetTypes,
		"collection_types": dim.CollectionTypes,
		"tags":             dim.Tags,
		"users":            dim.Users,
		"extensions":       dim.Extensions,
		"states":           dim.States,
		"current_user_id":  dim.CurrentUserID,
	}}
}

// resolveFilterValues looks up each requested term against vocab by id, exact name, or case-insensitive name.
// Returns the matched canonical entries plus any terms that could not be resolved.
func resolveFilterValues(requested []string, vocab []map[string]string, nameKeys []string) (matched []map[string]string, unmatched []string) {
	for _, term := range requested {
		t := strings.TrimSpace(term)
		if t == "" {
			continue
		}
		var hit map[string]string
		for _, v := range vocab {
			if v["id"] == t {
				hit = v
				break
			}
		}
		if hit == nil {
			lt := strings.ToLower(t)
			for _, v := range vocab {
				for _, k := range nameKeys {
					if strings.ToLower(v[k]) == lt {
						hit = v
						break
					}
				}
				if hit != nil {
					break
				}
			}
		}
		if hit == nil {
			unmatched = append(unmatched, term)
			continue
		}
		matched = append(matched, hit)
	}
	return matched, unmatched
}

// argStringSlice coerces a tool argument value into a []string.
func argStringSlice(args map[string]interface{}, key string) []string {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	raw, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

// argBoolDefault reads a bool argument; returns false if missing or wrong type.
func argBoolDefault(args map[string]interface{}, key string) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// argStringDefault reads a string argument; returns "" if missing or wrong type.
func argStringDefault(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// execApplyBrowserFilter resolves the LLM's filter payload against the current project vocabulary
// and returns a normalized payload that the frontend listener applies without changing the workspace.
func execApplyBrowserFilter(projectPath string, args map[string]interface{}) ToolResult {
	dim, err := loadFilterDimensions(projectPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	var unmatched []string
	collect := func(label string, missed []string) {
		for _, m := range missed {
			unmatched = append(unmatched, label+":"+m)
		}
	}

	assetFilters := []map[string]interface{}{}
	resourceFilters := []map[string]interface{}{}
	collectionFilters := []map[string]interface{}{}

	if reqs := argStringSlice(args, "statuses"); len(reqs) > 0 {
		hits, miss := resolveFilterValues(reqs, dim.Statuses, []string{"name", "short_name"})
		collect("status", miss)
		for _, h := range hits {
			assetFilters = append(assetFilters, map[string]interface{}{
				"type":       "status",
				"id":         h["id"],
				"name":       strings.ToLower(h["short_name"]),
				"short_name": h["short_name"],
				"color":      h["color"],
			})
		}
	}

	if reqs := argStringSlice(args, "tags"); len(reqs) > 0 {
		hits, miss := resolveFilterValues(reqs, dim.Tags, []string{"name"})
		collect("tag", miss)
		for _, h := range hits {
			assetFilters = append(assetFilters, map[string]interface{}{
				"type": "tags",
				"id":   h["id"],
				"name": h["name"],
			})
		}
	}

	if reqs := argStringSlice(args, "asset_types"); len(reqs) > 0 {
		hits, miss := resolveFilterValues(reqs, dim.AssetTypes, []string{"name"})
		collect("asset_type", miss)
		for _, h := range hits {
			assetFilters = append(assetFilters, map[string]interface{}{
				"type": "asset-type",
				"id":   h["id"],
				"name": h["name"],
			})
		}
	}

	if reqs := argStringSlice(args, "assignees"); len(reqs) > 0 {
		resolvedReqs := make([]string, 0, len(reqs))
		for _, r := range reqs {
			if strings.EqualFold(r, "@me") || strings.EqualFold(r, "me") {
				if dim.CurrentUserID == "" {
					unmatched = append(unmatched, "assignee:@me (no active user)")
					continue
				}
				resolvedReqs = append(resolvedReqs, dim.CurrentUserID)
				continue
			}
			resolvedReqs = append(resolvedReqs, r)
		}
		hits, miss := resolveFilterValues(resolvedReqs, dim.Users, []string{"username", "email", "first_name", "last_name", "name"})
		collect("assignee", miss)
		for _, h := range hits {
			name := h["name"]
			if name == "" {
				name = h["username"]
			}
			assetFilters = append(assetFilters, map[string]interface{}{
				"type":       "assignation",
				"id":         h["id"],
				"name":       name,
				"first_name": h["first_name"],
				"last_name":  h["last_name"],
				"email":      h["email"],
				"username":   h["username"],
			})
		}
	}

	if reqs := argStringSlice(args, "states"); len(reqs) > 0 {
		stateVocab := make([]map[string]string, 0, len(dim.States))
		for _, s := range dim.States {
			stateVocab = append(stateVocab, map[string]string{"id": s, "name": s})
		}
		hits, miss := resolveFilterValues(reqs, stateVocab, []string{"name"})
		collect("state", miss)
		for _, h := range hits {
			f := map[string]interface{}{"type": "state", "name": h["name"]}
			assetFilters = append(assetFilters, f)
			resourceFilters = append(resourceFilters, f)
		}
	}

	if reqs := argStringSlice(args, "extensions"); len(reqs) > 0 {
		extVocab := make([]map[string]string, 0, len(dim.Extensions))
		for _, e := range dim.Extensions {
			extVocab = append(extVocab, map[string]string{"id": e, "name": e})
		}
		norm := make([]string, 0, len(reqs))
		for _, r := range reqs {
			norm = append(norm, strings.TrimPrefix(strings.ToLower(strings.TrimSpace(r)), "."))
		}
		hits, miss := resolveFilterValues(norm, extVocab, []string{"name"})
		collect("extension", miss)
		for _, h := range hits {
			assetFilters = append(assetFilters, map[string]interface{}{
				"type":      "extension",
				"name":      h["name"],
				"extension": "." + h["name"],
			})
		}
	}

	if reqs := argStringSlice(args, "collection_types"); len(reqs) > 0 {
		hits, miss := resolveFilterValues(reqs, dim.CollectionTypes, []string{"name"})
		collect("collection_type", miss)
		for _, h := range hits {
			collectionFilters = append(collectionFilters, map[string]interface{}{
				"type": "collection-type",
				"id":   h["id"],
				"name": h["name"],
			})
		}
	}

	hasAssignees := argBoolDefault(args, "has_assignees")
	noAssignees := argBoolDefault(args, "no_assignees")
	deep := argBoolDefault(args, "deep")
	search := argStringDefault(args, "search")

	hasAssetLevelFilters := len(assetFilters) > 0 || hasAssignees || noAssignees
	onlyAssets := false
	onlyAssetsSet := false
	if v, ok := args["only_assets"].(bool); ok {
		onlyAssets = v
		onlyAssetsSet = true
	} else if hasAssetLevelFilters {
		onlyAssets = true
		onlyAssetsSet = true
	}
	if _, explicit := args["deep"]; !explicit && hasAssetLevelFilters {
		deep = true
	}

	applied := map[string]interface{}{
		"asset_filters":      assetFilters,
		"collection_filters": collectionFilters,
		"resource_filters":   resourceFilters,
		"has_assignees":      hasAssignees,
		"no_assignees":       noAssignees,
		"use_deep":           deep,
		"view_search_query":  search,
	}

	if onlyAssetsSet {
		applied["only_assets"] = onlyAssets
	}
	for _, key := range []string{"show_collections", "show_assets", "show_resources"} {
		if v, ok := args[key].(bool); ok {
			applied[key] = v
		}
	}

	if _, err := json.Marshal(applied); err != nil {
		return ToolResult{Success: false, Error: "internal: payload not serializable: " + err.Error()}
	}

	summary := summarizeApplied(assetFilters, collectionFilters, hasAssignees, noAssignees, deep, search)

	matchCount := -1
	if hasAssetLevelFilters || search != "" {
		matchCount = countMatchingAssets(dim.Assets, assetFilters, hasAssignees, noAssignees, search)
	}

	data := map[string]interface{}{
		"applied":   applied,
		"summary":   summary,
		"unmatched": unmatched,
	}
	if matchCount >= 0 {
		data["match_count"] = matchCount
		if matchCount == 0 {
			data["empty"] = true
		}
	}
	return ToolResult{Success: true, Data: data}
}

// countMatchingAssets mirrors the asset-filter logic in stores/assets.js so the agent can tell when a
// filter would surface no items. State filters are ignored, so the count is an upper bound in that case.
func countMatchingAssets(assets []models.Asset, assetFilters []map[string]interface{}, hasAssignees, noAssignees bool, search string) int {
	statusSet := map[string]struct{}{}
	tagSet := map[string]struct{}{}
	assigneeSet := map[string]struct{}{}
	extSet := map[string]struct{}{}
	typeSet := map[string]struct{}{}
	for _, f := range assetFilters {
		t, _ := f["type"].(string)
		switch t {
		case "status":
			if n, ok := f["name"].(string); ok {
				statusSet[strings.ToLower(n)] = struct{}{}
			}
		case "tags":
			if n, ok := f["name"].(string); ok {
				tagSet[strings.ToLower(strings.ReplaceAll(n, " ", ""))] = struct{}{}
			}
		case "assignation":
			if id, ok := f["id"].(string); ok {
				assigneeSet[id] = struct{}{}
			}
		case "extension":
			if e, ok := f["extension"].(string); ok {
				extSet[strings.ToLower(e)] = struct{}{}
			}
		case "asset-type":
			if n, ok := f["name"].(string); ok {
				typeSet[strings.ToLower(n)] = struct{}{}
			}
		}
	}
	q := strings.ToLower(strings.TrimSpace(search))

	count := 0
	for _, a := range assets {
		if len(statusSet) > 0 {
			if _, ok := statusSet[strings.ToLower(a.StatusShortName)]; !ok {
				continue
			}
		}
		if len(tagSet) > 0 {
			any := false
			for _, t := range a.Tags {
				key := strings.ToLower(strings.ReplaceAll(t, " ", ""))
				if _, ok := tagSet[key]; ok {
					any = true
					break
				}
			}
			if !any {
				continue
			}
		}
		if hasAssignees && a.AssigneeId == "" {
			continue
		}
		if noAssignees && a.AssigneeId != "" {
			continue
		}
		if !hasAssignees && !noAssignees && len(assigneeSet) > 0 {
			if _, ok := assigneeSet[a.AssigneeId]; !ok {
				continue
			}
		}
		if len(extSet) > 0 {
			if _, ok := extSet[strings.ToLower(a.Extension)]; !ok {
				continue
			}
		}
		if len(typeSet) > 0 {
			if _, ok := typeSet[strings.ToLower(a.AssetTypeName)]; !ok {
				continue
			}
		}
		if q != "" {
			path := strings.ToLower(strings.ReplaceAll(a.FilePath, "\\", "/"))
			if !strings.Contains(path, q) {
				continue
			}
		}
		count++
	}
	return count
}

// execClearBrowserFilter signals the frontend to reset all browser filters via
// commonStore.resetFilters().
func execClearBrowserFilter(projectPath string, args map[string]interface{}) ToolResult {
	return ToolResult{Success: true, Data: map[string]interface{}{"cleared": true}}
}

// summarizeApplied builds a short human-readable description of what was filtered.
func summarizeApplied(assetFilters, collectionFilters []map[string]interface{}, hasAssignees, noAssignees, deep bool, search string) string {
	parts := []string{}
	groups := map[string][]string{}
	for _, f := range assetFilters {
		t, _ := f["type"].(string)
		n, _ := f["name"].(string)
		groups[t] = append(groups[t], n)
	}
	for _, f := range collectionFilters {
		t, _ := f["type"].(string)
		n, _ := f["name"].(string)
		groups[t] = append(groups[t], n)
	}
	for t, names := range groups {
		parts = append(parts, fmt.Sprintf("%s: %s", t, strings.Join(names, ", ")))
	}
	if hasAssignees {
		parts = append(parts, "has assignees")
	}
	if noAssignees {
		parts = append(parts, "unassigned")
	}
	if search != "" {
		parts = append(parts, "search: "+search)
	}
	if deep {
		parts = append(parts, "deep search")
	}
	if len(parts) == 0 {
		return "No filters applied."
	}
	return "Filtering by " + strings.Join(parts, " Â· ")
}
