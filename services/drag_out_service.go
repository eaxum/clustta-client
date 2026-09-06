package services

import (
	"context"
	"errors"
	"sync/atomic"

	"clustta/internal/auth_service"
	bridgeassets "clustta/internal/bridge/assets"
	"github.com/eaxum/wails-dragout"
	wailsdrag "github.com/eaxum/wails-dragout/wails"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type DragOutService struct {
	active atomic.Bool
}

type DragOutRequest struct {
	ProjectPath string   `json:"project_path"`
	ProjectID   string   `json:"project_id"`
	AssetIDs    []string `json:"asset_ids"`
}

func (s *DragOutService) Available() bool {
	return dragout.Available
}

// StartDrag keeps the binding pending until the native session finishes.
func (s *DragOutService) StartDrag(ctx context.Context, request DragOutRequest) (string, error) {
	if !dragout.Available {
		return "", dragout.ErrUnsupported
	}
	window, ok := ctx.Value(application.WindowKey).(*application.WebviewWindow)
	if !ok || window == nil {
		return "", errors.New("dragging requires a desktop window")
	}
	if !s.active.CompareAndSwap(false, true) {
		return "", errors.New("a file drag is already in progress")
	}
	defer s.active.Store(false)
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return "", err
	}
	paths, err := bridgeassets.ResolveLocalFiles(ctx, request.ProjectPath, request.ProjectID, user.Id, request.AssetIDs)
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	result, err := wailsdrag.Start(ctx, window, paths)
	return string(result), err
}
