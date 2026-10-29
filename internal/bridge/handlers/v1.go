package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/project_index"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/settings"
	"clustta/internal/utils"
	"clustta/services"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	bridgeAPIVersion      = "1"
	trackJobStartProgress = 5
	trackJobSyncProgress  = 90
)

type capabilitiesResponse struct {
	APIVersion string   `json:"api_version"`
	Operations []string `json:"operations"`
}

type bootstrapResponse struct {
	APIVersion    string               `json:"api_version"`
	Accounts      []accountResponse    `json:"accounts"`
	Studios       []studioResponse     `json:"studios"`
	Projects      []dccProjectResponse `json:"projects"`
	ActiveAccount *accountResponse     `json:"active_account,omitempty"`
	ActiveStudio  string               `json:"active_studio"`
}

type workspaceResponse struct {
	Statuses []models.Status `json:"statuses"`
	Assets   []assetResponse `json:"assets"`
}

type contextResponse struct {
	Studio  string             `json:"studio"`
	Project dccProjectResponse `json:"project"`
	Asset   *assetResponse     `json:"asset"`
}

type getLatestRequest struct {
	AllowModified bool `json:"allow_modified"`
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
	SourceAssetID      string `json:"sourceAssetId"`
	SourceCheckpointID string `json:"sourceCheckpointId"`
	FilePath           string `json:"filePath"`
	Message            string `json:"message"`
	PreviewPath        string `json:"previewPath"`
	UseAsThumbnail     bool   `json:"useAsThumbnail"`
	Sync               *bool  `json:"sync"`
}

type trackAssetRequest struct {
	SourceAssetID       string `json:"sourceAssetId"`
	SourceCheckpointID  string `json:"sourceCheckpointId"`
	FilePath            string `json:"filePath"`
	Message             string `json:"message"`
	PreviewPath         string `json:"previewPath"`
	AssignToCurrentUser bool   `json:"assignToCurrentUser"`
	Sync                *bool  `json:"sync"`
}

type trackAssetResponse struct {
	Asset      assetResponse      `json:"asset"`
	Checkpoint checkpointResponse `json:"checkpoint"`
	Created    bool               `json:"created"`
}

type statusRequest struct {
	StatusID string `json:"statusId"`
}

type revertRequest struct {
	CheckpointID string `json:"checkpointId"`
}

type dependencyRequest struct {
	DependencyID         string `json:"dependency_id"`
	ResolutionMode       string `json:"resolution_mode"`
	CheckpointID         string `json:"checkpoint_id"`
	AssetCheckpointTagID string `json:"asset_checkpoint_tag_id"`
}

type dependencyResponse struct {
	models.AssetDependencyEdge
	Dependency assetResponse `json:"dependency"`
}

type dependencyReference struct {
	FilePath string `json:"file_path"`
	Kind     string `json:"kind"`
}

type importDependenciesRequest struct {
	References []dependencyReference `json:"references"`
}

type importDependenciesResult struct {
	Added      []string `json:"added"`
	Existing   []string `json:"existing"`
	Unresolved []string `json:"unresolved"`
}

type dependencySelectorRequest struct {
	ResolutionMode       string `json:"resolution_mode"`
	CheckpointID         string `json:"checkpoint_id"`
	AssetCheckpointTagID string `json:"asset_checkpoint_tag_id"`
}

type checkpointTagRequest struct {
	Name string `json:"name"`
}

type dependencyBuildRequest struct {
	PlanFingerprint string `json:"plan_fingerprint"`
	AllowModified   bool   `json:"allow_modified"`
}

func V1Capabilities(w http.ResponseWriter, _ *http.Request) {
	jsonResponse(w, http.StatusOK, capabilitiesResponse{
		APIVersion: bridgeAPIVersion,
		Operations: []string{
			"dcc.bootstrap",
			"project.workspace",
			"assets.list",
			"assets.context",
			"assets.track",
			"ayon.workfiles.resolve",
			"ayon.workfiles.prepare",
			"ayon.workfiles.import",
			"assets.status",
			"assets.open",
			"assets.reveal",
			"assets.dependencies",
			"assets.dependencies.import",
			"assets.dependency_selectors",
			"assets.build_plan",
			"assets.build",
			"assets.sync",
			"assets.get_latest",
			"checkpoints.list",
			"checkpoints.create",
			"checkpoints.sources",
			"checkpoints.revert",
			"checkpoint_tags.manage",
			"jobs.get",
			"jobs.cancel",
		},
	})
}

func V1TrackAsset(w http.ResponseWriter, r *http.Request) {
	project, ok := requestProject(w, r)
	if !ok {
		return
	}
	var body trackAssetRequest
	if err := decodeBody(r, &body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(body.FilePath) == "" {
		jsonError(w, http.StatusBadRequest, "filePath is required")
		return
	}
	filePath, ok := validateTrackAssetPath(w, project.WorkingDirectory, body.FilePath)
	if !ok {
		return
	}

	checkpointService := &services.CheckpointService{}
	sourceID, err := checkpointService.ResolveCheckpointSource(
		project.Uri,
		body.SourceAssetID,
		body.SourceCheckpointID,
	)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
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

	groupID := uuid.NewString()
	idempotencyKey, ok := requiredOperationKey(w, r, project.Id, filePath, "track-asset")
	if !ok {
		return
	}
	job := startJob("track-asset", idempotencyKey, false, func(update func(string, int)) (any, error) {
		return runTrackAssetJob(
			project,
			body,
			filePath,
			sourceID,
			groupID,
			syncAfterCheckpoint,
			update,
		)
	})
	jsonResponse(w, http.StatusAccepted, job)
}

func validateTrackAssetPath(
	w http.ResponseWriter,
	workingDirectory string,
	requestedPath string,
) (string, bool) {
	filePath := canonicalPath(requestedPath)
	if !pathWithin(filePath, workingDirectory) {
		jsonError(w, http.StatusBadRequest, "filePath must be inside the project working directory")
		return "", false
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "asset file is not available on disk")
		return "", false
	}
	if !fileInfo.Mode().IsRegular() || fileInfo.Size() == 0 {
		jsonError(w, http.StatusBadRequest, "asset file must be a non-empty regular file")
		return "", false
	}
	return filePath, true
}

func runTrackAssetJob(
	project repository.ProjectInfo,
	request trackAssetRequest,
	filePath string,
	sourceID string,
	groupID string,
	syncAfterCheckpoint bool,
	update func(string, int),
) (trackAssetResponse, error) {
	update("Tracking asset", trackJobStartProgress)
	snapshotPath, err := createStableFileSnapshot(filePath)
	if err != nil {
		return trackAssetResponse{}, err
	}
	defer os.Remove(snapshotPath)
	result, err := services.TrackAsset(
		project.Uri,
		services.TrackAssetOptions{
			FilePath:            filePath,
			ContentPath:         snapshotPath,
			Message:             request.Message,
			PreviewPath:         request.PreviewPath,
			GroupID:             groupID,
			SourceCheckpointID:  sourceID,
			AssignToCurrentUser: request.AssignToCurrentUser,
		},
		func(progress int) {
			update("Tracking asset", progress)
		},
	)
	if err != nil {
		return trackAssetResponse{}, err
	}
	if err = syncTrackedAsset(project.Uri, result.Asset.Id, syncAfterCheckpoint, update); err != nil {
		return trackAssetResponse{}, err
	}
	return trackAssetResponse{
		Asset:      assetToResponse(result.Asset),
		Checkpoint: checkpointToResponse(result.Checkpoint),
		Created:    result.Created,
	}, nil
}

func syncTrackedAsset(
	projectPath string,
	assetID string,
	syncAfterCheckpoint bool,
	update func(string, int),
) error {
	if !syncAfterCheckpoint {
		return nil
	}
	update("Synchronizing asset", trackJobSyncProgress)
	remoteURL, err := utils.ResolveProjectRemoteURL(projectPath)
	if err != nil || remoteURL == "" {
		return err
	}
	syncService := &services.SyncService{}
	return syncService.SyncAsset(projectPath, remoteURL, assetID)
}

func V1Bootstrap(w http.ResponseWriter, r *http.Request) {
	refresh := refreshRequested(r)
	accounts, err := auth_service.GetAllAccounts()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	accountItems := make([]accountResponse, 0, len(accounts))
	for id, token := range accounts {
		accountItems = append(accountItems, accountResponse{
			ID:        id,
			Email:     token.User.Email,
			Username:  token.User.Username,
			FirstName: token.User.FirstName,
			LastName:  token.User.LastName,
		})
	}

	studios, err := listStudiosCached(refresh)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	activeStudio, err := studioForRequest(r)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	projects, err := getNamedStudioProjects(activeStudio, refresh)
	if err != nil {
		projects = nil
	}
	projectItems := make([]dccProjectResponse, 0, len(projects))
	for _, project := range projects {
		projectItems = append(projectItems, projectToDCCResponse(project))
	}

	var activeAccount *accountResponse
	if token, activeErr := auth_service.GetActiveAccount(); activeErr == nil {
		activeAccount = &accountResponse{
			ID:        token.User.Id,
			Email:     token.User.Email,
			Username:  token.User.Username,
			FirstName: token.User.FirstName,
			LastName:  token.User.LastName,
		}
	}
	jsonResponse(w, http.StatusOK, bootstrapResponse{
		APIVersion:    bridgeAPIVersion,
		Accounts:      accountItems,
		Studios:       studiosToResponse(studios),
		Projects:      projectItems,
		ActiveAccount: activeAccount,
		ActiveStudio:  activeStudio,
	})
}

func V1ListProjects(w http.ResponseWriter, r *http.Request) {
	studio, err := studioForRequest(r)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	projects, err := getNamedStudioProjects(studio, refreshRequested(r))
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

func V1ProjectWorkspace(w http.ResponseWriter, r *http.Request) {
	project, ok := requestProject(w, r)
	if !ok {
		return
	}
	user, err := auth_service.GetActiveUser()
	if err != nil {
		jsonError(w, http.StatusUnauthorized, err.Error())
		return
	}
	workspace, err := loadProjectWorkspace(
		project,
		user.Id,
		strings.ToLower(r.URL.Query().Get("ext")),
	)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, workspace)
}

func loadProjectWorkspace(
	project repository.ProjectInfo,
	userID string,
	extension string,
) (workspaceResponse, error) {
	dbConn, err := sqlx.Connect("sqlite3", project.Uri)
	if err != nil {
		return workspaceResponse{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return workspaceResponse{}, err
	}
	defer tx.Rollback()

	statuses, err := repository.GetStatuses(tx)
	if err != nil {
		return workspaceResponse{}, err
	}
	assets, err := repository.GetUserAssets(tx, userID)
	if err != nil {
		return workspaceResponse{}, err
	}
	result := make([]assetResponse, 0, len(assets))
	for _, asset := range assets {
		if asset.AssigneeId != userID {
			continue
		}
		if extension != "" && strings.ToLower(asset.Extension) != extension {
			continue
		}
		result = append(result, assetToResponse(asset))
	}
	return workspaceResponse{Statuses: statuses, Assets: result}, nil
}

func V1ResolveContext(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("filePath")
	if filePath == "" {
		jsonError(w, http.StatusBadRequest, "filePath is required")
		return
	}

	entry, project, canonicalFile, found, err := resolveContextProject(filePath)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !found {
		jsonError(w, http.StatusNotFound, "file is not inside a tracked Clustta project")
		return
	}

	asset, found, err := findAssetByFilePath(
		project.Uri,
		project.WorkingDirectory,
		canonicalFile,
	)
	if err != nil {
		if errors.Is(err, error_service.ErrNotUnauthorized) {
			jsonError(w, http.StatusForbidden, "asset is outside the active user's visible scope")
			return
		}
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response := contextResponse{
		Studio:  entry.StudioName,
		Project: projectToDCCResponse(project),
	}
	if found {
		assetResponse := assetToResponse(asset)
		response.Asset = &assetResponse
	}
	jsonResponse(w, http.StatusOK, response)
}

func resolveContextProject(
	filePath string,
) (projectindex.Entry, repository.ProjectInfo, string, bool, error) {
	entry, canonicalFile, found, err := projectindex.ResolveCanonical(filePath)
	if err != nil {
		return projectindex.Entry{}, repository.ProjectInfo{}, "", false, err
	}
	if !found {
		if err = projectindex.EnsureLocalProjects(); err != nil {
			return projectindex.Entry{}, repository.ProjectInfo{}, "", false, err
		}
		entry, canonicalFile, found, err = projectindex.ResolveCanonical(canonicalFile)
		if err != nil || !found {
			return projectindex.Entry{}, repository.ProjectInfo{}, canonicalFile, found, err
		}
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return projectindex.Entry{}, repository.ProjectInfo{}, canonicalFile, false, err
	}
	if _, err = os.Stat(entry.DatabasePath); os.IsNotExist(err) {
		if removeErr := projectindex.RemoveProject(entry.ProjectID); removeErr != nil {
			log.Printf("Failed to remove missing project index entry: %v", removeErr)
		}
		return projectindex.Entry{}, repository.ProjectInfo{}, canonicalFile, false, nil
	}
	project, err := repository.GetProjectInfo(entry.DatabasePath, user)
	if err != nil {
		return projectindex.Entry{}, repository.ProjectInfo{}, canonicalFile, false, err
	}
	projectWorkingDirectory := canonicalPath(project.WorkingDirectory)
	if !pathWithinCanonical(canonicalFile, projectWorkingDirectory) {
		if removeErr := projectindex.RemoveProject(entry.ProjectID); removeErr != nil {
			log.Printf("Failed to remove stale project index entry: %v", removeErr)
		}
		return projectindex.Entry{}, repository.ProjectInfo{}, canonicalFile, false, nil
	}
	if project.Id != entry.ProjectID ||
		!sameCanonicalPath(projectWorkingDirectory, entry.WorkingDirectory) {
		if project.Id != entry.ProjectID {
			if removeErr := projectindex.RemoveProject(entry.ProjectID); removeErr != nil {
				log.Printf("Failed to remove replaced project index entry: %v", removeErr)
			}
		}
		entry.ProjectID = project.Id
		entry.ProjectName = project.Name
		entry.WorkingDirectory = projectWorkingDirectory
		if upsertErr := projectindex.Upsert(entry); upsertErr != nil {
			log.Printf("Failed to repair project index entry: %v", upsertErr)
		}
	}
	project.Uri = entry.DatabasePath
	project.WorkingDirectory = projectWorkingDirectory
	project.IsDownloaded = true
	project.IsTracked = true
	return entry, project, canonicalFile, true, nil
}

func V1SyncAsset(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}
	idempotencyKey, ok := requiredOperationKey(w, r, project.Id, asset.Id, "sync")
	if !ok {
		return
	}
	job := startJob("sync", idempotencyKey, true, func(update func(string, int)) (any, error) {
		update("Synchronizing asset", 10)
		remoteURL, err := utils.ResolveProjectRemoteURL(project.Uri)
		if err != nil {
			return nil, err
		}
		if remoteURL == "" {
			return nil, errors.New("project does not have a remote Studio")
		}
		syncService := &services.SyncService{}
		if err = syncService.SyncAsset(project.Uri, remoteURL, asset.Id); err != nil {
			return nil, err
		}
		emitAssetDataChanged(
			project,
			asset.Id,
			services.ProjectChangeAssetState,
			services.ProjectChangeSync,
		)
		update("Asset synchronized", 100)
		return map[string]string{"asset_id": asset.Id}, nil
	})
	jsonResponse(w, http.StatusAccepted, job)
}

func V1GetLatestAsset(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}
	body := getLatestRequest{}
	if err := decodeBody(r, &body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	idempotencyKey, ok := requiredOperationKey(w, r, project.Id, asset.Id, "get-latest")
	if !ok {
		return
	}
	job := startJob("get-latest", idempotencyKey, true, func(update func(string, int)) (any, error) {
		remoteURL, err := utils.ResolveProjectRemoteURL(project.Uri)
		if err != nil {
			return nil, err
		}
		if remoteURL == "" {
			return nil, errors.New("project does not have a remote Studio")
		}
		update("Refreshing project metadata", 10)
		syncService := &services.SyncService{}
		if err = syncService.UpdateProject(project.Uri, remoteURL); err != nil {
			return nil, err
		}
		update("Restoring latest asset", 50)
		checkpointService := &services.CheckpointService{}
		assetIDs := []string{asset.Id}
		var result services.FetchResult
		if body.AllowModified {
			result, err = checkpointService.Revert(project.Uri, remoteURL, assetIDs)
		} else {
			result, err = checkpointService.RevertUnmodified(project.Uri, remoteURL, assetIDs)
		}
		if err != nil {
			return nil, err
		}
		return result, nil
	})
	jsonResponse(w, http.StatusAccepted, job)
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
	edges, err := assetService.GetAssetDependencyEdges(project.Uri, asset.Id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result := make([]dependencyResponse, 0, len(edges))
	for _, edge := range edges {
		dependency, dependencyErr := assetService.GetAssetByID(project.Uri, edge.DependencyId)
		if dependencyErr != nil {
			jsonError(w, http.StatusInternalServerError, dependencyErr.Error())
			return
		}
		result = append(result, dependencyResponse{
			AssetDependencyEdge: edge,
			Dependency:          assetToResponse(dependency),
		})
	}
	jsonResponse(w, http.StatusOK, result)
}

func V1ListDependencyTypes(w http.ResponseWriter, r *http.Request) {
	project, ok := requestProject(w, r)
	if !ok {
		return
	}
	dependencyTypeService := &services.DependencyTypeService{}
	dependencyTypes, err := dependencyTypeService.GetDependencyTypes(project.Uri)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, dependencyTypes)
}

func V1ImportDependencies(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}
	body := importDependenciesRequest{}
	if err := decodeBody(r, &body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(body.References) == 0 {
		jsonError(w, http.StatusBadRequest, "at least one reference is required")
		return
	}

	idempotencyKey := operationKey(r, project.Id, asset.Id, "import-dependencies")
	job := startJob("import-dependencies", idempotencyKey, false, func(update func(string, int)) (any, error) {
		assetService := &services.AssetService{}
		edges, err := assetService.GetAssetDependencyEdges(project.Uri, asset.Id)
		if err != nil {
			return nil, err
		}
		existing := make(map[string]bool, len(edges))
		for _, edge := range edges {
			existing[edge.DependencyId] = true
		}
		result := importDependenciesResult{
			Added:      []string{},
			Existing:   []string{},
			Unresolved: []string{},
		}
		for index, reference := range body.References {
			update("Recording Blender references", 5+(index*90/len(body.References)))
			filePath := canonicalPath(reference.FilePath)
			if !pathWithin(filePath, project.WorkingDirectory) {
				result.Unresolved = append(result.Unresolved, reference.FilePath)
				continue
			}
			dependency, found, findErr := findAssetByFilePath(
				project.Uri,
				project.WorkingDirectory,
				filePath,
			)
			if findErr != nil {
				return nil, findErr
			}
			if !found || dependency.Id == asset.Id {
				result.Unresolved = append(result.Unresolved, reference.FilePath)
				continue
			}
			if existing[dependency.Id] {
				result.Existing = append(result.Existing, dependency.Id)
				continue
			}
			if _, err = assetService.AddAssetDependencyWithSelector(
				project.Uri,
				asset.Id,
				dependency.Id,
				repository.DependencyResolutionFloating,
				"",
				"",
			); err != nil {
				return nil, err
			}
			existing[dependency.Id] = true
			result.Added = append(result.Added, dependency.Id)
		}
		return result, nil
	})
	jsonResponse(w, http.StatusAccepted, job)
}

func V1CreateDependency(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}
	request := dependencyRequest{}
	if err := decodeBody(r, &request); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.DependencyID == "" || request.ResolutionMode == "" {
		jsonError(w, http.StatusBadRequest, "dependency_id and resolution_mode are required")
		return
	}

	assetService := &services.AssetService{}
	edge, err := assetService.AddAssetDependencyWithSelector(
		project.Uri,
		asset.Id,
		request.DependencyID,
		request.ResolutionMode,
		request.CheckpointID,
		request.AssetCheckpointTagID,
	)
	if err != nil {
		dependencyMutationError(w, err)
		return
	}
	jsonResponse(w, http.StatusCreated, edge)
}

func V1UpdateDependencySelector(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}
	request := dependencySelectorRequest{}
	if err := decodeBody(r, &request); err != nil || request.ResolutionMode == "" {
		jsonError(w, http.StatusBadRequest, "resolution_mode is required")
		return
	}

	assetService := &services.AssetService{}
	edge, err := assetService.UpdateAssetDependencySelector(
		project.Uri,
		asset.Id,
		r.PathValue("edgeId"),
		request.ResolutionMode,
		request.CheckpointID,
		request.AssetCheckpointTagID,
	)
	if err != nil {
		dependencyMutationError(w, err)
		return
	}
	jsonResponse(w, http.StatusOK, edge)
}

func V1DependencySelectorOptions(w http.ResponseWriter, r *http.Request) {
	project, _, ok := requestAsset(w, r)
	if !ok {
		return
	}
	assetService := &services.AssetService{}
	options, err := assetService.GetDependencySelectorOptions(project.Uri, r.PathValue("dependencyId"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, options)
}

func V1ListCheckpointTags(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}
	checkpointService := &services.CheckpointService{}
	tags, err := checkpointService.GetCheckpointTags(project.Uri, asset.Id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, tags)
}

func V1SetCheckpointTag(w http.ResponseWriter, r *http.Request) {
	project, ok := requestProject(w, r)
	if !ok {
		return
	}
	request := checkpointTagRequest{}
	if err := decodeBody(r, &request); err != nil || strings.TrimSpace(request.Name) == "" {
		jsonError(w, http.StatusBadRequest, "name is required")
		return
	}
	checkpointService := &services.CheckpointService{}
	tag, err := checkpointService.SetCheckpointTag(
		project.Uri,
		r.PathValue("tagId"),
		request.Name,
		r.PathValue("checkpointId"),
	)
	if err != nil {
		dependencyMutationError(w, err)
		return
	}
	jsonResponse(w, http.StatusOK, tag)
}

func V1SetCheckpointTagsForGroup(w http.ResponseWriter, r *http.Request) {
	project, ok := requestProject(w, r)
	if !ok {
		return
	}
	request := checkpointTagRequest{}
	if err := decodeBody(r, &request); err != nil || strings.TrimSpace(request.Name) == "" {
		jsonError(w, http.StatusBadRequest, "name is required")
		return
	}
	checkpointService := &services.CheckpointService{}
	tags, err := checkpointService.SetCheckpointTagsForGroup(project.Uri, request.Name, r.PathValue("groupId"))
	if err != nil {
		dependencyMutationError(w, err)
		return
	}
	jsonResponse(w, http.StatusOK, tags)
}

func V1DeleteCheckpointTag(w http.ResponseWriter, r *http.Request) {
	project, ok := requestProject(w, r)
	if !ok {
		return
	}
	checkpointService := &services.CheckpointService{}
	if err := checkpointService.DeleteCheckpointTag(project.Uri, r.PathValue("tagId")); err != nil {
		dependencyMutationError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func dependencyMutationError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, error_service.ErrNotUnauthorized) {
		status = http.StatusForbidden
	}
	jsonError(w, status, err.Error())
}

func V1ListCheckpoints(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
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

	checkpoints, err := repository.GetCheckpoints(tx, asset.Id, false)
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

	checkpointService := &services.CheckpointService{}
	sourceID, err := checkpointService.ResolveCheckpointSource(project.Uri, body.SourceAssetID, body.SourceCheckpointID)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	idempotencyKey, ok := requiredOperationKey(w, r, project.Id, asset.Id, "checkpoint")
	if !ok {
		return
	}
	job := startJob("checkpoint", idempotencyKey, false, func(update func(string, int)) (any, error) {
		update("Creating checkpoint", 20)
		snapshotPath, snapshotErr := createStableFileSnapshot(asset.GetFilePath())
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		defer os.Remove(snapshotPath)
		checkpointService := &services.CheckpointService{}
		checkpoints, addErr := checkpointService.AddCheckpointFromFileWithSource(
			project.Uri,
			asset.AssetPath,
			asset.Extension,
			snapshotPath,
			body.Message,
			body.PreviewPath,
			uuid.NewString(),
			body.UseAsThumbnail,
			true,
			sourceID,
		)
		if addErr != nil {
			return nil, addErr
		}
		emitAssetDataChanged(
			project,
			asset.Id,
			services.ProjectChangeAssetState,
			services.ProjectChangeCheckpoints,
		)
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
				emitAssetDataChanged(
					project,
					asset.Id,
					services.ProjectChangeAssetState,
					services.ProjectChangeSync,
				)
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

	idempotencyKey, ok := requiredOperationKey(w, r, project.Id, asset.Id, "status:"+body.StatusID)
	if !ok {
		return
	}
	job := startJob("status", idempotencyKey, false, func(update func(string, int)) (any, error) {
		update("Updating status", 25)
		assetService := &services.AssetService{}
		result, changeErr := assetService.ChangeStatus(
			project.Uri,
			[]string{asset.Id},
			body.StatusID,
		)
		if changeErr != nil {
			return nil, changeErr
		}
		emitAssetDataChanged(project, asset.Id, services.ProjectChangeStatus)
		return result, nil
	})
	jsonResponse(w, http.StatusAccepted, job)
}

func V1OpenAsset(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}
	studio, err := studioForRequest(r)
	if err != nil || studio == "" {
		jsonError(w, http.StatusConflict, "active studio is unavailable")
		return
	}
	if err := services.DispatchDeepLink(
		assetDeepLink(studio, project.Id, asset.Id),
	); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func V1RevealAsset(w http.ResponseWriter, r *http.Request) {
	_, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}
	filePath := asset.GetFilePath()
	if _, err := os.Stat(filePath); err != nil {
		jsonError(w, http.StatusConflict, "asset file is not available on disk")
		return
	}
	utils.RevealInExplorer(filePath)
	w.WriteHeader(http.StatusNoContent)
}

func V1BuildAsset(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}
	var body dependencyBuildRequest
	if err := decodeBody(r, &body); err != nil || body.PlanFingerprint == "" {
		jsonError(w, http.StatusBadRequest, "plan_fingerprint is required")
		return
	}

	idempotencyKey := operationKey(r, project.Id, asset.Id, "build:"+body.PlanFingerprint)
	job := startJob("build", idempotencyKey, true, func(update func(string, int)) (any, error) {
		update("Revalidating dependency plan", 10)
		remoteURL, err := utils.ResolveProjectRemoteURL(project.Uri)
		if err != nil {
			return nil, err
		}
		buildResult, err := services.ExecuteDependencyBuildPlanDirect(
			project.Uri,
			remoteURL,
			asset.Id,
			body.PlanFingerprint,
			body.AllowModified,
		)
		if err != nil {
			return nil, err
		}
		return buildResult, nil
	})
	jsonResponse(w, http.StatusAccepted, job)
}

func V1DependencyBuildPlan(w http.ResponseWriter, r *http.Request) {
	project, asset, ok := requestAsset(w, r)
	if !ok {
		return
	}
	assetService := &services.AssetService{}
	plan, err := assetService.ResolveDependencyBuildPlan(project.Uri, asset.Id)
	if err != nil {
		if errors.Is(err, error_service.ErrNotUnauthorized) {
			jsonError(w, http.StatusForbidden, err.Error())
			return
		}
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, plan)
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
	studio, err := studioForRequest(r)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return repository.ProjectInfo{}, false
	}
	project, err := resolveProject(r.PathValue("projectId"), studio)
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

func studioForRequest(r *http.Request) (string, error) {
	studio := strings.TrimSpace(r.Header.Get("X-Clustta-Studio"))
	if studio != "" {
		return studio, nil
	}
	return settings.GetLastStudio()
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
	asset, err := assetService.GetVisibleAssetByID(project.Uri, r.PathValue("assetId"))
	if err != nil {
		if errors.Is(err, error_service.ErrNotUnauthorized) {
			jsonError(w, http.StatusForbidden, "asset is outside the active user's visible scope")
			return repository.ProjectInfo{}, models.Asset{}, false
		}
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

func emitAssetDataChanged(
	project repository.ProjectInfo,
	assetID string,
	changes ...string,
) {
	services.EmitProjectDataChanged(services.ProjectDataChange{
		ProjectID:  project.Id,
		ProjectURI: project.Uri,
		AssetIDs:   []string{assetID},
		Changes:    changes,
	})
}

func findAssetByFilePath(
	projectPath string,
	workingDirectory string,
	filePath string,
) (models.Asset, bool, error) {
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
	assetPath, extension, relativeErr := assetLookupPath(workingDirectory, filePath)
	if relativeErr == nil {
		assetID, directFound, directErr := repository.GetAssetIDByPath(
			tx,
			assetPath,
			extension,
		)
		if directErr != nil {
			return models.Asset{}, false, directErr
		}
		if directFound {
			return visibleAssetByID(projectPath, assetID)
		}
	}

	assets, err := repository.GetAssets(tx, false)
	if err != nil {
		return models.Asset{}, false, err
	}
	for _, asset := range assets {
		if samePath(asset.GetFilePath(), filePath) {
			return visibleAssetByID(projectPath, asset.Id)
		}
	}
	return models.Asset{}, false, nil
}

func visibleAssetByID(projectPath, assetID string) (models.Asset, bool, error) {
	assetService := &services.AssetService{}
	asset, err := assetService.GetVisibleAssetByID(projectPath, assetID)
	return asset, err == nil, err
}

func assetLookupPath(workingDirectory, filePath string) (string, string, error) {
	relativePath, err := filepath.Rel(workingDirectory, filePath)
	if err != nil {
		return "", "", err
	}
	extension := filepath.Ext(relativePath)
	assetPath := strings.TrimSuffix(filepath.ToSlash(relativePath), extension)
	return "/" + strings.TrimPrefix(assetPath, "/"), extension, nil
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

func requiredOperationKey(
	w http.ResponseWriter,
	r *http.Request,
	projectID string,
	assetID string,
	operation string,
) (string, bool) {
	key := operationKey(r, projectID, assetID, operation)
	if key == "" {
		jsonError(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return "", false
	}
	return key, true
}

func assetDeepLink(studio, projectID, assetID string) string {
	query := url.Values{}
	query.Set("studio", studio)
	query.Set("project", projectID)
	query.Set("asset", assetID)
	return (&url.URL{
		Scheme:   "clustta",
		Host:     "open",
		RawQuery: query.Encode(),
	}).String()
}

func pathWithin(filePath, rootPath string) bool {
	fileCanonical := canonicalPath(filePath)
	rootCanonical := canonicalPath(rootPath)
	return pathWithinCanonical(fileCanonical, rootCanonical)
}

func pathWithinCanonical(filePath, rootPath string) bool {
	relative, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func samePath(firstPath, secondPath string) bool {
	first := canonicalPath(firstPath)
	second := canonicalPath(secondPath)
	return sameCanonicalPath(first, second)
}

func sameCanonicalPath(firstPath, secondPath string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(firstPath, secondPath)
	}
	return firstPath == secondPath
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
