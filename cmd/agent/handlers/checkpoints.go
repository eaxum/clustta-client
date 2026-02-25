package handlers

import (
	"net/http"

	"clustta/internal/repository"

	"github.com/jmoiron/sqlx"
)

// checkpointResponse is the JSON shape returned for each checkpoint.
type checkpointResponse struct {
	ID           string `json:"id"`
	Comment      string `json:"comment"`
	AuthorUID    string `json:"author_id"`
	CreatedAt    string `json:"created_at"`
	FileSize     int    `json:"file_size"`
	IsDownloaded bool   `json:"is_downloaded"`
}

// ListCheckpoints returns checkpoint history for a specific asset.
func ListCheckpoints(w http.ResponseWriter, r *http.Request) {
	activeProjectMu.RLock()
	proj := activeProject
	activeProjectMu.RUnlock()

	if proj == nil {
		jsonError(w, http.StatusBadRequest, "no active project selected")
		return
	}

	assetId := r.PathValue("assetId")
	if assetId == "" {
		jsonError(w, http.StatusBadRequest, "assetId is required")
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

	checkpoints, err := repository.GetCheckpoints(tx, assetId, false)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]checkpointResponse, 0, len(checkpoints))
	for _, c := range checkpoints {
		result = append(result, checkpointResponse{
			ID:           c.Id,
			Comment:      c.Comment,
			AuthorUID:    c.AuthorUID,
			CreatedAt:    c.CreatedAt,
			FileSize:     c.FileSize,
			IsDownloaded: c.IsDownloaded,
		})
	}

	jsonResponse(w, http.StatusOK, result)
}
