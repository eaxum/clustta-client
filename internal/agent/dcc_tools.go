package agent

import (
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetDCCToolDefinitions returns tool definitions for DCC application integrations.
func GetDCCToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "open_in_dcc",
			Description: "Open asset file(s) in a DCC application (Blender, Maya, Houdini, etc.). Auto-detects the app from the file extension, or specify it explicitly.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Asset IDs to open.",
					},
					"app": map[string]interface{}{
						"type":        "string",
						"description": "Optional DCC app name (blender, maya, houdini). Auto-detected from file extension if omitted.",
					},
				},
				"required": []string{"asset_ids"},
			},
		},
		{
			Name:        "blender_render",
			Description: "Launch a Blender headless render in a visible terminal window. Renders the specified .blend file(s) with optional frame range, output path, and engine settings. Fire-and-forget — the terminal stays open so the user can monitor progress.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Asset IDs of .blend files to render.",
					},
					"start_frame": map[string]interface{}{
						"type":        "integer",
						"description": "Start frame number. Omit to use the scene default.",
					},
					"end_frame": map[string]interface{}{
						"type":        "integer",
						"description": "End frame number. Omit to use the scene default.",
					},
					"output_path": map[string]interface{}{
						"type":        "string",
						"description": "Output path template using Blender conventions (e.g. //render_). Omit for the scene default.",
					},
					"engine": map[string]interface{}{
						"type":        "string",
						"description": "Render engine override: CYCLES, BLENDER_EEVEE_NEXT, or BLENDER_WORKBENCH.",
					},
				},
				"required": []string{"asset_ids"},
			},
		},
		{
			Name:        "blender_export",
			Description: "Export .blend file(s) to another format (FBX, OBJ, glTF, USD) by launching Blender in a visible terminal. Fire-and-forget.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Asset IDs of .blend files to export.",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"fbx", "obj", "gltf", "usd"},
						"description": "Export format.",
					},
					"output_dir": map[string]interface{}{
						"type":        "string",
						"description": "Output directory. Defaults to the same directory as the source file.",
					},
				},
				"required": []string{"asset_ids", "format"},
			},
		},
		{
			Name:        "run_terminal_command",
			Description: "Run a command in a new visible terminal window. Fire-and-forget — the terminal stays open after the command finishes. Use for custom scripts, batch processing, or any command-line task.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The command to run in the terminal.",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Window title for the terminal. Defaults to 'Clustta Terminal'.",
					},
				},
				"required": []string{"command"},
			},
		},
		{
			Name:        "blender_run_script",
			Description: "Run a Python script file inside a .blend file via Blender's headless mode. Launches in a visible terminal. Use for custom automation, batch cleanup, material setup, etc.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Asset IDs of .blend files to run the script on.",
					},
					"script_path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute path to the .py script file to run.",
					},
				},
				"required": []string{"asset_ids", "script_path"},
			},
		},
		{
			Name:        "blender_run_python",
			Description: "Run inline Python code on .blend file(s) via Blender's headless mode. Writes the code to a temp file and launches Blender in a visible terminal. The file is saved after execution. Use this for any Blender automation: creating collections, modifying materials, renaming objects, adjusting scene settings, etc.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Asset IDs of .blend files to run the Python code on.",
					},
					"python_code": map[string]interface{}{
						"type":        "string",
						"description": "Python code to execute inside Blender. Has access to bpy and all Blender APIs. The file is saved automatically after execution.",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Short description of what the script does (shown in the terminal title).",
					},
				},
				"required": []string{"asset_ids", "python_code"},
			},
		},
		{
			Name:        "blender_set_settings",
			Description: "Modify render settings on .blend file(s) via Blender's headless mode. Can change render engine, resolution, frame rate, and output format. Launches in a visible terminal.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Asset IDs of .blend files to modify.",
					},
					"engine": map[string]interface{}{
						"type":        "string",
						"description": "Render engine: CYCLES, BLENDER_EEVEE_NEXT, or BLENDER_WORKBENCH.",
					},
					"resolution_x": map[string]interface{}{
						"type":        "integer",
						"description": "Horizontal resolution in pixels.",
					},
					"resolution_y": map[string]interface{}{
						"type":        "integer",
						"description": "Vertical resolution in pixels.",
					},
					"fps": map[string]interface{}{
						"type":        "integer",
						"description": "Frames per second.",
					},
					"output_format": map[string]interface{}{
						"type":        "string",
						"description": "Output file format: PNG, JPEG, OPEN_EXR, FFMPEG, etc.",
					},
					"samples": map[string]interface{}{
						"type":        "integer",
						"description": "Render samples (Cycles only).",
					},
				},
				"required": []string{"asset_ids"},
			},
		},
		{
			Name:        "blender_link",
			Description: "Link (append or reference) objects from dependency assets into a target .blend file. Uses Clustta's dependency graph to auto-resolve source files, or you can specify source asset IDs. Launches in a visible terminal.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"target_asset_id": map[string]interface{}{
						"type":        "string",
						"description": "Asset ID of the target .blend file to link into.",
					},
					"source_asset_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Optional. Specific asset IDs to link from. If omitted, uses the target's Clustta dependency graph.",
					},
					"link_mode": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"append", "link"},
						"description": "Whether to append (copy into file) or link (reference). Defaults to 'link'.",
					},
					"object_types": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Blender data types to link: Collection, Object, Material, NodeTree, etc. Defaults to ['Collection'].",
					},
					"data_names": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Optional. Specific data block names to link (e.g., ['Chair', 'Table']). If omitted, links ALL data blocks of the specified types. When linking from dependency graph, use the asset names.",
					},
				},
				"required": []string{"target_asset_id"},
			},
		},
	}
}

// --- DCC tool execution functions ---

// execOpenInDCC opens asset files in the appropriate DCC application.
func execOpenInDCC(projectPath string, args map[string]interface{}) ToolResult {
	assetIDs := getStringSliceArg(args, "asset_ids")
	if len(assetIDs) == 0 {
		return ToolResult{Success: false, Error: "asset_ids is required"}
	}
	appName := getStringArg(args, "app", "")

	paths, err := resolveAssetFilePaths(projectPath, assetIDs)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	opened := 0
	for _, p := range paths {
		app := appName
		if app == "" {
			app = detectDCCFromExtension(filepath.Ext(p))
		}
		if app == "" {
			if err := openWithDefault(p); err != nil {
				continue
			}
			opened++
			continue
		}

		exePath, err := findDCCExecutable(app)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("%s not found: %s", app, err.Error())}
		}

		cmd := exec.Command(exePath, p)
		cmd.Dir = filepath.Dir(p)
		if err := cmd.Start(); err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to open %s: %s", filepath.Base(p), err.Error())}
		}
		opened++
	}

	return ToolResult{Success: true, Data: fmt.Sprintf("Opened %d file(s)", opened)}
}

// execBlenderRender launches a headless Blender render in a visible terminal.
func execBlenderRender(projectPath string, args map[string]interface{}) ToolResult {
	assetIDs := getStringSliceArg(args, "asset_ids")
	if len(assetIDs) == 0 {
		return ToolResult{Success: false, Error: "asset_ids is required"}
	}

	blenderPath, err := findDCCExecutable("blender")
	if err != nil {
		return ToolResult{Success: false, Error: "Blender not found: " + err.Error()}
	}

	paths, err := resolveAssetFilePaths(projectPath, assetIDs)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	outputPath := getStringArg(args, "output_path", "")
	engine := getStringArg(args, "engine", "")
	startFrame := getIntArg(args, "start_frame", -1)
	endFrame := getIntArg(args, "end_frame", -1)

	for _, p := range paths {
		if !strings.HasSuffix(strings.ToLower(p), ".blend") {
			return ToolResult{Success: false, Error: fmt.Sprintf("%s is not a .blend file", filepath.Base(p))}
		}

		cmdParts := []string{quoteArg(blenderPath), "--background", quoteArg(p)}
		if engine != "" {
			cmdParts = append(cmdParts, "--engine", engine)
		}
		if outputPath != "" {
			cmdParts = append(cmdParts, "--render-output", quoteArg(outputPath))
		}
		if startFrame >= 0 {
			cmdParts = append(cmdParts, "--frame-start", fmt.Sprintf("%d", startFrame))
		}
		if endFrame >= 0 {
			cmdParts = append(cmdParts, "--frame-end", fmt.Sprintf("%d", endFrame))
		}
		cmdParts = append(cmdParts, "--render-anim")

		command := strings.Join(cmdParts, " ")
		title := fmt.Sprintf("Blender Render - %s", filepath.Base(p))
		if err := launchInTerminal(title, filepath.Dir(p), command); err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to launch render: %s", err.Error())}
		}
	}

	return ToolResult{Success: true, Data: fmt.Sprintf("Launched render for %d file(s) in terminal", len(paths))}
}

// execBlenderExport launches a Blender export in a visible terminal.
func execBlenderExport(projectPath string, args map[string]interface{}) ToolResult {
	assetIDs := getStringSliceArg(args, "asset_ids")
	format := getStringArg(args, "format", "")
	if len(assetIDs) == 0 || format == "" {
		return ToolResult{Success: false, Error: "asset_ids and format are required"}
	}

	blenderPath, err := findDCCExecutable("blender")
	if err != nil {
		return ToolResult{Success: false, Error: "Blender not found: " + err.Error()}
	}

	paths, err := resolveAssetFilePaths(projectPath, assetIDs)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	outputDir := getStringArg(args, "output_dir", "")

	exportOps := map[string]string{
		"fbx":  "bpy.ops.export_scene.fbx(filepath=r'%s')",
		"obj":  "bpy.ops.wm.obj_export(filepath=r'%s')",
		"gltf": "bpy.ops.export_scene.gltf(filepath=r'%s')",
		"usd":  "bpy.ops.wm.usd_export(filepath=r'%s')",
	}

	opTemplate, ok := exportOps[strings.ToLower(format)]
	if !ok {
		return ToolResult{Success: false, Error: fmt.Sprintf("unsupported format: %s (use fbx, obj, gltf, or usd)", format)}
	}

	for _, p := range paths {
		if !strings.HasSuffix(strings.ToLower(p), ".blend") {
			return ToolResult{Success: false, Error: fmt.Sprintf("%s is not a .blend file", filepath.Base(p))}
		}

		outDir := outputDir
		if outDir == "" {
			outDir = filepath.Dir(p)
		}
		baseName := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		outFile := filepath.Join(outDir, baseName+"."+strings.ToLower(format))

		pythonExpr := fmt.Sprintf("import bpy; "+opTemplate, outFile)

		command := fmt.Sprintf("%s --background %s --python-expr %s",
			quoteArg(blenderPath), quoteArg(p), quoteArg(pythonExpr))

		title := fmt.Sprintf("Blender Export - %s to %s", filepath.Base(p), strings.ToUpper(format))
		if err := launchInTerminal(title, filepath.Dir(p), command); err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to launch export: %s", err.Error())}
		}
	}

	return ToolResult{Success: true, Data: fmt.Sprintf("Launched %s export for %d file(s) in terminal", strings.ToUpper(format), len(paths))}
}

// execRunTerminalCommand opens a terminal and runs a command string.
func execRunTerminalCommand(projectPath string, args map[string]interface{}) ToolResult {
	command := getStringArg(args, "command", "")
	if command == "" {
		return ToolResult{Success: false, Error: "command is required"}
	}

	title := getStringArg(args, "title", "Clustta Terminal")
	workDir := filepath.Dir(projectPath)

	if err := launchInTerminal(title, workDir, command); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Data: "Command launched in terminal"}
}

// execBlenderRunScript runs a user-provided .py script on .blend files via terminal.
func execBlenderRunScript(projectPath string, args map[string]interface{}) ToolResult {
	assetIDs := getStringSliceArg(args, "asset_ids")
	scriptPath := getStringArg(args, "script_path", "")
	if len(assetIDs) == 0 || scriptPath == "" {
		return ToolResult{Success: false, Error: "asset_ids and script_path are required"}
	}

	if _, err := os.Stat(scriptPath); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("script file not found: %s", scriptPath)}
	}

	blenderPath, err := findDCCExecutable("blender")
	if err != nil {
		return ToolResult{Success: false, Error: "Blender not found: " + err.Error()}
	}

	paths, err := resolveAssetFilePaths(projectPath, assetIDs)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	for _, p := range paths {
		if !strings.HasSuffix(strings.ToLower(p), ".blend") {
			return ToolResult{Success: false, Error: fmt.Sprintf("%s is not a .blend file", filepath.Base(p))}
		}

		command := fmt.Sprintf("%s --background %s --python %s",
			quoteArg(blenderPath), quoteArg(p), quoteArg(scriptPath))

		title := fmt.Sprintf("Blender Script - %s on %s", filepath.Base(scriptPath), filepath.Base(p))
		if err := launchInTerminal(title, filepath.Dir(p), command); err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to launch script: %s", err.Error())}
		}
	}

	return ToolResult{Success: true, Data: fmt.Sprintf("Launched script %s on %d file(s) in terminal", filepath.Base(scriptPath), len(paths))}
}

// execBlenderRunPython writes inline Python code to a temp file, runs it on .blend file(s) via Blender headless, and saves the file.
func execBlenderRunPython(projectPath string, args map[string]interface{}) ToolResult {
	assetIDs := getStringSliceArg(args, "asset_ids")
	pythonCode := getStringArg(args, "python_code", "")
	if len(assetIDs) == 0 || pythonCode == "" {
		return ToolResult{Success: false, Error: "asset_ids and python_code are required"}
	}

	description := getStringArg(args, "description", "Python script")

	blenderPath, err := findDCCExecutable("blender")
	if err != nil {
		return ToolResult{Success: false, Error: "Blender not found: " + err.Error()}
	}

	paths, err := resolveAssetFilePaths(projectPath, assetIDs)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	// Write the Python code to a temp file, appending a save at the end
	tmpScript, err := os.CreateTemp("", "clustta-py-*.py")
	if err != nil {
		return ToolResult{Success: false, Error: "failed to create temp script: " + err.Error()}
	}
	scriptContent := "import bpy\n" + pythonCode + "\nbpy.ops.wm.save_mainfile()\n"
	if _, err := tmpScript.WriteString(scriptContent); err != nil {
		tmpScript.Close()
		return ToolResult{Success: false, Error: err.Error()}
	}
	tmpScript.Close()
	scriptPath := tmpScript.Name()

	for _, p := range paths {
		if !strings.HasSuffix(strings.ToLower(p), ".blend") {
			return ToolResult{Success: false, Error: fmt.Sprintf("%s is not a .blend file", filepath.Base(p))}
		}

		command := fmt.Sprintf("%s --background %s --python %s",
			quoteArg(blenderPath), quoteArg(p), quoteArg(scriptPath))

		title := fmt.Sprintf("Blender - %s - %s", description, filepath.Base(p))
		if err := launchInTerminal(title, filepath.Dir(p), command); err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to launch: %s", err.Error())}
		}
	}

	return ToolResult{Success: true, Data: fmt.Sprintf("Launched '%s' on %d file(s) in terminal", description, len(paths))}
}

// execBlenderSetSettings modifies render settings on .blend files via a generated Python expression.
func execBlenderSetSettings(projectPath string, args map[string]interface{}) ToolResult {
	assetIDs := getStringSliceArg(args, "asset_ids")
	if len(assetIDs) == 0 {
		return ToolResult{Success: false, Error: "asset_ids is required"}
	}

	blenderPath, err := findDCCExecutable("blender")
	if err != nil {
		return ToolResult{Success: false, Error: "Blender not found: " + err.Error()}
	}

	paths, err := resolveAssetFilePaths(projectPath, assetIDs)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	// Build Python statements for each setting
	var stmts []string
	stmts = append(stmts, "import bpy", "s = bpy.context.scene")

	engine := getStringArg(args, "engine", "")
	if engine != "" {
		stmts = append(stmts, fmt.Sprintf("s.render.engine = '%s'", engine))
	}
	resX := getIntArg(args, "resolution_x", -1)
	if resX > 0 {
		stmts = append(stmts, fmt.Sprintf("s.render.resolution_x = %d", resX))
	}
	resY := getIntArg(args, "resolution_y", -1)
	if resY > 0 {
		stmts = append(stmts, fmt.Sprintf("s.render.resolution_y = %d", resY))
	}
	fps := getIntArg(args, "fps", -1)
	if fps > 0 {
		stmts = append(stmts, fmt.Sprintf("s.render.fps = %d", fps))
	}
	outputFormat := getStringArg(args, "output_format", "")
	if outputFormat != "" {
		stmts = append(stmts, fmt.Sprintf("s.render.image_settings.file_format = '%s'", outputFormat))
	}
	samples := getIntArg(args, "samples", -1)
	if samples > 0 {
		stmts = append(stmts, fmt.Sprintf("s.cycles.samples = %d", samples))
	}

	// Must have at least one setting change beyond the imports
	if len(stmts) <= 2 {
		return ToolResult{Success: false, Error: "at least one setting must be specified (engine, resolution_x, resolution_y, fps, output_format, or samples)"}
	}

	stmts = append(stmts, "bpy.ops.wm.save_mainfile()")
	pythonExpr := strings.Join(stmts, "; ")

	for _, p := range paths {
		if !strings.HasSuffix(strings.ToLower(p), ".blend") {
			return ToolResult{Success: false, Error: fmt.Sprintf("%s is not a .blend file", filepath.Base(p))}
		}

		command := fmt.Sprintf("%s --background %s --python-expr %s",
			quoteArg(blenderPath), quoteArg(p), quoteArg(pythonExpr))

		title := fmt.Sprintf("Blender Settings - %s", filepath.Base(p))
		if err := launchInTerminal(title, filepath.Dir(p), command); err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to launch settings change: %s", err.Error())}
		}
	}

	return ToolResult{Success: true, Data: fmt.Sprintf("Launched settings update for %d file(s) in terminal", len(paths))}
}

// execBlenderLink links/appends objects from source .blend files into a target .blend file.
// If no source_asset_ids are provided, uses the target asset's Clustta dependency graph.
func execBlenderLink(projectPath string, args map[string]interface{}) ToolResult {
	targetAssetID := getStringArg(args, "target_asset_id", "")
	if targetAssetID == "" {
		return ToolResult{Success: false, Error: "target_asset_id is required"}
	}

	linkMode := getStringArg(args, "link_mode", "link")
	if linkMode != "append" && linkMode != "link" {
		return ToolResult{Success: false, Error: "link_mode must be 'append' or 'link'"}
	}

	objectTypes := getStringSliceArg(args, "object_types")
	if len(objectTypes) == 0 {
		objectTypes = []string{"Collection"}
	}

	dataNames := getStringSliceArg(args, "data_names")

	blenderPath, err := findDCCExecutable("blender")
	if err != nil {
		return ToolResult{Success: false, Error: "Blender not found: " + err.Error()}
	}

	// Resolve target file
	targetPaths, err := resolveAssetFilePaths(projectPath, []string{targetAssetID})
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	targetFile := targetPaths[0]
	if !strings.HasSuffix(strings.ToLower(targetFile), ".blend") {
		return ToolResult{Success: false, Error: fmt.Sprintf("target %s is not a .blend file", filepath.Base(targetFile))}
	}

	// Resolve source files — explicit IDs or from dependency graph
	sourceAssetIDs := getStringSliceArg(args, "source_asset_ids")

	if len(sourceAssetIDs) == 0 {
		// Auto-resolve from Clustta dependency graph
		dbConn, err := utils.OpenDb(projectPath)
		if err != nil {
			return ToolResult{Success: false, Error: "failed to open database: " + err.Error()}
		}
		defer dbConn.Close()
		tx, err := dbConn.Beginx()
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		defer tx.Rollback()

		deps, err := repository.GetAssetDependencies(tx, targetAssetID)
		if err != nil {
			return ToolResult{Success: false, Error: "failed to get dependencies: " + err.Error()}
		}
		if len(deps) == 0 {
			return ToolResult{Success: false, Error: "no dependencies found for this asset and no source_asset_ids specified"}
		}
		for _, d := range deps {
			sourceAssetIDs = append(sourceAssetIDs, d.DependencyId)
		}
	}

	sourcePaths, err := resolveAssetFilePaths(projectPath, sourceAssetIDs)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	// Filter to only .blend source files
	var blendSources []string
	for _, sp := range sourcePaths {
		if strings.HasSuffix(strings.ToLower(sp), ".blend") {
			blendSources = append(blendSources, sp)
		}
	}
	if len(blendSources) == 0 {
		return ToolResult{Success: false, Error: "no .blend files found among source assets"}
	}

	// Build Python script that links/appends from each source.
	// We write a multi-line script to a temp .py file since --python-expr has quoting issues
	// with complex scripts. The script uses bpy.data.libraries.load() for reliable linking.
	var pyLines []string
	pyLines = append(pyLines, "import bpy")

	linkBool := "True"
	if linkMode == "append" {
		linkBool = "False"
	}

	// Build a Python set of allowed names for filtering
	hasNameFilter := len(dataNames) > 0
	if hasNameFilter {
		nameSet := make([]string, 0, len(dataNames))
		for _, n := range dataNames {
			nameSet = append(nameSet, fmt.Sprintf("'%s'", strings.ReplaceAll(n, "'", "\\''")))
		}
		pyLines = append(pyLines, fmt.Sprintf("_names = {%s}", strings.Join(nameSet, ", ")))
	}

	for _, srcFile := range blendSources {
		srcFileEscaped := strings.ReplaceAll(srcFile, `\`, `\\`)
		for _, objType := range objectTypes {
			dataAttr := strings.ToLower(objType) + "s"
			if hasNameFilter {
				// Only load data blocks whose names match the filter
				pyLines = append(pyLines, fmt.Sprintf(
					`with bpy.data.libraries.load(r'%s', link=%s) as (src, dst): dst.%s = [n for n in src.%s if n in _names]`,
					srcFileEscaped, linkBool, dataAttr, dataAttr,
				))
			} else {
				pyLines = append(pyLines, fmt.Sprintf(
					`with bpy.data.libraries.load(r'%s', link=%s) as (src, dst): dst.%s = src.%s`,
					srcFileEscaped, linkBool, dataAttr, dataAttr,
				))
			}
		}
	}

	// Instance loaded data into the active scene
	for _, objType := range objectTypes {
		switch strings.ToLower(objType) {
		case "collection":
			pyLines = append(pyLines,
				`for c in bpy.data.collections:`,
				`    if c.users == 0 or c.library is not None:`,
				`        try: bpy.context.scene.collection.children.link(c)`,
				`        except: pass`,
			)
		case "object":
			pyLines = append(pyLines,
				`for o in bpy.data.objects:`,
				`    if o.users == 0 or o.library is not None:`,
				`        try: bpy.context.scene.collection.objects.link(o)`,
				`        except: pass`,
			)
		}
	}

	pyLines = append(pyLines, "bpy.ops.wm.save_mainfile()")

	// Write script to a temp .py file
	tmpScript, err := os.CreateTemp("", "clustta-link-*.py")
	if err != nil {
		return ToolResult{Success: false, Error: "failed to create temp script: " + err.Error()}
	}
	scriptContent := strings.Join(pyLines, "\n") + "\n"
	if _, err := tmpScript.WriteString(scriptContent); err != nil {
		tmpScript.Close()
		return ToolResult{Success: false, Error: err.Error()}
	}
	tmpScript.Close()

	command := fmt.Sprintf("%s --background %s --python %s",
		quoteArg(blenderPath), quoteArg(targetFile), quoteArg(tmpScript.Name()))

	title := fmt.Sprintf("Blender Link - %s (%d sources)", filepath.Base(targetFile), len(blendSources))
	if err := launchInTerminal(title, filepath.Dir(targetFile), command); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to launch link operation: %s", err.Error())}
	}

	return ToolResult{Success: true, Data: fmt.Sprintf("Launched %s of %d source(s) into %s in terminal", linkMode, len(blendSources), filepath.Base(targetFile))}
}

// --- Shared helpers ---

// resolveAssetFilePaths looks up asset IDs and returns their absolute file paths.
// All resolved paths are validated to live inside the project directory so that
// a maliciously crafted asset.GetFilePath() (e.g. "../../etc/passwd" or an
// absolute path to a system file) cannot trick DCC tools into opening files
// outside the project root.
func resolveAssetFilePaths(projectPath string, assetIDs []string) ([]string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %s", err.Error())
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	paths := make([]string, 0, len(assetIDs))
	for _, id := range assetIDs {
		asset, err := repository.GetAsset(tx, id)
		if err != nil {
			return nil, fmt.Errorf("asset %s not found: %s", id, err.Error())
		}

		fp := asset.GetFilePath()
		if fp == "" {
			return nil, fmt.Errorf("asset %s (%s) has no file path", asset.Name, id)
		}

		safePath, err := validateAssetPath(projectPath, fp)
		if err != nil {
			return nil, fmt.Errorf("asset %s (%s): %w", asset.Name, id, err)
		}

		paths = append(paths, safePath)
	}

	return paths, nil
}

// validateAssetPath resolves fp relative to the project directory and rejects
// paths that escape the project root. Returns the cleaned absolute path.
func validateAssetPath(projectPath, fp string) (string, error) {
	projectDir, err := filepath.Abs(filepath.Dir(projectPath))
	if err != nil {
		return "", fmt.Errorf("invalid project path: %w", err)
	}

	if !filepath.IsAbs(fp) {
		fp = filepath.Join(projectDir, fp)
	}
	abs, err := filepath.Abs(filepath.Clean(fp))
	if err != nil {
		return "", fmt.Errorf("invalid file path: %w", err)
	}

	// Compare with a trailing separator so "/projectAbc" doesn't match "/project".
	rootWithSep := projectDir + string(filepath.Separator)
	if !strings.EqualFold(abs, projectDir) && !strings.HasPrefix(strings.ToLower(abs)+string(filepath.Separator), strings.ToLower(rootWithSep)) {
		return "", fmt.Errorf("path %s is outside the project directory", abs)
	}
	return abs, nil
}

// detectDCCFromExtension returns the DCC app name for a file extension.
func detectDCCFromExtension(ext string) string {
	switch strings.ToLower(ext) {
	case ".blend", ".blend1":
		return "blender"
	case ".ma", ".mb":
		return "maya"
	case ".hip", ".hipnc", ".hiplc":
		return "houdini"
	case ".nk", ".nknc":
		return "nuke"
	case ".spp":
		return "substance_painter"
	case ".sbsar":
		return "substance_designer"
	case ".zpr":
		return "zbrush"
	case ".c4d":
		return "cinema4d"
	default:
		return ""
	}
}

// findDCCExecutable locates a DCC application executable on the system.
func findDCCExecutable(name string) (string, error) {
	// Check environment variable first (e.g., BLENDER_PATH, MAYA_PATH)
	envKey := strings.ToUpper(name) + "_PATH"
	if envPath := os.Getenv(envKey); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
	}

	// Check common installation paths
	for _, p := range dccDefaultPaths(name) {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	// Fall back to PATH lookup
	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}

	return "", fmt.Errorf("%s not found — set %s environment variable or add it to PATH", name, envKey)
}

// dccDefaultPaths returns common installation paths for a DCC application.
func dccDefaultPaths(name string) []string {
	switch runtime.GOOS {
	case "windows":
		return dccWindowsPaths(name)
	case "darwin":
		return dccDarwinPaths(name)
	default:
		return dccLinuxPaths(name)
	}
}

func dccWindowsPaths(name string) []string {
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = `C:\Program Files`
	}

	switch strings.ToLower(name) {
	case "blender":
		var paths []string
		base := filepath.Join(programFiles, "Blender Foundation")
		if entries, err := os.ReadDir(base); err == nil {
			for i := len(entries) - 1; i >= 0; i-- {
				paths = append(paths, filepath.Join(base, entries[i].Name(), "blender.exe"))
			}
		}
		return paths
	case "maya":
		var paths []string
		for year := 2026; year >= 2020; year-- {
			paths = append(paths, filepath.Join(programFiles, fmt.Sprintf(`Autodesk\Maya%d\bin\maya.exe`, year)))
		}
		return paths
	case "houdini":
		var paths []string
		base := filepath.Join(programFiles, "Side Effects Software")
		if entries, err := os.ReadDir(base); err == nil {
			for i := len(entries) - 1; i >= 0; i-- {
				paths = append(paths, filepath.Join(base, entries[i].Name(), "bin", "houdini.exe"))
			}
		}
		return paths
	default:
		return nil
	}
}

func dccDarwinPaths(name string) []string {
	switch strings.ToLower(name) {
	case "blender":
		return []string{"/Applications/Blender.app/Contents/MacOS/Blender"}
	case "maya":
		var paths []string
		for year := 2026; year >= 2020; year-- {
			paths = append(paths, fmt.Sprintf("/Applications/Autodesk/maya%d/Maya.app/Contents/bin/maya", year))
		}
		return paths
	case "houdini":
		return []string{"/Applications/Houdini/Current/Houdini FX.app/Contents/MacOS/houdini"}
	default:
		return nil
	}
}

func dccLinuxPaths(name string) []string {
	switch strings.ToLower(name) {
	case "blender":
		return []string{"/usr/bin/blender", "/snap/bin/blender", "/usr/local/bin/blender"}
	case "maya":
		var paths []string
		for year := 2026; year >= 2020; year-- {
			paths = append(paths, fmt.Sprintf("/usr/autodesk/maya%d/bin/maya", year))
		}
		return paths
	case "houdini":
		return []string{"/opt/hfs/bin/houdini"}
	default:
		return nil
	}
}

// --- Terminal launch ---

// launchInTerminal opens a visible terminal window and runs the given command string.
// The terminal closes automatically when the command finishes.
func launchInTerminal(title, workDir, command string) error {
	if command == "" {
		return fmt.Errorf("no command specified")
	}

	switch runtime.GOOS {
	case "windows":
		return launchTerminalWindows(title, workDir, command)
	case "darwin":
		return launchTerminalDarwin(title, workDir, command)
	default:
		return launchTerminalLinux(title, workDir, command)
	}
}

// launchTerminalWindows writes a temp .bat script and opens it in a new cmd window.
func launchTerminalWindows(title, workDir, command string) error {
	tmpFile, err := os.CreateTemp("", "clustta-*.bat")
	if err != nil {
		return fmt.Errorf("failed to create temp script: %s", err.Error())
	}

	batContent := fmt.Sprintf("@echo off\r\ntitle %s\r\ncd /d \"%s\"\r\necho.\r\necho %s\r\necho.\r\n%s\r\necho.\r\necho === Done ===\r\ndel \"%%~f0\"\r\n",
		title, workDir, title, command)

	if _, err := tmpFile.WriteString(batContent); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	cmd := exec.Command("cmd", "/c", "start", "", tmpFile.Name())
	return cmd.Start()
}

// launchTerminalDarwin opens Terminal.app with the command via osascript.
func launchTerminalDarwin(title, workDir, command string) error {
	script := fmt.Sprintf(`tell application "Terminal"
	activate
	do script "cd %q && echo '%s' && %s; exit"
end tell`, workDir, title, command)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Start()
}

// launchTerminalLinux tries common terminal emulators.
func launchTerminalLinux(title, workDir, command string) error {
	fullCmd := command + "; echo; echo '=== Done ==='"

	terminals := []struct {
		name string
		args []string
	}{
		{"gnome-terminal", []string{"--title=" + title, "--working-directory=" + workDir, "--", "bash", "-c", fullCmd}},
		{"konsole", []string{"--workdir", workDir, "-e", "bash", "-c", fullCmd}},
		{"xfce4-terminal", []string{"--title=" + title, "--working-directory=" + workDir, "-e", "bash -c '" + fullCmd + "'"}},
		{"xterm", []string{"-T", title, "-e", "bash", "-c", "cd " + workDir + " && " + fullCmd}},
	}

	for _, t := range terminals {
		if p, err := exec.LookPath(t.name); err == nil {
			cmd := exec.Command(p, t.args...)
			return cmd.Start()
		}
	}

	return fmt.Errorf("no supported terminal emulator found (tried gnome-terminal, konsole, xfce4-terminal, xterm)")
}

// openWithDefault opens a file with the OS default application.
func openWithDefault(filePath string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", filePath).Start()
	case "darwin":
		return exec.Command("open", filePath).Start()
	default:
		return exec.Command("xdg-open", filePath).Start()
	}
}

// quoteArg wraps an argument in quotes if it contains spaces.
func quoteArg(arg string) string {
	if strings.Contains(arg, " ") {
		return `"` + arg + `"`
	}
	return arg
}

// getIntArg extracts an integer argument with a default fallback.
func getIntArg(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return defaultVal
}
