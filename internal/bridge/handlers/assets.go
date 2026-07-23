package handlers

import (
	"net/http"

	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"

	"github.com/jmoiron/sqlx"
)

// assetResponse is the JSON shape returned for each asset.
type assetResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Extension      string `json:"extension"`
	CollectionName string `json:"collection_name"`
	CollectionPath string `json:"collection_path"`
	AssetPath      string `json:"asset_path"`
	FilePath       string `json:"file_path"`
	AssigneeID     string `json:"assignee_id"`
	AssigneeName   string `json:"assignee_name"`
	StatusID       string `json:"status_id"`
	StatusName     string `json:"status_short_name"`
	AssetTypeID    string `json:"asset_type_id"`
	AssetTypeName  string `json:"asset_type_name"`
	AssetTypeIcon  string `json:"asset_type_icon"`
	FileStatus     string `json:"file_status"`
	IsResource     bool   `json:"is_resource"`
	PreviewID      string `json:"preview_id"`
}

// ListAssets returns assets for the active project, filtered to the active user.
func ListAssets(w http.ResponseWriter, r *http.Request) {
	activeProjectMu.RLock()
	proj := activeProject
	activeProjectMu.RUnlock()

	if proj == nil {
		jsonError(w, http.StatusBadRequest, "no active project selected")
		return
	}

	ext := r.URL.Query().Get("ext")

	user, err := auth_service.GetActiveUser()
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "no active account: "+err.Error())
		return
	}

	dbConn, err := sqlx.Connect("sqlite3", proj.Uri)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "cannot open project: "+err.Error())
		return
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	// Check user role to decide between all assets or user-assigned assets
	var assets []models.Asset
	userData, err := repository.GetUser(tx, user.Id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "user not in project: "+err.Error())
		return
	}

	userRole, err := repository.GetRole(tx, userData.RoleId)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if userRole.ViewAsset {
		assets, err = repository.GetAssets(tx, false)
	} else {
		assets, err = repository.GetUserAssets(tx, user.Id)
	}
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]assetResponse, 0, len(assets))
	for _, t := range assets {
		// Filter by extension if specified
		if ext != "" && t.Extension != ext {
			continue
		}

		result = append(result, assetToResponse(t))
	}

	jsonResponse(w, http.StatusOK, result)
}

func assetToResponse(asset models.Asset) assetResponse {
	return assetResponse{
		ID:             asset.Id,
		Name:           asset.Name,
		Extension:      asset.Extension,
		CollectionName: asset.CollectionName,
		CollectionPath: asset.CollectionPath,
		AssetPath:      asset.AssetPath,
		FilePath:       asset.GetFilePath(),
		AssigneeID:     asset.AssigneeId,
		AssigneeName:   asset.AssigneeName,
		StatusID:       asset.StatusId,
		StatusName:     asset.StatusShortName,
		AssetTypeID:    asset.AssetTypeId,
		AssetTypeName:  asset.AssetTypeName,
		AssetTypeIcon:  asset.AssetTypeIcon,
		FileStatus:     asset.FileStatus,
		IsResource:     asset.IsResource,
		PreviewID:      asset.PreviewId,
	}
}
