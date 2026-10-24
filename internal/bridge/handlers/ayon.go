package handlers

import (
	"errors"
	"net/http"

	"clustta/services"
)

type ayonWorkfileRequest struct {
	AYONProjectID string              `json:"ayon_project_id"`
	FolderID      string              `json:"folder_id"`
	FolderPath    string              `json:"folder_path"`
	Folders       []ayonFolderRequest `json:"folders"`
	TaskID        string              `json:"task_id"`
	TaskName      string              `json:"task_name"`
	TaskTypeID    string              `json:"task_type_id"`
	TaskType      string              `json:"task_type"`
	Application   string              `json:"application"`
	Extension     string              `json:"extension"`
}

type ayonFolderRequest struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
}

type ayonWorkfilePreparationResponse struct {
	Asset          *assetResponse `json:"asset,omitempty"`
	TargetPath     string         `json:"target_path,omitempty"`
	RequiresImport bool           `json:"requires_import"`
	FileExists     bool           `json:"file_exists"`
}

type ayonWorkfileImportRequest struct {
	Context  ayonWorkfileRequest `json:"context"`
	FilePath string              `json:"file_path"`
}

type ayonWorkfileImportResponse struct {
	Asset      assetResponse      `json:"asset"`
	Checkpoint checkpointResponse `json:"checkpoint"`
	Created    bool               `json:"created"`
}

type ayonWorkfileResponse struct {
	Project dccProjectResponse `json:"project"`
	Asset   assetResponse      `json:"asset"`
	Created bool               `json:"created"`
}

// V1ResolveAYONWorkfile resolves or lazily creates a mapped AYON task workfile.
func V1ResolveAYONWorkfile(w http.ResponseWriter, r *http.Request) {
	project, ok := requestProject(w, r)
	if !ok {
		return
	}

	var body ayonWorkfileRequest
	if err := decodeBody(r, &body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	integrationService := &services.IntegrationService{}
	result, err := integrationService.ResolveOrCreateAYONWorkfile(
		project.Uri,
		services.AYONWorkfileContext{
			ProjectID:   body.AYONProjectID,
			FolderID:    body.FolderID,
			TaskID:      body.TaskID,
			TaskName:    body.TaskName,
			TaskTypeID:  body.TaskTypeID,
			TaskType:    body.TaskType,
			Application: body.Application,
			Extension:   body.Extension,
		},
	)
	if err != nil {
		status := http.StatusBadRequest
		switch {
		case errors.Is(err, services.ErrAYONIntegrationNotConfigured):
			status = http.StatusConflict
		case errors.Is(err, services.ErrAYONMappingNotConfigured):
			status = http.StatusConflict
		case errors.Is(err, services.ErrAYONPermissionDenied):
			status = http.StatusForbidden
		}
		jsonError(w, status, err.Error())
		return
	}

	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
	}
	jsonResponse(w, status, ayonWorkfileResponse{
		Project: projectToDCCResponse(project),
		Asset:   assetToResponse(result.Asset),
		Created: result.Created,
	})
}

// V1PrepareAYONWorkfile creates hierarchy and returns an existing asset or target path.
func V1PrepareAYONWorkfile(w http.ResponseWriter, r *http.Request) {
	project, ok := requestProject(w, r)
	if !ok {
		return
	}

	var body ayonWorkfileRequest
	if err := decodeBody(r, &body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	integrationService := &services.IntegrationService{}
	result, err := integrationService.PrepareAYONWorkfile(project.Uri, ayonContextFromRequest(body))
	if err != nil {
		writeAYONError(w, err)
		return
	}
	response := ayonWorkfilePreparationResponse{
		TargetPath:     result.TargetPath,
		RequiresImport: result.RequiresImport,
		FileExists:     result.FileExists,
	}
	if result.Asset != nil {
		asset := assetToResponse(*result.Asset)
		response.Asset = &asset
	}
	jsonResponse(w, http.StatusOK, response)
}

// V1ImportAYONWorkfile tracks the DCC-created file and its AYON mapping.
func V1ImportAYONWorkfile(w http.ResponseWriter, r *http.Request) {
	project, ok := requestProject(w, r)
	if !ok {
		return
	}

	var body ayonWorkfileImportRequest
	if err := decodeBody(r, &body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.FilePath == "" {
		jsonError(w, http.StatusBadRequest, "file_path is required")
		return
	}
	context := ayonContextFromRequest(body.Context)
	idempotencyKey := operationKey(r, project.Id, context.TaskID, "ayon-import")
	job := startJob("ayon-import", idempotencyKey, false, func(update func(string, int)) (any, error) {
		update("Importing AYON workfile", 10)
		integrationService := &services.IntegrationService{}
		result, err := integrationService.ImportAYONWorkfile(
			project.Uri,
			context,
			body.FilePath,
			func(progress int) { update("Creating initial checkpoint", progress) },
		)
		if err != nil {
			return nil, err
		}
		return ayonWorkfileImportResponse{
			Asset:      assetToResponse(result.Asset),
			Checkpoint: checkpointToResponse(result.Checkpoint),
			Created:    result.Created,
		}, nil
	})
	jsonResponse(w, http.StatusAccepted, job)
}

func ayonContextFromRequest(body ayonWorkfileRequest) services.AYONWorkfileContext {
	folders := make([]services.AYONFolderContext, 0, len(body.Folders))
	for _, folder := range body.Folders {
		folders = append(folders, services.AYONFolderContext{
			ID:       folder.ID,
			ParentID: folder.ParentID,
			Name:     folder.Name,
			Path:     folder.Path,
			Type:     folder.Type,
		})
	}
	return services.AYONWorkfileContext{
		ProjectID:   body.AYONProjectID,
		FolderID:    body.FolderID,
		FolderPath:  body.FolderPath,
		Folders:     folders,
		TaskID:      body.TaskID,
		TaskName:    body.TaskName,
		TaskTypeID:  body.TaskTypeID,
		TaskType:    body.TaskType,
		Application: body.Application,
		Extension:   body.Extension,
	}
}

func writeAYONError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, services.ErrAYONIntegrationNotConfigured):
		status = http.StatusConflict
	case errors.Is(err, services.ErrAYONMappingNotConfigured):
		status = http.StatusConflict
	case errors.Is(err, services.ErrAYONPermissionDenied):
		status = http.StatusForbidden
	}
	jsonError(w, status, err.Error())
}
