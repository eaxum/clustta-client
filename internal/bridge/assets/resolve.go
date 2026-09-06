package assets

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// ResolveLocalFiles validates project identity, permissions and stored asset paths without fetching files.
func ResolveLocalFiles(ctx context.Context, projectPath, projectID, userID string, assetIDs []string) ([]string, error) {
	if projectID == "" || userID == "" || len(assetIDs) == 0 {
		return nil, errors.New("select assets in an authenticated project")
	}
	canonical, err := filepath.EvalSymlinks(projectPath)
	if err != nil {
		return nil, fmt.Errorf("project unavailable: %w", err)
	}
	if !filepath.IsAbs(canonical) {
		return nil, errors.New("project path must be absolute")
	}
	uriPath := filepath.ToSlash(canonical)
	if !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}
	databaseURI := url.URL{Scheme: "file", Path: uriPath, RawQuery: "mode=ro"}
	db, err := sqlx.Open("sqlite3", databaseURI.String())
	if err != nil {
		return nil, err
	}
	defer db.Close()
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	storedID, err := utils.GetProjectId(tx)
	if err != nil {
		return nil, err
	}
	if storedID != projectID {
		return nil, errors.New("project identity does not match")
	}
	root, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		return nil, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("working directory unavailable: %w", err)
	}
	var permissions struct {
		ViewAsset bool `db:"view_asset"`
	}
	err = tx.GetContext(ctx, &permissions, `SELECT role.view_asset
		FROM user JOIN role ON role.id = user.role_id WHERE user.id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("project access unavailable: %w", err)
	}
	visible := make(map[string]bool)
	if !permissions.ViewAsset {
		assets, err := repository.GetUserAssetsMinimal(tx, userID)
		if err != nil {
			return nil, err
		}
		for _, asset := range assets {
			visible[asset.Id] = true
		}
	}
	paths := make([]string, 0, len(assetIDs))
	seen := make(map[string]bool)
	for _, id := range assetIDs {
		if seen[id] {
			continue
		}
		seen[id] = true
		var asset models.Asset
		err := tx.GetContext(ctx, &asset, `SELECT id, name, extension, collection_path, is_link, pointer,
			trashed, local_path FROM full_asset WHERE id = ?`, id)
		if err != nil {
			return nil, fmt.Errorf("asset %s unavailable: %w", id, err)
		}
		if !permissions.ViewAsset && !visible[id] {
			return nil, fmt.Errorf("you do not have permission to export %s", asset.Name)
		}
		asset.FilePath, err = utils.BuildAssetPath(root, asset.CollectionPath, asset.Name, asset.Extension)
		if err != nil {
			return nil, err
		}
		path, err := localAssetPath(root, asset)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", asset.Name, err)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func localAssetPath(root string, asset models.Asset) (string, error) {
	if asset.IsLink || asset.Pointer != "" || asset.Trashed {
		return "", errors.New("only regular, non-trashed project assets can be dragged")
	}
	path := asset.GetFilePath()
	if !filepath.IsAbs(root) || !filepath.IsAbs(path) {
		return "", errors.New("asset has no valid local path")
	}
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", errors.New("local file unavailable; download the asset first")
	}
	relative, err := filepath.Rel(root, canonical)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("asset path is outside its project working directory")
	}
	file, err := os.Open(canonical)
	if err != nil {
		return "", fmt.Errorf("cannot read local file: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("asset is not a regular file")
	}
	return canonical, nil
}
