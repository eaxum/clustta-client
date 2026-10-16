package handlers

import (
	"errors"
	"net/http"

	"clustta/services"
)

type ayonWorkfileRequest struct {
	AYONProjectID string `json:"ayon_project_id"`
	FolderID      string `json:"folder_id"`
	TaskID        string `json:"task_id"`
	TaskName      string `json:"task_name"`
	TaskTypeID    string `json:"task_type_id"`
	TaskType      string `json:"task_type"`
	Application   string `json:"application"`
	Extension     string `json:"extension"`
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
