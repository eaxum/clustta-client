package agent

import (
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"fmt"
	"strings"
)

// defaultShotTasks is the per-shot task asset list created inside every shot when the caller does not
// pass shot_tasks. Pass an empty array to skip per-shot tasks entirely.
var defaultShotTasks = []string{"Animation", "Lighting", "FX"}

// defaultAssetTasks is the per-library-asset task list created inside every character/environment/prop
// folder when the caller does not pass asset_tasks. Pass an empty array to skip these.
var defaultAssetTasks = []string{"Model", "Rig", "Texture"}

// taskAssetTypeIcons maps common task names to icons so auto-created asset types render reasonably.
// Unknown task names fall back to "cube".
var taskAssetTypeIcons = map[string]string{
	"Animation":   "man-running",
	"Lighting":    "bulb",
	"FX":          "fire",
	"Compositing": "film-reel",
	"Layout":      "flow-chart",
	"Cleanup":     "bezier",
	"Color":       "palette",
	"Model":       "cube",
	"Rig":         "bone",
	"Texture":     "texture",
}

// animationSetupToolDef registers setup_animation_production with the agent.
// It is called from GetToolDefinitions in tools.go.
func animationSetupToolDef() ToolDefinition {
	return ToolDefinition{
		Name: "setup_animation_production",
		Description: "Scaffold an animation project in one call. Creates the standard asset/collection types, " +
			"Production/ and Assets/ roots, the EP/SEQ/SH tree with padded names (EP### 3-digit, SEQ### 3-digit step 10, " +
			"SH#### 4-digit step 10), and TASK ASSET FILES inside every shot AND every library entry (character/environment/prop). " +
			"Defaults: shots get Animation+Lighting+FX files; library entries get Model+Rig+Texture files. " +
			"Pass shot_tasks:[] or asset_tasks:[] to skip those. " +
			"BEFORE calling, run list_templates to see template names. Pass the chosen template name in `template`. " +
			"Use this whenever the user asks to set up an animation project (typically with a script attached). Single transaction â€” either everything is created or nothing is.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"is_series": map[string]interface{}{
					"type":        "boolean",
					"description": "True for episodic projects (creates EP layer). False for one-offs (sequences live directly under Production/).",
				},
				"episodes": map[string]interface{}{
					"type":        "array",
					"description": "Episode list. For non-series projects pass a single episode whose sequences will live under Production/. shot_count per sequence drives step-10 shot creation.",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"sequences": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"shot_count": map[string]interface{}{"type": "integer", "description": "Number of shots in this sequence (SH0010, SH0020, ...)."},
									},
									"required": []string{"shot_count"},
								},
							},
						},
						"required": []string{"sequences"},
					},
				},
				"template": map[string]interface{}{
					"type":        "string",
					"description": "Template name (case-insensitive, e.g. \"blender\", \"maya\") used for every task asset file. Required when the project has more than one template. Run list_templates first to discover names. If the project has exactly one template, leave this blank.",
				},
				"shot_tasks": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Per-shot task files to create inside every shot. Defaults to [\"Animation\", \"Lighting\", \"FX\"]. Pass [] to skip. Override with e.g. [\"Layout\",\"Animation\",\"Lighting\",\"Compositing\"].",
				},
				"asset_tasks": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Per-library-asset task files to create inside every character/environment/prop folder. Defaults to [\"Model\", \"Rig\", \"Texture\"]. Pass [] to skip.",
				},
				"characters": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Character names. Each becomes a folder under Assets/Characters/ containing the asset_tasks files.",
				},
				"environments": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Environment names. Each becomes a folder under Assets/Environments/ containing the asset_tasks files.",
				},
				"props": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Prop names. Each becomes a folder under Assets/Props/ containing the asset_tasks files.",
				},
			},
			"required": []string{"is_series", "episodes"},
		},
	}
}

// taskListArg parses a task-list argument with smart defaults: missing key OR null â†’ defaults applied;
// explicit empty array â†’ caller opts out.
func taskListArg(args map[string]interface{}, key string, defaults []string) []string {
	raw, present := args[key]
	if !present || raw == nil {
		out := make([]string, len(defaults))
		copy(out, defaults)
		return out
	}
	arr, ok := raw.([]interface{})
	if !ok {
		return []string{}
	}
	out := []string{}
	for _, v := range arr {
		s, ok := v.(string)
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// execSetupAnimationProduction runs the full animation scaffold inside one transaction. On any failure the
// entire scaffold is rolled back so the project never ends up in a partial state.
func execSetupAnimationProduction(projectPath string, args map[string]interface{}) ToolResult {
	isSeries := getBoolArg(args, "is_series", false)
	episodes := getObjSliceArg(args, "episodes")
	if len(episodes) == 0 {
		return ToolResult{Success: false, Error: "episodes is required and must not be empty"}
	}

	shotTasks := taskListArg(args, "shot_tasks", defaultShotTasks)
	assetTasks := taskListArg(args, "asset_tasks", defaultAssetTasks)
	templateName := strings.TrimSpace(getStringArg(args, "template", ""))

	characters := getStringSliceArg(args, "characters")
	environments := getStringSliceArg(args, "environments")
	props := getStringSliceArg(args, "props")

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

	preset := projectTypePresets["animation"]
	for _, at := range preset.AssetTypes {
		_, _ = repository.GetOrCreateAssetType(tx, at.Name, at.Icon)
	}
	for _, ct := range preset.CollectionTypes {
		_, _ = repository.GetOrCreateCollectionType(tx, ct.Name, ct.Icon)
	}

	collectionTypes, err := repository.GetCollectionTypes(tx)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	collectionTypeByName := map[string]string{}
	for _, t := range collectionTypes {
		collectionTypeByName[t.Name] = t.Id
	}
	requireCollectionType := func(name string) (string, error) {
		if id, ok := collectionTypeByName[name]; ok && id != "" {
			return id, nil
		}
		return "", fmt.Errorf("collection type %q is unavailable. It may be soft-deleted â€” restore or rename it from the Types panel, then retry", name)
	}

	libraryTypeID, err := requireCollectionType("Library")
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	sequenceTypeID, err := requireCollectionType("Sequence")
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	shotTypeID, err := requireCollectionType("Shot")
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	bucketTypeForName := func(name string) string {
		switch name {
		case "Characters":
			if id, ok := collectionTypeByName["Character"]; ok && id != "" {
				return id
			}
		case "Environments":
			if id, ok := collectionTypeByName["Environment"]; ok && id != "" {
				return id
			}
		case "Props":
			if id, ok := collectionTypeByName["Prop"]; ok && id != "" {
				return id
			}
		}
		return libraryTypeID
	}

	taskNames := append([]string{}, shotTasks...)
	taskNames = append(taskNames, assetTasks...)
	taskAssetTypeID := map[string]string{}
	for _, task := range taskNames {
		if _, done := taskAssetTypeID[task]; done {
			continue
		}
		icon := taskAssetTypeIcons[task]
		if icon == "" {
			icon = "cube"
		}
		at, err := repository.GetOrCreateAssetType(tx, task, icon)
		if err != nil || at.Id == "" {
			return ToolResult{Success: false, Error: fmt.Sprintf("could not ensure an asset type named %q (it may be soft-deleted). Restore or rename it from the Types panel, then retry.", task)}
		}
		taskAssetTypeID[task] = at.Id
	}

	var templateID string
	needTemplate := len(shotTasks) > 0 || (len(assetTasks) > 0 && (len(characters)+len(environments)+len(props) > 0))
	if needTemplate {
		templates, err := repository.GetTemplates(tx, false)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to list templates: %s", err.Error())}
		}
		if len(templates) == 0 {
			return ToolResult{Success: false, Error: "this project has no templates. Add a template (e.g. a Blender starter file) under Settings â†’ Templates, then retry. To scaffold without any task files, pass shot_tasks:[] and asset_tasks:[]."}
		}
		available := make([]string, 0, len(templates))
		byLowerName := map[string]string{}
		for _, t := range templates {
			byLowerName[strings.ToLower(t.Name)] = t.Id
			available = append(available, t.Name)
		}
		if templateName != "" {
			id, ok := byLowerName[strings.ToLower(templateName)]
			if !ok {
				return ToolResult{Success: false, Error: fmt.Sprintf("template %q not found. Available templates: %s", templateName, strings.Join(available, ", "))}
			}
			templateID = id
		} else if len(templates) == 1 {
			templateID = templates[0].Id
		} else {
			return ToolResult{Success: false, Error: fmt.Sprintf("this project has multiple templates and no `template` was provided. Available templates: %s. Pass one in the `template` field and call again.", strings.Join(available, ", "))}
		}
	}

	production, err := repository.CreateCollection(tx, "", "Production", "", libraryTypeID, "", "", false)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to create Production/: %s", err.Error())}
	}
	assetsRoot, err := repository.CreateCollection(tx, "", "Assets", "", libraryTypeID, "", "", false)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to create Assets/: %s", err.Error())}
	}

	bucketIDs := map[string]string{}
	for _, bucket := range []string{"Characters", "Environments", "Props"} {
		col, err := repository.CreateCollection(tx, "", bucket, "", bucketTypeForName(bucket), assetsRoot.Id, "", false)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to create Assets/%s: %s", bucket, err.Error())}
		}
		bucketIDs[bucket] = col.Id
	}

	userID := activeUserID(projectPath)
	noopProgress := func(int, int, string, string) {}

	createTaskAssets := func(parentCollection models.Collection, tasks []string, ownerLabel string) (int, error) {
		count := 0
		for _, task := range tasks {
			asset, err := repository.CreateAsset(tx, "", task, taskAssetTypeID[task], parentCollection.Id, false, templateID, "", "", nil, "", false, "", userID, "Created by Clustta Agent", "", noopProgress)
			if err != nil {
				return count, fmt.Errorf("create %s/%s: %w", ownerLabel, task, err)
			}
			if err := repository.RevertToLatestCheckpoint(tx, asset.Id, asset.GetFilePath(), noopProgress); err != nil {
				return count, fmt.Errorf("materialize %s/%s: %w", ownerLabel, task, err)
			}
			count++
		}
		return count, nil
	}

	createLibraryEntries := func(bucket string, names []string) (int, int, error) {
		entries := 0
		tasks := 0
		parentID, ok := bucketIDs[bucket]
		if !ok || len(names) == 0 {
			return entries, tasks, nil
		}
		typeID := bucketTypeForName(bucket)
		for _, name := range names {
			col, err := repository.CreateCollection(tx, "", name, "", typeID, parentID, "", false)
			if err != nil {
				return entries, tasks, fmt.Errorf("create %s/%s: %w", bucket, name, err)
			}
			entries++
			n, err := createTaskAssets(col, assetTasks, fmt.Sprintf("%s/%s", bucket, name))
			if err != nil {
				return entries, tasks, err
			}
			tasks += n
		}
		return entries, tasks, nil
	}

	libraryEntries := 0
	libraryTaskAssets := 0
	for _, bucket := range []struct {
		Name  string
		Names []string
	}{
		{"Characters", characters},
		{"Environments", environments},
		{"Props", props},
	} {
		entries, tasks, err := createLibraryEntries(bucket.Name, bucket.Names)
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		libraryEntries += entries
		libraryTaskAssets += tasks
	}

	episodeCount := 0
	sequenceCount := 0
	shotCount := 0
	shotTaskAssets := 0

	for epIdx, ep := range episodes {
		sequences := getObjSliceArg(ep, "sequences")
		if len(sequences) == 0 {
			return ToolResult{Success: false, Error: fmt.Sprintf("episode at index %d has no sequences", epIdx)}
		}

		var sequenceParent models.Collection
		if isSeries {
			epName := fmt.Sprintf("EP%03d", epIdx+1)
			episodeTypeID := libraryTypeID
			if id, ok := collectionTypeByName["Episode"]; ok && id != "" {
				episodeTypeID = id
			}
			ep, err := repository.CreateCollection(tx, "", epName, "", episodeTypeID, production.Id, "", false)
			if err != nil {
				return ToolResult{Success: false, Error: fmt.Sprintf("failed to create %s: %s", epName, err.Error())}
			}
			sequenceParent = ep
			episodeCount++
		} else {
			sequenceParent = production
		}

		for seqIdx, seq := range sequences {
			shotsRequested := getIntArg(seq, "shot_count", 0)
			if shotsRequested < 0 {
				return ToolResult{Success: false, Error: fmt.Sprintf("episode %d sequence %d has a negative shot_count", epIdx+1, seqIdx+1)}
			}
			seqName := fmt.Sprintf("SEQ%03d", (seqIdx+1)*10)
			seqCol, err := repository.CreateCollection(tx, "", seqName, "", sequenceTypeID, sequenceParent.Id, "", false)
			if err != nil {
				return ToolResult{Success: false, Error: fmt.Sprintf("failed to create %s: %s", seqName, err.Error())}
			}
			sequenceCount++

			for i := 1; i <= shotsRequested; i++ {
				shotName := fmt.Sprintf("SH%04d", i*10)
				shotCol, err := repository.CreateCollection(tx, "", shotName, "", shotTypeID, seqCol.Id, "", false)
				if err != nil {
					return ToolResult{Success: false, Error: fmt.Sprintf("failed to create %s/%s: %s", seqName, shotName, err.Error())}
				}
				shotCount++
				n, err := createTaskAssets(shotCol, shotTasks, fmt.Sprintf("%s/%s", seqName, shotName))
				if err != nil {
					return ToolResult{Success: false, Error: err.Error()}
				}
				shotTaskAssets += n
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: map[string]interface{}{
		"is_series":           isSeries,
		"production_id":       production.Id,
		"assets_id":           assetsRoot.Id,
		"bucket_ids":          bucketIDs,
		"episodes_created":    episodeCount,
		"sequences_created":   sequenceCount,
		"shots_created":       shotCount,
		"shot_tasks":          shotTasks,
		"shot_task_files":     shotTaskAssets,
		"library_entries":     libraryEntries,
		"asset_tasks":         assetTasks,
		"library_task_files":  libraryTaskAssets,
		"characters_seeded":   len(characters),
		"environments_seeded": len(environments),
		"props_seeded":        len(props),
	}}
}
