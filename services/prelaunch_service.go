package services

import (
	"clustta/internal/agent"
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
)

const projectRootToken = "<ProjectRoot>"

type PreLaunchTrustInfo struct {
	Required bool     `json:"required"`
	HookID   string   `json:"hook_id"`
	HookName string   `json:"hook_name"`
	Digest   string   `json:"digest"`
	Scripts  []string `json:"scripts"`
}

func (f *FSService) GetPreLaunchTrust(projectPath, assetID string) (PreLaunchTrustInfo, error) {
	targetPath, err := resolveAssetLaunchTarget(projectPath, assetID)
	if err != nil {
		return PreLaunchTrustInfo{}, err
	}
	_, _, hook, scripts, err := resolvePreLaunch(projectPath, targetPath)
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
	targetPath, projectRoot, hook, scripts, err := resolvePreLaunch(projectPath, targetPath)
	if err != nil {
		return err
	}
	if hook == nil {
		return f.LaunchFile(targetPath)
	}
	if err := launchWithPreLaunchHook(targetPath, projectRoot, *hook, scripts); err != nil {
		if hook.FailurePolicy == repository.PreLaunchFailureWarn {
			return launchWithEnvironment(targetPath, buildHookEnvironment(projectRoot, hook.EnvironmentVariables))
		}
		return err
	}
	return nil
}

func resolvePreLaunch(projectPath, targetPath string) (string, string, *repository.PreLaunchHook, []string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", "", nil, nil, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return "", "", nil, nil, err
	}
	defer tx.Rollback()
	projectRoot, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return "", "", nil, nil, err
	}
	targetPath, err = validateLaunchPath(projectRoot, targetPath)
	if err != nil {
		return "", "", nil, nil, err
	}
	settings, err := repository.GetPreLaunchHookSettings(tx)
	if err != nil {
		return "", "", nil, nil, err
	}
	extension := strings.ToLower(filepath.Ext(targetPath))
	for index := range settings.Hooks {
		hook := &settings.Hooks[index]
		if !hook.Enabled || !containsString(hook.Extensions, extension) {
			continue
		}
		scripts, err := agent.ResolveTrackedScriptPaths(projectPath, hook.ScriptAssetIDs)
		if err != nil {
			return "", "", nil, nil, err
		}
		return targetPath, projectRoot, hook, scripts, nil
	}
	return targetPath, projectRoot, nil, nil, nil
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

func launchWithPreLaunchHook(targetPath, projectRoot string, hook repository.PreLaunchHook, scripts []string) error {
	environment := buildHookEnvironment(projectRoot, hook.EnvironmentVariables)
	dcc := repository.PreLaunchDCCForExtension(filepath.Ext(targetPath))
	if len(scripts) == 0 {
		return launchDCCFile(dcc, targetPath, environment)
	}
	if dcc == "" {
		return fmt.Errorf("pre-launch scripts do not support %s files", filepath.Ext(targetPath))
	}
	pythonSource, err := buildDCCBootstrap(targetPath, projectRoot, hook, scripts)
	if err != nil {
		return err
	}
	executable, err := agent.FindDCCExecutable(dcc)
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

func buildHookEnvironment(projectRoot string, variables []repository.PreLaunchEnvironmentVariable) []string {
	environment := append([]string(nil), os.Environ()...)
	for _, variable := range variables {
		value := strings.ReplaceAll(variable.Value, projectRootToken, projectRoot)
		environment = setEnvironmentValue(environment, variable.Name, value)
	}
	return environment
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

func launchDCCFile(dcc, targetPath string, environment []string) error {
	if dcc != "" {
		executable, err := agent.FindDCCExecutable(dcc)
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
