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

	settings, err := repository.GetProjectScriptSettings(tx)
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
		if entry.IsDir() {
			continue
		}
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if !allowedExtensions[extension] {
			continue
		}
		path, err := validateProjectFilePath(workingDir, filepath.Join(scriptDir, entry.Name()))
		if err != nil {
			continue
		}
		tracked, isTracked := trackedByPath[normalizedFilePath(path)]
		reference := ScriptReference{
			Type: "untracked_asset", Name: entry.Name(), Path: path,
			Extension: extension,
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

func ResolveTrackedScriptPaths(projectPath string, assetIDs []string) ([]string, error) {
	references, err := ListScriptReferences(projectPath)
	if err != nil {
		return nil, err
	}
	pathsByID := make(map[string]string, len(references))
	for _, reference := range references {
		if reference.Tracked {
			pathsByID[reference.ID] = reference.Path
		}
	}
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
	paths := make([]string, 0, len(assetIDs))
	for _, assetID := range assetIDs {
		path, exists := pathsByID[assetID]
		if !exists {
			return nil, fmt.Errorf("script asset %s is not tracked in the configured Scripts directory", assetID)
		}
		asset, err := repository.GetAsset(tx, assetID)
		if err != nil {
			return nil, err
		}
		if asset.FileStatus != "normal" {
			return nil, fmt.Errorf("script asset %s must be current before launch, status is %s", asset.Name, asset.FileStatus)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

type trackedScript struct {
	ID   string
	Name string
}

func trackedScriptsByPath(tx interface {
	Select(interface{}, string, ...interface{}) error
}, workingDir string) (map[string]trackedScript, error) {
	var assets []struct {
		ID             string `db:"id"`
		Name           string `db:"name"`
		Extension      string `db:"extension"`
		CollectionPath string `db:"collection_path"`
		Pointer        string `db:"pointer"`
	}
	query := `
		SELECT
			a.id,
			a.name,
			a.extension,
			a.pointer,
			COALESCE(c.collection_path, '') AS collection_path
		FROM asset a
		LEFT JOIN collection c ON a.collection_id = c.id
		WHERE a.trashed = 0
	`
	if err := tx.Select(&assets, query); err != nil {
		return nil, err
	}
	tracked := make(map[string]trackedScript, len(assets))
	for _, asset := range assets {
		path := asset.Pointer
		if path == "" {
			var err error
			path, err = utils.BuildAssetPath(workingDir, asset.CollectionPath, asset.Name, asset.Extension)
			if err != nil {
				continue
			}
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
