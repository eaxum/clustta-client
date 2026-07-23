package handlers

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/settings"
	"clustta/internal/utils"
	"clustta/services"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const bridgeAPIVersion = "1"

type capabilitiesResponse struct {
	APIVersion string   `json:"api_version"`
	Operations []string `json:"operations"`
}

type contextResponse struct {
	Project dccProjectResponse `json:"project"`
	Asset   assetResponse      `json:"asset"`
}

type dccProjectResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	WorkingDirectory string `json:"working_directory"`
	IsDownloaded     bool   `json:"is_downloaded"`
	IsTracked        bool   `json:"is_tracked"`
	HasRemote        bool   `json:"has_remote"`
}

type checkpointRequest struct {
	FilePath       string `json:"filePath"`
	Message        string `json:"message"`
	PreviewPath    string `json:"previewPath"`
	UseAsThumbnail bool   `json:"useAsThumbnail"`
	Sync           *bool  `json:"sync"`
}

type statusRequest struct {
	StatusID string `json:"statusId"`
}

type revertRequest struct {
	CheckpointID string `json:"checkpointId"`
}

func V1Capabilities(w http.ResponseWriter, _ *http.Request) {
	jsonResponse(w, http.StatusOK, capabilitiesResponse{
		APIVersion: bridgeAPIVersion,
		Operations: []string{
			"assets.list",
			"assets.context",
			"assets.status",
			"assets.dependencies",
			"assets.build",
			"checkpoints.list",
			"checkpoints.create",
			"checkpoints.revert",
			"jobs.get",
			"jobs.cancel",
		},
	})
}

func V1ListProjects(w http.ResponseWriter, _ *http.Request) {
	projects, err := listStudioProjects()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]dccProjectResponse, 0, len(projects))
	for _, project := range projects {
		result = append(result, projectToDCCResponse(project))
	}
	jsonResponse(w, http.StatusOK, result)
}

func V1ResolveContext(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("filePath")
	if filePath == "" {
		jsonError(w, http.StatusBadRequest, "filePath is required")
		return
	}

	projects, err := listStudioProjects()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, project := range projects {
		if !pathWithin(filePath, project.WorkingDirectory) {
			continue
		}
		asset, found, findErr := findAssetByFilePath(project.Uri, filePath)
		if findErr != nil {
			jsonError(w, http.StatusInternalServerError, findErr.Error())
			return
		}
		if found {
			jsonResponse(w, http.StatusOK, contextResponse{
				Project: projectToDCCResponse(project),
				Asset:   assetToResponse(asset),
			})
			return
		}
	}

	jsonError(w, http.StatusNotFound, "file is not a tracked Clustta asset")
}

func V1ListAssignedAssets(w http.ResponseWriter, r *http.Request) {
	project, ok := requestProject(w, r)
	if !ok {
		return
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		jsonError(w, http.StatusUnauthorized, err.Error())
		return
	}

	dbConn, err := sqlx.Connect("sqlite3", project.Uri)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	assets, err := repository.GetUserAssets(tx, user.Id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	extension := strings.ToLower(r.URL.Query().Get("ext"))
	result := make([]assetResponse, 0, len(assets))
	for _, asset := range assets {
		if asset.AssigneeId != user.Id {
			continue
		}
		if extension != "" && strings.ToLower(asset.Extension) != extension {
			continue
		}
		result = append(result, assetToResponse(asset))
	}
	jsonResponse(w, http.StatusOK, result)
}

func V1ListStatuses(w http.ResponseWriter, r *http.Request) {
	project, ok := requestProject(w, r)
	if !ok {
		return
	}

	statusService := &services.StatusService{}
	statuses, err := statusService.GetStatuses(project.Uri)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, statuses)
}

func V1ListDependencies(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}

	assetService := &services.AssetService{}
	assetIDs, err := assetService.ResolveBuildDependencies(project.Uri, asset.Id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]assetResponse, 0, len(assetIDs))
	for _, assetID := range assetIDs {
		if assetID == asset.Id {
			continue
		}
		dependency, getErr := assetService.GetAssetByID(project.Uri, assetID)
		if getErr != nil {
			jsonError(w, http.StatusInternalServerError, getErr.Error())
			return
		}
		result = append(result, assetToResponse(dependency))
	}
	jsonResponse(w, http.StatusOK, result)
}

func V1ListCheckpoints(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}

	checkpointService := &services.CheckpointService{}
	checkpoints, err := checkpointService.GetCheckpoints(project.Uri, asset.Id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]checkpointResponse, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		result = append(result, checkpointToResponse(checkpoint))
	}
	jsonResponse(w, http.StatusOK, result)
}

func V1CreateCheckpoint(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}

	var body checkpointRequest
	if err := decodeBody(r, &body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(body.Message) == "" {
		jsonError(w, http.StatusBadRequest, "message is required")
		return
	}
	if body.FilePath == "" || !samePath(body.FilePath, asset.GetFilePath()) {
		jsonError(w, http.StatusBadRequest, "filePath does not match the tracked asset")
		return
	}
	if _, err := os.Stat(asset.GetFilePath()); err != nil {
		jsonError(w, http.StatusBadRequest, "asset file is not available on disk")
		return
	}
	assetService := &services.AssetService{}
	if err := assetService.AuthorizeCheckpoint(project.Uri, []string{asset.Id}); err != nil {
		jsonError(w, http.StatusForbidden, err.Error())
		return
	}

	syncAfterCheckpoint, err := settings.GetSyncAfterCheckpoint()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if body.Sync != nil {
		syncAfterCheckpoint = *body.Sync
	}

	idempotencyKey := operationKey(r, project.Id, asset.Id, "checkpoint")
	job := startJob("checkpoint", idempotencyKey, false, func(update func(string, int)) (any, error) {
		update("Creating checkpoint", 20)
		checkpointService := &services.CheckpointService{}
		checkpoints, addErr := checkpointService.AddCheckpoint(
			project.Uri,
			[]string{asset.AssetPath},
			[]string{asset.Extension},
			strings.TrimSpace(body.Message),
			body.PreviewPath,
			uuid.NewString(),
			body.UseAsThumbnail,
			true,
		)
		if addErr != nil {
			return nil, addErr
		}
		if syncAfterCheckpoint {
			update("Synchronizing checkpoint", 70)
			remoteURL, resolveErr := utils.ResolveProjectRemoteURL(project.Uri)
			if resolveErr != nil {
				return nil, resolveErr
			}
			if remoteURL != "" {
				syncService := &services.SyncService{}
				if syncErr := syncService.SyncAsset(project.Uri, remoteURL, asset.Id); syncErr != nil {
					return nil, syncErr
				}
			}
		}
		return checkpoints, nil
	})
	jsonResponse(w, http.StatusAccepted, job)
}

func V1ChangeStatus(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}

	var body statusRequest
	if err := decodeBody(r, &body); err != nil || body.StatusID == "" {
		jsonError(w, http.StatusBadRequest, "statusId is required")
		return
	}
	statusService := &services.StatusService{}
	statuses, err := statusService.GetStatuses(project.Uri)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !hasStatus(statuses, body.StatusID) {
		jsonError(w, http.StatusBadRequest, "statusId does not belong to the project")
		return
	}

	idempotencyKey := operationKey(r, project.Id, asset.Id, "status:"+body.StatusID)
	job := startJob("status", idempotencyKey, false, func(update func(string, int)) (any, error) {
		update("Updating status", 25)
		assetService := &services.AssetService{}
		return assetService.ChangeStatus(project.Uri, []string{asset.Id}, body.StatusID)
	})
	jsonResponse(w, http.StatusAccepted, job)
}

func V1BuildAsset(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}
	assetService := &services.AssetService{}
	if err := assetService.AuthorizeRevert(project.Uri, []string{asset.Id}); err != nil {
		jsonError(w, http.StatusForbidden, err.Error())
		return
	}

	idempotencyKey := operationKey(r, project.Id, asset.Id, "build")
	job := startJob("build", idempotencyKey, true, func(update func(string, int)) (any, error) {
		update("Resolving dependencies", 10)
		assetIDs, err := assetService.ResolveBuildDependencies(project.Uri, asset.Id)
		if err != nil {
			return nil, err
		}
		update("Restoring asset files", 25)
		remoteURL, err := utils.ResolveProjectRemoteURL(project.Uri)
		if err != nil {
			return nil, err
		}
		checkpointService := &services.CheckpointService{}
		if err = checkpointService.Revert(project.Uri, remoteURL, assetIDs); err != nil {
			return nil, err
		}
		return map[string]any{"asset_ids": assetIDs}, nil
	})
	jsonResponse(w, http.StatusAccepted, job)
}

func V1RevertAsset(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}

	var body revertRequest
	if err := decodeBody(r, &body); err != nil || body.CheckpointID == "" {
		jsonError(w, http.StatusBadRequest, "checkpointId is required")
		return
	}
	if !assetHasCheckpoint(asset, body.CheckpointID) {
		jsonError(w, http.StatusBadRequest, "checkpoint does not belong to the asset")
		return
	}
	assetService := &services.AssetService{}
	if err := assetService.AuthorizeRevert(project.Uri, []string{asset.Id}); err != nil {
		jsonError(w, http.StatusForbidden, err.Error())
		return
	}

	idempotencyKey := operationKey(r, project.Id, asset.Id, "revert:"+body.CheckpointID)
	job := startJob("revert", idempotencyKey, true, func(update func(string, int)) (any, error) {
		update("Restoring checkpoint", 20)
		remoteURL, err := utils.ResolveProjectRemoteURL(project.Uri)
		if err != nil {
			return nil, err
		}
		checkpointService := &services.CheckpointService{}
		if err = checkpointService.RevertToCheckpoint(
			project.Uri,
			remoteURL,
			asset.Id,
			body.CheckpointID,
		); err != nil {
			return nil, err
		}
		return map[string]string{"file_path": asset.GetFilePath()}, nil
	})
	jsonResponse(w, http.StatusAccepted, job)
}

func V1GetJob(w http.ResponseWriter, r *http.Request) {
	job, err := getJob(r.PathValue("jobId"))
	if err != nil {
		jsonError(w, http.StatusNotFound, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, job)
}

func V1CancelJob(w http.ResponseWriter, r *http.Request) {
	job, err := cancelJob(r.PathValue("jobId"))
	if errors.Is(err, errJobNotFound) {
		jsonError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, errNotCancelable) {
		jsonError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, job)
}

func requestProject(w http.ResponseWriter, r *http.Request) (repository.ProjectInfo, bool) {
	project, err := resolveProject(r.PathValue("projectId"))
	if errors.Is(err, errProjectNotFound) {
		jsonError(w, http.StatusNotFound, err.Error())
		return repository.ProjectInfo{}, false
	}
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return repository.ProjectInfo{}, false
	}
	return project, true
}

func requestAsset(
	w http.ResponseWriter,
	r *http.Request,
) (repository.ProjectInfo, models.Asset, bool) {
	project, ok := requestProject(w, r)
	if !ok {
		return repository.ProjectInfo{}, models.Asset{}, false
	}

	assetService := &services.AssetService{}
	asset, err := assetService.GetAssetByID(project.Uri, r.PathValue("assetId"))
	if err != nil {
		jsonError(w, http.StatusNotFound, err.Error())
		return repository.ProjectInfo{}, models.Asset{}, false
	}
	return project, asset, true
}

func projectToDCCResponse(project repository.ProjectInfo) dccProjectResponse {
	return dccProjectResponse{
		ID:               project.Id,
		Name:             project.Name,
		WorkingDirectory: project.WorkingDirectory,
		IsDownloaded:     project.IsDownloaded,
		IsTracked:        project.IsTracked,
		HasRemote:        project.HasRemote,
	}
}

func findAssetByFilePath(projectPath, filePath string) (models.Asset, bool, error) {
	dbConn, err := sqlx.Connect("sqlite3", projectPath)
	if err != nil {
		return models.Asset{}, false, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return models.Asset{}, false, err
	}
	defer tx.Rollback()

	assets, err := repository.GetAssets(tx, false)
	if err != nil {
		return models.Asset{}, false, err
	}
	for _, asset := range assets {
		if samePath(asset.GetFilePath(), filePath) {
			resolvedAsset, getErr := repository.GetAsset(tx, asset.Id)
			return resolvedAsset, getErr == nil, getErr
		}
	}
	return models.Asset{}, false, nil
}

func assetHasCheckpoint(asset models.Asset, checkpointID string) bool {
	for _, checkpoint := range asset.Checkpoints {
		if checkpoint.Id == checkpointID {
			return true
		}
	}
	return false
}

func hasStatus(statuses []models.Status, statusID string) bool {
	for _, status := range statuses {
		if status.Id == statusID {
			return true
		}
	}
	return false
}

func operationKey(r *http.Request, projectID, assetID, operation string) string {
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		return ""
	}
	return strings.Join([]string{projectID, assetID, operation, key}, ":")
}

func pathWithin(filePath, rootPath string) bool {
	fileCanonical := canonicalPath(filePath)
	rootCanonical := canonicalPath(rootPath)
	relative, err := filepath.Rel(rootCanonical, fileCanonical)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func samePath(firstPath, secondPath string) bool {
	first := canonicalPath(firstPath)
	second := canonicalPath(secondPath)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(first, second)
	}
	return first == second
}

func canonicalPath(filePath string) string {
	absolute, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		absolute = filepath.Clean(filePath)
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		return resolved
	}
	return absolute
}
