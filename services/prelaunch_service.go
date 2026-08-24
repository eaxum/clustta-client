package services

import (
	"clustta/internal/agent"
	dcctools "clustta/internal/dcc"
	"clustta/internal/repository"
	"clustta/internal/utils"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	projectRootToken        = "<ProjectRoot>"
	ocioEnvironmentVariable = "OCIO"
	fetchableAssetStatus    = "fetchable"
)

type PreLaunchTrustInfo struct {
	Required bool     `json:"required"`
	HookID   string   `json:"hook_id"`
	HookName string   `json:"hook_name"`
	Digest   string   `json:"digest"`
	Scripts  []string `json:"scripts"`
}

type preLaunchAssetPath struct {
	ID             string `db:"id"`
	Name           string `db:"name"`
	Extension      string `db:"extension"`
	Pointer        string `db:"pointer"`
	CollectionPath string `db:"collection_path"`
}

func (f *FSService) PreparePreLaunch(projectPath, remoteURL, assetID string) (PreLaunchTrustInfo, error) {
	targetPath, err := resolveAssetLaunchTarget(projectPath, assetID)
	if err != nil {
		return PreLaunchTrustInfo{}, err
	}
	assetIDs, err := resolvePreLaunchFetchableAssetIDs(projectPath, targetPath)
	if err != nil {
		return PreLaunchTrustInfo{}, err
	}
	if len(assetIDs) > 0 {
		if _, err := (&CheckpointService{}).Revert(projectPath, remoteURL, assetIDs); err != nil {
			return PreLaunchTrustInfo{}, fmt.Errorf("fetch pre-launch assets: %w", err)
		}
	}
	return f.GetPreLaunchTrust(projectPath, assetID)
}

func (f *FSService) GetPreLaunchTrust(projectPath, assetID string) (PreLaunchTrustInfo, error) {
	targetPath, err := resolveAssetLaunchTarget(projectPath, assetID)
	if err != nil {
		return PreLaunchTrustInfo{}, err
	}
	_, _, hook, scripts, variables, err := resolvePreLaunch(projectPath, targetPath)
	if err != nil {
		return PreLaunchTrustInfo{}, err
	}
	if hook == nil || len(scripts) == 0 {
		return PreLaunchTrustInfo{}, nil
	}
	hash := sha256.New()
	hookJSON, err := json.Marshal(hook)
	if err != nil {
		return PreLaunchTrustInfo{}, err
	}
	hash.Write(hookJSON)
	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		return PreLaunchTrustInfo{}, err
	}
	hash.Write(variablesJSON)
	scriptNames := make([]string, 0, len(scripts))
	for _, scriptPath := range scripts {
		scriptHash, err := utils.GenerateXXHashChecksum(scriptPath)
		if err != nil {
			return PreLaunchTrustInfo{}, err
		}
		hash.Write([]byte(scriptHash))
		scriptNames = append(scriptNames, filepath.Base(scriptPath))
	}
	return PreLaunchTrustInfo{
		Required: true, HookID: hook.ID, HookName: hook.Name,
		Digest: fmt.Sprintf("%x", hash.Sum(nil)), Scripts: scriptNames,
	}, nil
}

func (f *FSService) LaunchProjectAsset(projectPath, assetID string) error {
	targetPath, err := resolveAssetLaunchTarget(projectPath, assetID)
	if err != nil {
		return err
	}
	return f.launchProjectFile(projectPath, targetPath)
}

func resolveAssetLaunchTarget(projectPath, assetID string) (string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	_, role, err := activeAssetRole(tx)
	if err != nil {
		return "", err
	}
	if !role.PullChunk {
		return "", fmt.Errorf("user does not have pull_chunk permission")
	}
	asset, err := repository.GetAsset(tx, assetID)
	if err != nil {
		return "", err
	}
	return asset.GetFilePath(), nil
}

func (f *FSService) launchProjectFile(projectPath, targetPath string) error {
	targetPath, projectRoot, hook, scripts, variables, err := resolvePreLaunch(projectPath, targetPath)
	if err != nil {
		return err
	}
	if hook == nil {
		return f.LaunchFile(targetPath)
	}
	environment, err := buildHookEnvironment(projectRoot, variables)
	if err != nil {
		return err
	}
	if err := launchWithPreLaunchHook(targetPath, projectRoot, *hook, scripts, environment); err != nil {
		if dcctools.IsExecutableNotFound(err) {
			return err
		}
		if hook.FailurePolicy == repository.PreLaunchFailureWarn {
			return launchWithEnvironment(targetPath, environment)
		}
		return err
	}
	return nil
}

func resolvePreLaunch(projectPath, targetPath string) (
	string, string, *repository.PreLaunchHook, []string, []repository.PreLaunchEnvironmentVariable, error,
) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", "", nil, nil, nil, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return "", "", nil, nil, nil, err
	}
	defer tx.Rollback()
	projectRoot, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return "", "", nil, nil, nil, err
	}
	targetPath, err = validateLaunchPath(projectRoot, targetPath)
	if err != nil {
		return "", "", nil, nil, nil, err
	}
	settings, err := repository.GetPreLaunchHookSettings(tx)
	if err != nil {
		return "", "", nil, nil, nil, err
	}
	hook := matchingPreLaunchHook(settings, filepath.Ext(targetPath))
	if hook != nil {
		scripts, err := agent.ResolveTrackedScriptPaths(projectPath, hook.ScriptAssetIDs)
		if err != nil {
			return "", "", nil, nil, nil, err
		}
		variables := resolveHookEnvironmentVariables(settings.EnvironmentVariables, hook.EnvironmentVariableIDs)
		return targetPath, projectRoot, hook, scripts, variables, nil
	}
	return targetPath, projectRoot, nil, nil, nil, nil
}

func resolvePreLaunchFetchableAssetIDs(projectPath, targetPath string) ([]string, error) {
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
	projectRoot, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return nil, err
	}
	settings, err := repository.GetPreLaunchHookSettings(tx)
	if err != nil {
		return nil, err
	}
	hook := matchingPreLaunchHook(settings, filepath.Ext(targetPath))
	if hook == nil {
		return nil, nil
	}
	rootAssetIDs := append([]string(nil), hook.ScriptAssetIDs...)
	variables := resolveHookEnvironmentVariables(settings.EnvironmentVariables, hook.EnvironmentVariableIDs)
	environmentAssetIDs, err := resolveEnvironmentAssetIDs(tx, projectRoot, variables)
	if err != nil {
		return nil, err
	}
	rootAssetIDs = append(rootAssetIDs, environmentAssetIDs...)
	return resolveFetchableDependencyIDs(tx, rootAssetIDs)
}

func matchingPreLaunchHook(settings repository.PreLaunchHookSettings, extension string) *repository.PreLaunchHook {
	extension = strings.ToLower(extension)
	for index := range settings.Hooks {
		hook := &settings.Hooks[index]
		if hook.Enabled && containsString(hook.Extensions, extension) {
			return hook
		}
	}
	return nil
}

func resolveEnvironmentAssetIDs(
	tx *sqlx.Tx, projectRoot string, variables []repository.PreLaunchEnvironmentVariable,
) ([]string, error) {
	assets := []preLaunchAssetPath{}
	err := tx.Select(&assets, `
		SELECT a.id, a.name, a.extension, a.pointer, COALESCE(c.collection_path, '') AS collection_path
		FROM asset a
		LEFT JOIN collection c ON c.id = a.collection_id
		WHERE a.trashed = 0
	`)
	if err != nil {
		return nil, err
	}
	result := []string{}
	seen := map[string]bool{}
	for _, variable := range variables {
		targetPath, isProjectPath := resolveHookEnvironmentPath(projectRoot, variable)
		if !isProjectPath || !pathContains(projectRoot, targetPath) {
			continue
		}
		for _, asset := range assets {
			assetPath, err := asset.resolvePath(projectRoot)
			if err != nil || !pathContains(targetPath, assetPath) || seen[asset.ID] {
				continue
			}
			seen[asset.ID] = true
			result = append(result, asset.ID)
		}
	}
	return result, nil
}

func (asset preLaunchAssetPath) resolvePath(projectRoot string) (string, error) {
	if asset.Pointer != "" {
		if filepath.IsAbs(asset.Pointer) {
			return filepath.Clean(asset.Pointer), nil
		}
		return filepath.Join(projectRoot, asset.Pointer), nil
	}
	return utils.BuildAssetPath(projectRoot, asset.CollectionPath, asset.Name, asset.Extension)
}

func resolveFetchableDependencyIDs(tx *sqlx.Tx, rootAssetIDs []string) ([]string, error) {
	result := []string{}
	seen := map[string]bool{}
	for _, rootAssetID := range rootAssetIDs {
		dependencyIDs, err := repository.ResolveBuildDependencies(tx, rootAssetID)
		if err != nil {
			return nil, err
		}
		for _, dependencyID := range dependencyIDs {
			if seen[dependencyID] {
				continue
			}
			seen[dependencyID] = true
			status, err := repository.GetAssetState(tx, dependencyID)
			if err != nil {
				return nil, err
			}
			if status == fetchableAssetStatus {
				result = append(result, dependencyID)
			}
		}
	}
	return result, nil
}

func pathContains(parent, child string) bool {
	parent, err := filepath.Abs(filepath.Clean(parent))
	if err != nil {
		return false
	}
	child, err = filepath.Abs(filepath.Clean(child))
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(parent, child)
	return err == nil && !filepath.IsAbs(relative) && relative != ".." &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func resolveHookEnvironmentVariables(
	variables []repository.PreLaunchEnvironmentVariable, selectedIDs []string,
) []repository.PreLaunchEnvironmentVariable {
	selected := make(map[string]bool, len(selectedIDs))
	for _, variableID := range selectedIDs {
		selected[variableID] = true
	}
	result := make([]repository.PreLaunchEnvironmentVariable, 0, len(selectedIDs))
	for _, variable := range variables {
		if selected[variable.ID] {
			result = append(result, variable)
		}
	}
	return result
}

func validateLaunchPath(projectRoot, targetPath string) (string, error) {
	root, err := filepath.Abs(filepath.Clean(projectRoot))
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Clean(targetPath))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, target)
	if err != nil || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("asset is outside the project working directory")
	}
	if _, err := os.Stat(target); err != nil {
		return "", fmt.Errorf("file not found: %s", target)
	}
	return target, nil
}

func launchWithPreLaunchHook(
	targetPath, projectRoot string,
	hook repository.PreLaunchHook,
	scripts []string,
	environment []string,
) error {
	dcc := repository.PreLaunchDCCForExtension(filepath.Ext(targetPath))
	if len(scripts) == 0 {
		return launchDCCFile(dcc, hook.ApplicationVersion, targetPath, environment)
	}
	if dcc == "" {
		return fmt.Errorf("pre-launch scripts do not support %s files", filepath.Ext(targetPath))
	}
	pythonSource, err := buildDCCBootstrap(targetPath, projectRoot, hook, scripts)
	if err != nil {
		return err
	}
	executable, err := dcctools.FindExecutable(dcc, hook.ApplicationVersion)
	if err != nil {
		return err
	}
	var command *exec.Cmd
	if dcc == repository.PreLaunchDCCBlender {
		command = exec.Command(executable, "--python-expr", pythonSource)
	} else {
		encodedSource := base64.StdEncoding.EncodeToString([]byte(pythonSource))
		mayaCommand := fmt.Sprintf(`python("import base64;exec(base64.b64decode('%s'))")`, encodedSource)
		command = exec.Command(executable, "-command", mayaCommand)
	}
	command.Env = environment
	return command.Start()
}

func buildDCCBootstrap(targetPath, projectRoot string, hook repository.PreLaunchHook, scripts []string) (string, error) {
	targetJSON, _ := json.Marshal(targetPath)
	rootJSON, _ := json.Marshal(projectRoot)
	scriptsJSON, err := json.Marshal(scripts)
	if err != nil {
		return "", err
	}
	failurePolicyJSON, _ := json.Marshal(hook.FailurePolicy)
	contextSource := fmt.Sprintf(
		"target=%s\nproject_root=%s\nscripts=%s\nfailure_policy=%s\n",
		targetJSON, rootJSON, scriptsJSON, failurePolicyJSON,
	)
	dcc := repository.PreLaunchDCCForExtension(filepath.Ext(targetPath))
	if dcc == repository.PreLaunchDCCMaya {
		return contextSource + mayaBootstrapSource(), nil
	}
	if dcc == repository.PreLaunchDCCBlender {
		return contextSource + blenderBootstrapSource(), nil
	}
	return "", fmt.Errorf("pre-launch scripts do not support %s files", filepath.Ext(targetPath))
}

func mayaBootstrapSource() string {
	source := "import os, runpy, traceback\nimport maya.cmds as cmds\nimport maya.mel as mel\n"
	source += "context={'project_id':'','project_root':project_root,'target_file':target,'dcc':'maya'}\n"
	source += "for script in scripts:\n"
	source += "    try:\n"
	source += "        if os.path.splitext(script)[1].lower()=='.mel': mel.eval('source \"'+script.replace('\\\\','/').replace('\"','\\\\\"')+'\"')\n"
	source += "        else: runpy.run_path(script, init_globals={'CLUSTTA_CONTEXT':context})\n"
	source += "    except Exception:\n"
	source += "        if failure_policy=='block': raise\n"
	source += "        traceback.print_exc()\n"
	return source + "cmds.file(target, open=True, force=False)\n"
}

func blenderBootstrapSource() string {
	source := "import os, runpy, traceback\nimport bpy\n"
	source += "context={'project_id':'','project_root':project_root,'target_file':target,'dcc':'blender'}\n"
	source += "bpy.ops.wm.open_mainfile(filepath=target)\n"
	source += "for script in scripts:\n"
	source += "    try: runpy.run_path(script, init_globals={'CLUSTTA_CONTEXT':context})\n"
	source += "    except Exception:\n"
	source += "        if failure_policy=='block': raise\n"
	source += "        traceback.print_exc()\n"
	return source
}

func buildHookEnvironment(
	projectRoot string, variables []repository.PreLaunchEnvironmentVariable,
) ([]string, error) {
	environment := append([]string(nil), os.Environ()...)
	for _, variable := range variables {
		value, isProjectPath := resolveHookEnvironmentPath(projectRoot, variable)
		isOCIOConfig := strings.EqualFold(variable.Name, ocioEnvironmentVariable)
		if isProjectPath || isOCIOConfig {
			if _, err := os.Stat(value); err != nil {
				if os.IsNotExist(err) {
					return nil, fmt.Errorf(
						"environment variable %s points to a missing path: %s",
						variable.Name,
						value,
					)
				}
				return nil, fmt.Errorf("validate environment variable %s: %w", variable.Name, err)
			}
		}
		environment = setEnvironmentValue(environment, variable.Name, value)
	}
	return environment, nil
}

func resolveHookEnvironmentPath(
	projectRoot string, variable repository.PreLaunchEnvironmentVariable,
) (string, bool) {
	value := strings.ReplaceAll(variable.Value, projectRootToken, projectRoot)
	isProjectPath := strings.Contains(variable.Value, projectRootToken)
	isOCIOConfig := strings.EqualFold(variable.Name, ocioEnvironmentVariable)
	if isOCIOConfig && !filepath.IsAbs(value) {
		value = filepath.Join(projectRoot, value)
	}
	return filepath.Clean(value), isProjectPath || isOCIOConfig
}

func setEnvironmentValue(environment []string, name, value string) []string {
	prefix := name + "="
	for index, entry := range environment {
		if strings.EqualFold(strings.SplitN(entry, "=", 2)[0], name) {
			environment[index] = prefix + value
			return environment
		}
	}
	return append(environment, prefix+value)
}

func launchDCCFile(dcc, version, targetPath string, environment []string) error {
	if dcc != "" {
		executable, err := dcctools.FindExecutable(dcc, version)
		if err != nil {
			return err
		}
		command := exec.Command(executable, targetPath)
		command.Env = environment
		return command.Start()
	}
	return launchWithEnvironment(targetPath, environment)
}

func launchWithEnvironment(targetPath string, environment []string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("cmd", "/C", "start", "", targetPath)
	case "darwin":
		command = exec.Command("open", targetPath)
	default:
		command = exec.Command("xdg-open", targetPath)
	}
	command.Env = environment
	return command.Start()
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
