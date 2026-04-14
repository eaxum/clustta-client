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
	ID           string `json:"id"`
	Name         string `json:"name"`
	Extension    string `json:"extension"`
	CollectionName   string `json:"collection_name"`
	CollectionPath   string `json:"collection_path"`
	AssetPath     string `json:"asset_path"`
	FilePath     string `json:"file_path"`
	AssigneeId   string `json:"assignee_id"`
	AssigneeName string `json:"assignee_name"`
	StatusName   string `json:"status_short_name"`
	AssetTypeName string `json:"asset_type_name"`
	AssetTypeIcon string `json:"asset_type_icon"`
	FileStatus   string `json:"file_status"`
	IsResource   bool   `json:"is_resource"`
	PreviewId    string `json:"preview_id"`
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

		result = append(result, assetResponse{
			ID:           t.Id,
			Name:         t.Name,
			Extension:    t.Extension,
			CollectionName:   t.CollectionName,
			CollectionPath:   t.CollectionPath,
			AssetPath:     t.AssetPath,
			FilePath:     t.GetFilePath(),
			AssigneeId:   t.AssigneeId,
			AssigneeName: t.AssigneeName,
			StatusName:   t.StatusShortName,
			AssetTypeName: t.AssetTypeName,
			AssetTypeIcon: t.AssetTypeIcon,
			FileStatus:   t.FileStatus,
			IsResource:   t.IsResource,
			PreviewId:    t.PreviewId,
		})
	}

	jsonResponse(w, http.StatusOK, result)
}
