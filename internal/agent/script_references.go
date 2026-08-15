package agent

import (
	"clustta/internal/repository"
	"clustta/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ScriptReference struct {
	Type      string `json:"type"`
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Extension string `json:"extension"`
	Tracked   bool   `json:"tracked"`
}

func ListScriptReferences(projectPath string) ([]ScriptReference, error) {
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

	settings, err := repository.GetAgentScriptSettings(tx)
	if err != nil {
		return nil, err
	}
	workingDir, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return nil, err
	}
	scriptDir, err := resolveScriptDirectory(workingDir, settings.Directory)
	if err != nil {
		if os.IsNotExist(err) {
			return []ScriptReference{}, nil
		}
		return nil, err
	}
	entries, err := os.ReadDir(scriptDir)
	if err != nil {
		return nil, err
	}

	allowedExtensions := map[string]bool{}
	for _, extension := range settings.Extensions {
		allowedExtensions[strings.ToLower(extension)] = true
	}
	trackedByPath, err := trackedScriptsByPath(tx, workingDir)
	if err != nil {
		return nil, err
	}
	references := make([]ScriptReference, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !allowedExtensions[strings.ToLower(filepath.Ext(entry.Name()))] {
			continue
		}
		path, err := validateProjectFilePath(workingDir, filepath.Join(scriptDir, entry.Name()))
		if err != nil {
			continue
		}
		tracked, isTracked := trackedByPath[normalizedFilePath(path)]
		reference := ScriptReference{
			Type: "untracked_asset", Name: entry.Name(), Path: path,
			Extension: strings.ToLower(filepath.Ext(entry.Name())),
			Tracked:   isTracked,
		}
		if isTracked {
			reference.Type = "asset"
			reference.ID = tracked.ID
			reference.Name = tracked.Name
		}
		references = append(references, reference)
	}
	sort.Slice(references, func(i, j int) bool {
		return strings.ToLower(references[i].Name) < strings.ToLower(references[j].Name)
	})
	return references, nil
}

type trackedScript struct {
	ID   string
	Name string
}

func trackedScriptsByPath(tx interface {
	Select(interface{}, string, ...interface{}) error
}, workingDir string) (map[string]trackedScript, error) {
	var assets []struct {
		ID       string `db:"id"`
		Name     string `db:"name"`
		FilePath string `db:"file_path"`
		Pointer  string `db:"pointer"`
	}
	if err := tx.Select(&assets, "SELECT id, name, file_path, pointer FROM asset WHERE trashed = 0"); err != nil {
		return nil, err
	}
	tracked := make(map[string]trackedScript, len(assets))
	for _, asset := range assets {
		path := asset.FilePath
		if asset.Pointer != "" {
			path = asset.Pointer
		}
		resolved, err := validateProjectFilePath(workingDir, path)
		if err == nil {
			tracked[normalizedFilePath(resolved)] = trackedScript{ID: asset.ID, Name: asset.Name}
		}
	}
	return tracked, nil
}

func resolveScriptDirectory(workingDir, relativeDirectory string) (string, error) {
	root, err := filepath.Abs(filepath.Clean(workingDir))
	if err != nil {
		return "", err
	}
	candidate, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(relativeDirectory)))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, candidate)
	if err != nil || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("scripts directory is outside the project root")
	}
	info, err := os.Stat(candidate)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("configured scripts path is not a directory")
	}
	canonicalRoot, rootErr := filepath.EvalSymlinks(root)
	canonicalCandidate, candidateErr := filepath.EvalSymlinks(candidate)
	if rootErr == nil && candidateErr == nil {
		canonicalRelative, relErr := filepath.Rel(canonicalRoot, canonicalCandidate)
		if relErr != nil || filepath.IsAbs(canonicalRelative) || canonicalRelative == ".." || strings.HasPrefix(canonicalRelative, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("scripts directory resolves outside the project root")
		}
		candidate = canonicalCandidate
	}
	return candidate, nil
}

func normalizedFilePath(path string) string {
	return strings.ToLower(filepath.Clean(path))
}
