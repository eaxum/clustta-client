package services

import "github.com/wailsapp/wails/v3/pkg/application"

const (
	ProjectDataChangedEvent  = "project-data-changed"
	ProjectChangeAssetState  = "asset_state"
	ProjectChangeCheckpoints = "checkpoints"
	ProjectChangeStatus      = "status"
	ProjectChangeSync        = "sync"
)

type ProjectDataChange struct {
	ProjectID  string   `json:"project_id"`
	ProjectURI string   `json:"project_uri"`
	AssetIDs   []string `json:"asset_ids"`
	Changes    []string `json:"changes"`
}

func EmitProjectDataChanged(change ProjectDataChange) {
	app := application.Get()
	if app == nil {
		return
	}
	app.Event.Emit(ProjectDataChangedEvent, change)
}
