package services

import (
	"clustta/internal/activity"
	"clustta/internal/auth_service"
	"clustta/internal/chunk_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/sync_service"
	"clustta/internal/settings"
	"clustta/internal/transfer"
	"clustta/internal/utils"
	"clustta/output"
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"math"
	"os"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var activities = activity.NewManager(func(snapshot activity.Snapshot) {
	if app := application.Get(); app != nil {
		app.Event.Emit("operations-updated", snapshot)
	}
})

type ActivityService struct{}

func (s *ActivityService) List() activity.Snapshot         { return activities.List() }
func (s *ActivityService) Cancel(operationID string) error { return activities.Cancel(operationID) }
func (s *ActivityService) Dismiss(operationID string)      { activities.Dismiss(operationID) }
func (s *ActivityService) ClearFinished()                  { activities.Dismiss("") }

type transferAction struct {
	id      string
	kind    string
	ctx     context.Context
	path    string
	user    auth_service.User
	release func()
}

func beginTransfer(parent context.Context, projectPath, kind, title string) (*transferAction, error) {
	if _, err := os.Stat(projectPath); err != nil {
		return nil, err
	}
	path, err := transfer.CanonicalProject(projectPath)
	if err != nil {
		return nil, err
	}
	release, err := transfer.Enter(path)
	if err != nil {
		return nil, err
	}
	user, err := auth_service.GetActiveUser()
	if err != nil {
		release()
		return nil, err
	}
	db, err := utils.OpenDb(path)
	if err != nil {
		release()
		return nil, err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		release()
		return nil, err
	}
	defer tx.Rollback()
	projectID, err := utils.GetProjectId(tx)
	if err != nil {
		release()
		return nil, err
	}
	name, err := utils.GetProjectName(tx)
	if err != nil {
		release()
		return nil, err
	}
	workingDir, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		release()
		return nil, err
	}
	releaseDestination, err := transfer.ReserveDestination(path, workingDir)
	if err != nil {
		release()
		return nil, err
	}
	releaseProject := release
	release = func() { releaseDestination(); releaseProject() }
	id, ctx := activities.Start(parent, activity.Operation{ProjectID: projectID, ProjectURI: projectPath, ProjectName: name, Kind: kind, Title: title})
	return &transferAction{id: id, kind: kind, ctx: chunk_service.DownloadContext(ctx), path: path, user: user, release: release}, nil
}

func (a *transferAction) finish(err error) {
	metadataOnly, settingsErr := settings.GetMetadataOnlyStorage()
	if settingsErr != nil {
		log.Printf("Reading chunk cleanup preference: %s", settingsErr)
	}
	if metadataOnly && err == nil && a.kind == "fetch" {
		transfer.CleanupWhenIdle(a.path, func() {
			db, cleanupErr := utils.OpenDb(a.path)
			if cleanupErr == nil {
				cleanupErr = reclaimChunkCache(a.path, db, func() {})
				db.Close()
			}
			if cleanupErr != nil {
				log.Printf("Deferred transfer cleanup failed: %s", cleanupErr)
			}
		})
	}
	a.release()
	activities.Finish(a.id, err)
}

func (a *transferAction) prepareAndWait(assetIDs []string, collectionIDs, checkpointID string) error {
	if err := a.ctx.Err(); err != nil {
		return err
	}
	db, err := utils.OpenDb(a.path)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if checkpointID != "" {
		checkpoint, err := repository.GetCheckpoint(tx, checkpointID)
		if err != nil {
			return err
		}
		assetIDs = []string{checkpoint.AssetId}
	}
	assets := []models.Asset{}
	for _, id := range assetIDs {
		asset, err := repository.GetAsset(tx, id)
		if err != nil {
			return err
		}
		assets = append(assets, asset)
	}
	a.assets(assets)
	if collectionIDs != "" {
		names := []string{}
		for _, id := range strings.Split(collectionIDs, ",") {
			collection, err := repository.GetCollection(tx, strings.TrimSpace(id))
			if err != nil {
				return err
			}
			a.collection(collection)
			names = append(names, collection.Name)
		}
		a.title("Fetching " + strings.Join(names, ", "))
	}
	if err := tx.Rollback(); err != nil {
		return err
	}
	if err := db.Close(); err != nil {
		return err
	}
	return activities.Wait(a.ctx, a.id)
}

func (a *transferAction) report(progress output.ProgressReport) {
	if math.IsNaN(progress.Percentage) || math.IsInf(progress.Percentage, 0) {
		progress.Percentage = 0
	}
	activities.Update(a.id, func(operation *activity.Operation) {
		operation.Message = progress.Message
		operation.Current = progress.Current
		operation.Total = progress.Total
		operation.Percentage = min(max(progress.Percentage, 0), 99)
	})
}

func (a *transferAction) rebuilding(index, count, current, total int, name string) {
	fraction := 0.0
	if total > 0 {
		fraction = min(max(float64(current)/float64(total), 0), 1)
	}
	percentage := 0.0
	if count > 0 {
		percentage = (float64(index) + fraction) / float64(count) * 100
	}
	a.report(output.ProgressReport{Message: name, Current: index + 1, Total: count, Percentage: percentage})
}

func (a *transferAction) downloaded(current, total int, message, extra string) {
	percentage := 0.0
	if total > 0 {
		percentage = float64(current) / float64(total) * 99
	}
	a.report(output.ProgressReport{Message: message, Current: current, Total: total, Percentage: percentage})
	activities.Update(a.id, func(operation *activity.Operation) {
		operation.DownloadMessage = message
		operation.ExtraMessage = extra
	})
}

func (a *transferAction) assets(assets []models.Asset) {
	details := make([]activity.Asset, 0, len(assets))
	seen := map[string]bool{}
	for _, asset := range assets {
		if seen[asset.Id] {
			continue
		}
		seen[asset.Id] = true
		details = append(details, activity.Asset{ID: asset.Id, Name: asset.Name, Extension: asset.Extension, CollectionID: asset.CollectionId, CollectionPath: asset.CollectionPath})
	}
	activities.Update(a.id, func(operation *activity.Operation) {
		operation.Assets = details
		if len(operation.Collections) > 0 {
			return
		}
		if operation.Title != "Fetching assets" && operation.Title != "Downloading checkpoint" {
			return
		}
		verb := "Fetching"
		if a.kind == "checkpoint" {
			verb = "Downloading"
		}
		if len(details) == 1 {
			operation.Title = verb + " " + details[0].Name
		} else if len(details) > 1 {
			operation.Title = fmt.Sprintf("%s %d assets", verb, len(details))
		}
	})
}

func (a *transferAction) collection(collection models.Collection) {
	activities.Update(a.id, func(operation *activity.Operation) {
		for _, existing := range operation.Collections {
			if existing.ID == collection.Id {
				return
			}
		}
		operation.Collections = append(operation.Collections, activity.Collection{ID: collection.Id, Name: collection.Name, Path: collection.CollectionPath})
	})
}

func (a *transferAction) restored(ids []string) {
	activities.Update(a.id, func(operation *activity.Operation) { operation.RestoredAssetIDs = append([]string{}, ids...) })
}

func (a *transferAction) title(title string) {
	activities.Update(a.id, func(operation *activity.Operation) { operation.Title = title })
}

func (a *transferAction) authorize(tx *sqlx.Tx, ids []string, download bool) error {
	user, role, err := activeAssetRole(tx)
	if err != nil {
		return err
	}
	if user.Id != a.user.Id {
		return errors.New("account changed during transfer; retry using the original account")
	}
	if err := authorizeCheckpointRestoreTx(tx, user, role, ids); err != nil {
		return err
	}
	if download && !role.PullChunk {
		return errors.New("user does not have pull_chunk permission")
	}
	return nil
}

func (a *transferAction) downloadCheckpoint(remoteURL, checkpointID string) error {
	db, err := utils.OpenDb(a.path)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	checkpoint, err := repository.GetCheckpoint(tx, checkpointID)
	if err != nil {
		return err
	}
	asset, err := repository.GetAsset(tx, checkpoint.AssetId)
	if err != nil {
		return err
	}
	if err := a.authorize(tx, []string{asset.Id}, true); err != nil {
		return err
	}
	a.assets([]models.Asset{asset})
	if err = tx.Rollback(); err != nil {
		return err
	}
	return sync_service.DownloadCheckpoint(a.ctx, a.path, remoteURL, checkpointID, a.user.Id, a.downloaded)
}

type transferFileState struct {
	path string
	info os.FileInfo
}

func captureTransferFile(path string) (transferFileState, error) {
	info, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return transferFileState{}, err
	}
	return transferFileState{path: path, info: info}, nil
}

func (state transferFileState) verify(path string) error {
	current, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	unchanged := state.path == path && state.info == nil && current == nil
	if state.info != nil && current != nil {
		unchanged = state.path == path && os.SameFile(state.info, current) && state.info.Size() == current.Size() && state.info.ModTime().Equal(current.ModTime())
	}
	if !unchanged {
		return fmt.Errorf("file changed while downloading: %s; review local changes before retrying", path)
	}
	return nil
}
