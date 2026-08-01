// Package update implements the non-destructive polling-driven merge that
// brings the local project DB up to date with the server without dropping
// tables. It is the experimental counterpart to sync_service.PullData.
//
// UpdateProject is intended to be called from the App.vue polling loop
// behind an experimental settings flag. It is safe to run while the local
// project has unsynced rows: mtime-gated upserts preserve any local row
// whose mtime is newer than the server's, and the destructive PullData
// path remains the only thing that wipes local state.
package update

import (
	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/sync_service"
	"clustta/internal/utils"
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// UpdateProject performs the non-destructive merge against the remote
// project. It opens its own DB connection, runs everything in a single
// transaction, and commits only on success. Chunks are not pulled - this
// is metadata-only.
//
// Short-circuits with a no-op commit when the local sync token already
// matches the server's.
func UpdateProject(ctx context.Context, projectPath, remoteUrl, userId string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	syncToken, err := utils.GetProjectSyncToken(tx)
	if err != nil {
		return err
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	projectInfo, err := repository.GetProjectInfo(remoteUrl, user)
	if err != nil {
		return err
	}

	if err := applyProjectInfo(tx, projectInfo); err != nil {
		return err
	}

	// Short-circuit when local already matches the server token.
	if projectInfo.SyncToken != "" && projectInfo.SyncToken == syncToken {
		return tx.Commit()
	}

	data, err := sync_service.FetchData(remoteUrl, userId)
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Pull any previews referenced by the merged payload. WriteProjectData
	// rejects payloads with missing previews even in non-strict mode.
	missingPreviews, err := sync_service.CalculateMissingPreviews(tx, data)
	if err != nil {
		return err
	}
	if len(missingPreviews) > 0 {
		noop := func(int, int, string, string) {}
		if err := repository.PullPreviews(tx, remoteUrl, missingPreviews, noop); err != nil {
			return err
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Snapshot which payload ids the merge will actually touch (insert or
	// update). This must happen BEFORE WriteProjectData runs, because the
	// merge's update path overwrites local mtime with the wall clock -
	// making post-merge "did the merge accept this row?" impossible to
	// derive from mtime comparison alone.
	targets, err := computeSyncTargets(tx, data)
	if err != nil {
		return err
	}

	// Merge: mtime-gated upsert that preserves local unsynced rows whose
	// mtime is newer than the server's. strict=false because chunks for
	// any newly-arrived checkpoints have not been pulled.
	if err := sync_service.WriteProjectData(tx, data, false); err != nil {
		return err
	}

	// Apply server-side deletions. Local rows with unsynced edits are safe:
	// their ids are absent from data.Tombs.
	if err := repository.AddItemsToTomb(tx, data.Tombs); err != nil {
		return err
	}
	if err := reconcileRestrictedAssetAssignments(tx, data, userId); err != nil {
		return err
	}

	// Mark the rows the merge touched (and the tombs we just applied) as
	// synced. Bypasses the per-table "AFTER UPDATE WHEN OLD.mtime !=
	// NEW.mtime" triggers that would otherwise leave merged rows at
	// synced=0 - and crucially does NOT touch rows that the merge skipped
	// because the local version was newer, so user unsynced edits survive
	// and still get pushed on the next sync.
	if err := markRowsSynced(tx, targets); err != nil {
		return err
	}
	if err := markTombsSynced(tx, data.Tombs); err != nil {
		return err
	}

	if err := utils.SetProjectSyncToken(tx, projectInfo.SyncToken); err != nil {
		return err
	}
	if err := utils.SetLastSyncTime(tx, utils.GetEpochTime()); err != nil {
		return err
	}

	return tx.Commit()
}

func reconcileRestrictedAssetAssignments(tx *sqlx.Tx, data sync_service.ProjectData, userId string) error {
	user, err := repository.GetUser(tx, userId)
	if err != nil {
		return err
	}
	role, err := repository.GetRole(tx, user.RoleId)
	if err != nil {
		return err
	}
	if role.ViewAsset {
		return nil
	}
	return clearRevokedAssetAssignments(tx, data.Assets, userId)
}

func clearRevokedAssetAssignments(tx *sqlx.Tx, visibleAssets []models.Asset, userId string) error {
	// Restricted-user payloads are complete visibility snapshots, so an absent direct assignment is revoked.
	visibleAssetIds := make(map[string]struct{}, len(visibleAssets))
	for _, asset := range visibleAssets {
		visibleAssetIds[asset.Id] = struct{}{}
	}

	assignedAssetIds := []string{}
	if err := tx.Select(&assignedAssetIds, "SELECT id FROM asset WHERE assignee_id = ?", userId); err != nil {
		return err
	}
	for _, assetId := range assignedAssetIds {
		if _, visible := visibleAssetIds[assetId]; visible {
			continue
		}
		if _, err := tx.Exec(
			"UPDATE asset SET assignee_id = '', assigner_id = '' WHERE id = ?",
			assetId,
		); err != nil {
			return fmt.Errorf("reconcile revoked asset assignment: %w", err)
		}
	}
	return nil
}

// applyProjectInfo writes the project-level metadata that does not flow
// through ProjectData. Cheap idempotent writes; the underlying setters
// are responsible for skipping no-op writes when present.
func applyProjectInfo(tx *sqlx.Tx, info repository.ProjectInfo) error {
	if err := utils.SetIsClosed(tx, info.IsClosed); err != nil {
		return err
	}
	if err := utils.SetProjectIcon(tx, info.Icon); err != nil {
		return err
	}
	return utils.SetProjectIgnoreList(tx, info.IgnoreList)
}

// computeSyncTargets returns, per table, the ids in the payload that the
// merge will insert or update. Rows the merge will skip (local mtime >=
// payload mtime) are excluded, so they remain synced=0 and get pushed on
// the next sync.
func computeSyncTargets(tx *sqlx.Tx, data sync_service.ProjectData) (map[string][]string, error) {
	out := map[string][]string{}

	if err := scan(tx, out, "role", data.Roles); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "user", data.Users); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "status", data.Statuses); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "tag", data.Tags); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "asset_type", data.AssetTypes); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "dependency_type", data.DependencyTypes); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "collection_type", data.CollectionTypes); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "collection", data.Collections); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "collection_assignee", data.CollectionAssignees); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "asset", data.Assets); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "asset_checkpoint", data.AssetCheckpoints); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "asset_dependency", data.AssetDependencies); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "collection_dependency", data.CollectionDependencies); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "template", data.Templates); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "workflow", data.Workflows); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "workflow_link", data.WorkflowLinks); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "workflow_collection", data.WorkflowCollections); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "workflow_asset", data.WorkflowAssets); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "asset_tag", data.AssetTags); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "integration_project", data.IntegrationProjects); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "integration_collection_mapping", data.IntegrationCollectionMappings); err != nil {
		return nil, err
	}
	if err := scan(tx, out, "integration_asset_mapping", data.IntegrationAssetMappings); err != nil {
		return nil, err
	}

	return out, nil
}

// scan records into out[table] the subset of `rows` that the merge will
// insert or update for the given table.
func scan[T models.Syncable](tx *sqlx.Tx, out map[string][]string, table string, rows []T) error {
	touched, err := touchedPayloadIds(tx, table, rows)
	if err != nil {
		return fmt.Errorf("compute sync targets for %s: %w", table, err)
	}
	if len(touched) > 0 {
		out[table] = touched
	}
	return nil
}

// touchedPayloadIds returns the subset of rows that will be inserted or
// updated by the merge: ids not present locally, or whose local mtime is
// strictly less than the payload's mtime.
func touchedPayloadIds[T models.Syncable](tx *sqlx.Tx, table string, rows []T) ([]string, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.SyncId()
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf("SELECT id, mtime FROM %s WHERE id IN (%s)", table, placeholders)
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	dbRows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer dbRows.Close()
	localMt := make(map[string]int, len(ids))
	for dbRows.Next() {
		var id string
		var mt int
		if err := dbRows.Scan(&id, &mt); err != nil {
			return nil, err
		}
		localMt[id] = mt
	}
	if err := dbRows.Err(); err != nil {
		return nil, err
	}
	touched := make([]string, 0, len(rows))
	for _, r := range rows {
		local, exists := localMt[r.SyncId()]
		if !exists || local < r.SyncMTime() {
			touched = append(touched, r.SyncId())
		}
	}
	return touched, nil
}

// markRowsSynced flips synced=1 for every id the merge touched. The UPDATE
// intentionally does not change mtime, so the per-table reset triggers
// do not fire.
func markRowsSynced(tx *sqlx.Tx, targets map[string][]string) error {
	for table, ids := range targets {
		if err := markIdsSyncedInTable(tx, table, ids); err != nil {
			return fmt.Errorf("mark synced in %s: %w", table, err)
		}
	}
	return nil
}

// markTombsSynced flips synced=1 for the tomb rows AddItemsToTomb just
// produced (the per-table DELETE triggers insert fresh synced=0 tombs
// that would otherwise get pushed back to the server).
func markTombsSynced(tx *sqlx.Tx, tombs []repository.Tomb) error {
	for _, t := range tombs {
		_, err := tx.Exec(
			"UPDATE tomb SET synced = 1 WHERE id = ? AND table_name = ?",
			t.Id, t.TableName,
		)
		if err != nil {
			return fmt.Errorf("mark tomb synced: %w", err)
		}
	}
	return nil
}

func markIdsSyncedInTable(tx *sqlx.Tx, table string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf("UPDATE %s SET synced = 1 WHERE id IN (%s)", table, placeholders)
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	_, err := tx.Exec(query, args...)
	return err
}
