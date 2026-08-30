package repository

import (
	"clustta/internal/base_service"
	"clustta/internal/error_service"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
)

const checkpointVersionWidth = 4

type Timeline struct {
	CreatedAt string `db:"created_at" json:"created_at"`
	AssetId   string `db:"asset_id" json:"asset_id"`
	AssetPath string `db:"asset_path" json:"asset_path"`
	Extension string `db:"extension" json:"extension"`
	Comment   string `db:"comment" json:"comment"`
	AuthorUID string `db:"author_id" json:"author_id"`
	GroupId   string `db:"group_id" json:"group_id"`
	Preview   []byte `db:"preview" json:"preview"`
}
type CompatTimeline struct {
	CreatedAt     string                      `db:"created_at" json:"created_at"`
	AssetPaths    []string                    `db:"asset_paths" json:"asset_paths"`
	Extensions    []string                    `db:"extensions" json:"extensions"`
	GroupId       string                      `db:"group_id" json:"group_id"`
	Comment       string                      `db:"comment" json:"comment"`
	AuthorUID     string                      `db:"author_id" json:"author_id"`
	Preview       []byte                      `db:"preview" json:"preview"`
	Tags          []models.AssetCheckpointTag `json:"tags"`
	FollowerCount int                         `json:"follower_count"`
}

// nextCheckpointTime keeps checkpoint creation monotonic even when the system
// clock is moved backwards. Checkpoints use second precision, so advancing by
// one second also gives checkpoints created in the same second a stable order.
func nextCheckpointTime(tx *sqlx.Tx, currentTime int64) (int64, error) {
	var latestTime sql.NullInt64
	err := tx.Get(&latestTime, `
		SELECT MAX(
			CASE
				WHEN typeof(created_at) IN ('integer', 'real') THEN CAST(created_at AS INTEGER)
				ELSE unixepoch(created_at)
			END
		)
		FROM asset_checkpoint
		WHERE trashed = 0
	`)
	if err != nil {
		return 0, err
	}

	if latestTime.Valid && latestTime.Int64 >= currentTime {
		return latestTime.Int64 + 1, nil
	}
	return currentTime, nil
}

func CreateNewAssetCheckpoint(
	tx *sqlx.Tx, assetId, message, chunkSequence, checksum string, timeModified int, fileSize int, filePath, author_id, previewId, groupId string,
	callback func(int, int, string, string)) error {
	if groupId == "" {
		return errors.New("group_id can't be empty")
	}

	if checksum == "" {
		filePathParent := filepath.Dir(filePath)
		err := os.MkdirAll(filePathParent, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", filePathParent, err)
			return err
		}
	}
	var checkpointChecksum string
	if chunkSequence == "" {
		if checksum == "" {
			xXHashChecksum, err := utils.GenerateXXHashChecksum(filePath)
			if err != nil {
				fmt.Printf("Error generating XXHash checksum for %s: %v\n", filePath, err)
				return err
			}
			checksum = xXHashChecksum
		}

		Sequence, err := StoreFileChunks(tx, filePath, callback)
		if err != nil {
			fmt.Printf("Error storing file chunks for %s: %v\n", filePath, err)
			return err
		}
		chunkSequence = Sequence
		checkpointChecksum = checksum
		// fileStat, err := os.Stat(filePath)
		// timeModified = int(fileStat.ModTime().Unix())
	} else {
		//check if file exist in the file system
		if checksum == "" {
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				err = RestoreFileFromChunks(tx, chunkSequence, filePath, int64(0), func(i1, i2 int, s1, s2 string) {})
				if err != nil {
					fmt.Printf("Error restoring file %s: %v\n", filePath, err)
					return err
				}
			}
			xXHashChecksum, err := utils.GenerateXXHashChecksum(filePath)
			if err != nil {
				fmt.Printf("Error generating XXHash checksum for %s (from chunks): %v\n", filePath, err)
				return err
			}
			checksum = xXHashChecksum
		}

		checkpointChecksum = checksum
	}

	if timeModified == 0 || fileSize == 0 {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("Error getting file info for %s: %v\n", filePath, err)
			return err
		}
		timeModified = int(fileInfo.ModTime().Unix())
		fileSize = int(fileInfo.Size())
	}
	createdAt, err := nextCheckpointTime(tx, utils.GetEpochTime())
	if err != nil {
		return err
	}

	params := map[string]interface{}{
		"created_at":      createdAt,
		"asset_id":        assetId,
		"xxhash_checksum": checkpointChecksum,
		"time_modified":   timeModified,
		"file_size":       fileSize,
		"comment":         message,
		"chunks":          chunkSequence,
		"author_id":       author_id,
		"preview_id":      previewId,
		"group_id":        groupId,
	}
	err = base_service.Create(tx, "asset_checkpoint", params)
	if err != nil {
		return err
	}
	return nil
}

func CreateCheckpoint(
	tx *sqlx.Tx, assetId, message, chunkSequence, checksum string, timeModified int, fileSize int, filePath, author_id, previewId, groupId string,
	callback func(int, int, string, string)) (models.Checkpoint, error) {
	if groupId == "" {
		return models.Checkpoint{}, errors.New("group_id can't be empty")
	}

	if checksum == "" {
		filePathParent := filepath.Dir(filePath)
		err := os.MkdirAll(filePathParent, os.ModePerm)
		if err != nil {
			return models.Checkpoint{}, err
		}
	}
	var checkpointChecksum string
	lastCheckpoint, err := GetLatestCheckpoint(tx, assetId)
	if err != nil && err.Error() == "no checkpoints" {
		// do nothing
	} else if err != nil {
		return models.Checkpoint{}, err
	}
	if chunkSequence == "" {
		if checksum == "" {
			checksum, err = utils.GenerateXXHashChecksum(filePath)
			if err != nil {
				return models.Checkpoint{}, err
			}
		}

		if lastCheckpoint.Id != "" {
			if checksum == lastCheckpoint.XXHashChecksum {
				return models.Checkpoint{}, errors.New("file not modified")
			}
		}

		Sequence, err := StoreFileChunks(tx, filePath, callback)
		if err != nil {
			return models.Checkpoint{}, err
		}
		chunkSequence = Sequence
		checkpointChecksum = checksum
		// fileStat, err := os.Stat(filePath)
		// timeModified = int(fileStat.ModTime().Unix())
	} else {
		//check if file exist in the file system
		if checksum == "" {
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				err = RestoreFileFromChunks(tx, chunkSequence, filePath, int64(0), func(i1, i2 int, s1, s2 string) {})
				if err != nil {
					return models.Checkpoint{}, err
				}
			}
			checksum, err = utils.GenerateXXHashChecksum(filePath)
			if err != nil {
				return models.Checkpoint{}, err
			}
		}

		if lastCheckpoint.Id != "" {
			if checksum == lastCheckpoint.XXHashChecksum {
				return models.Checkpoint{}, errors.New("file not modified")
			}
		}
		checkpointChecksum = checksum
	}

	if timeModified == 0 || fileSize == 0 {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return models.Checkpoint{}, err
		}
		timeModified = int(fileInfo.ModTime().Unix())
		fileSize = int(fileInfo.Size())
	}
	message, err = resolveCheckpointComment(tx, assetId, message)
	if err != nil {
		return models.Checkpoint{}, err
	}
	createdAt, err := nextCheckpointTime(tx, utils.GetEpochTime())
	if err != nil {
		return models.Checkpoint{}, err
	}

	params := map[string]interface{}{
		"created_at":      createdAt,
		"asset_id":        assetId,
		"xxhash_checksum": checkpointChecksum,
		"time_modified":   timeModified,
		"file_size":       fileSize,
		"comment":         message,
		"chunks":          chunkSequence,
		"author_id":       author_id,
		"preview_id":      previewId,
		"group_id":        groupId,
	}
	err = base_service.Create(tx, "asset_checkpoint", params)
	if err != nil {
		return models.Checkpoint{}, err
	}

	checkpoint := models.Checkpoint{}
	conditions := map[string]interface{}{
		"asset_id":        assetId,
		"xxhash_checksum": checkpointChecksum,
	}
	err = base_service.GetBy(tx, "asset_checkpoint", conditions, &checkpoint)
	if err != nil {
		return models.Checkpoint{}, err
	}
	// hasMissingChunks, err := checkpoint.HasMissingChunks(tx)
	// if err != nil {
	// 	return models.Checkpoint{}, err
	// }
	// if hasMissingChunks {
	// 	return models.Checkpoint{}, error_service.ErrCheckpointExists
	// }
	return checkpoint, nil
}

func resolveCheckpointComment(tx *sqlx.Tx, assetId, message string) (string, error) {
	trimmedMessage := strings.TrimSpace(message)
	if trimmedMessage != "" {
		return trimmedMessage, nil
	}

	checkpointCount := 0
	if err := tx.Get(&checkpointCount, "SELECT COUNT(*) FROM asset_checkpoint WHERE asset_id = ?", assetId); err != nil {
		return "", err
	}

	return fmt.Sprintf("v%0*d", checkpointVersionWidth, checkpointCount+1), nil
}

func AddCheckpoint(
	tx *sqlx.Tx, id string, created_at int64,
	assetId string, xxHashChecksum string, timeModified int, fileSize int, message string,
	chunkSequence string, author_id string, preview_id string, synced bool) error {

	params := map[string]interface{}{
		"id":              id,
		"created_at":      created_at,
		"asset_id":        assetId,
		"xxhash_checksum": xxHashChecksum,
		"time_modified":   timeModified,
		"file_size":       fileSize,
		"comment":         message,
		"chunks":          chunkSequence,
		"author_id":       author_id,
		"preview_id":      preview_id,
		"synced":          synced,
	}
	err := base_service.Create(tx, "asset_checkpoint", params)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return error_service.ErrCheckpointExists
			}
		}
		return err
	}
	return nil
}

func GetLatestCheckpoint(tx *sqlx.Tx, assetId string) (models.Checkpoint, error) {
	checkpoint := models.Checkpoint{}
	query := `SELECT 
		asset_checkpoint.*,
		IFNULL(preview.extension, '') AS preview_extension,
		preview.preview AS preview
	FROM 
		asset_checkpoint
	LEFT JOIN 
		preview ON asset_checkpoint.preview_id = preview.hash
	WHERE asset_checkpoint.asset_id = ? AND trashed = 0 ORDER BY created_at DESC LIMIT 1;`
	err := tx.Get(&checkpoint, query, assetId)
	if err != nil && err == sql.ErrNoRows {
		return checkpoint, errors.New("no checkpoints")
	} else if err != nil {
		return checkpoint, err
	}
	return checkpoint, nil
}

func GetSimpleCheckpoints(tx *sqlx.Tx) ([]models.Checkpoint, error) {
	checkpoints := []models.Checkpoint{}
	query := `SELECT * FROM asset_checkpoint;`
	err := tx.Select(&checkpoints, query)
	if err != nil && err == sql.ErrNoRows {
		return checkpoints, error_service.ErrCheckpointNotFound
	} else if err != nil {
		return checkpoints, err
	}
	return checkpoints, nil
}

func GetCheckpoint(tx *sqlx.Tx, id string) (models.Checkpoint, error) {
	checkpoint := models.Checkpoint{}
	query := `SELECT 
		asset_checkpoint.*,
		IFNULL(preview.extension, '') AS preview_extension,
		preview.preview AS preview
	FROM 
		asset_checkpoint
	LEFT JOIN 
		preview ON asset_checkpoint.preview_id = preview.hash
	WHERE asset_checkpoint.id = ?;`
	err := tx.Get(&checkpoint, query, id)
	if err != nil && err == sql.ErrNoRows {
		return checkpoint, error_service.ErrCheckpointNotFound
	} else if err != nil {
		return checkpoint, err
	}
	checkpoint.HasMissingChunks(tx)
	return checkpoint, nil
}
func SetPublished(tx *sqlx.Tx, id string) error {
	params := map[string]interface{}{
		"synced": true,
	}
	err := base_service.Update(tx, "asset_checkpoint", id, params)
	if err != nil {
		return err
	}
	return nil
}

func GetCheckpoints(tx *sqlx.Tx, assetId string, withDeleted bool) ([]models.Checkpoint, error) {
	checkpoints := []models.Checkpoint{}
	queryWhereClause := "WHERE asset_id = ?"
	if !withDeleted {
		queryWhereClause = "WHERE asset_id = ? AND trashed = 0"
	}
	query := fmt.Sprintf(`SELECT 
		asset_checkpoint.*,
		IFNULL(preview.extension, '') AS preview_extension,
		preview.preview AS preview
	FROM 
		asset_checkpoint
	LEFT JOIN 
		preview ON asset_checkpoint.preview_id = preview.hash 
	%s ORDER BY created_at DESC;`, queryWhereClause)
	err := tx.Select(&checkpoints, query, assetId)
	if err != nil && err == sql.ErrNoRows {
		return checkpoints, errors.New("no checkpoints")
	} else if err != nil {
		return checkpoints, err
	}
	for i, _ := range checkpoints {
		checkpoints[i].HasMissingChunks(tx)
	}
	return checkpoints, nil
}

func GetTimeline(tx *sqlx.Tx) ([]CompatTimeline, error) {
	timeline := []CompatTimeline{}
	checkpoints := []Timeline{}
	query := `SELECT 
		asset_checkpoint.created_at,
		asset_checkpoint.comment,
		asset_checkpoint.asset_id,
		asset_checkpoint.author_id,
		asset_checkpoint.group_id,
		preview.preview AS preview,
		IFNULL(full_asset.asset_path, '') AS asset_path,
		IFNULL(full_asset.extension, '') AS extension
	FROM 
		asset_checkpoint
	LEFT JOIN 
		preview ON asset_checkpoint.preview_id = preview.hash
	LEFT JOIN 
		full_asset ON asset_checkpoint.asset_id = full_asset.id
	WHERE asset_checkpoint.trashed = 0
	ORDER BY asset_checkpoint.created_at DESC;`
	err := tx.Select(&checkpoints, query)
	if err != nil && err == sql.ErrNoRows {
		return timeline, errors.New("no checkpoints")
	} else if err != nil {
		return timeline, err
	}

	previousCheckpoint := CompatTimeline{}
	for i, checkpoint := range checkpoints {
		if previousCheckpoint.GroupId == "" {
			previousCheckpoint = CompatTimeline{
				CreatedAt:  checkpoint.CreatedAt,
				AssetPaths: []string{checkpoint.AssetPath},
				Extensions: []string{checkpoint.Extension},
				GroupId:    checkpoint.GroupId,
				Comment:    checkpoint.Comment,
				AuthorUID:  checkpoint.AuthorUID,
				Preview:    checkpoint.Preview,
			}
			if i == len(checkpoints)-1 {
				timeline = append(timeline, previousCheckpoint)
			}
			continue
		}

		if previousCheckpoint.GroupId == checkpoint.GroupId {
			previousCheckpoint.AssetPaths = append(previousCheckpoint.AssetPaths, checkpoint.AssetPath)
			previousCheckpoint.Extensions = append(previousCheckpoint.Extensions, checkpoint.Extension)
		} else {
			timeline = append(timeline, previousCheckpoint)
			previousCheckpoint = CompatTimeline{
				CreatedAt:  checkpoint.CreatedAt,
				AssetPaths: []string{checkpoint.AssetPath},
				Extensions: []string{checkpoint.Extension},
				GroupId:    checkpoint.GroupId,
				Comment:    checkpoint.Comment,
				AuthorUID:  checkpoint.AuthorUID,
				Preview:    checkpoint.Preview,
			}
		}
		if i == len(checkpoints)-1 {
			timeline = append(timeline, previousCheckpoint)
		}
	}

	for i := range timeline {
		if err = tx.Select(&timeline[i].Tags, `
			SELECT act.*, t.name
			FROM asset_checkpoint_tag act
			JOIN tag t ON t.id = act.tag_id
			JOIN asset_checkpoint ac ON ac.id = act.checkpoint_id
			WHERE ac.group_id = ?
			GROUP BY act.tag_id
			ORDER BY t.name COLLATE NOCASE
		`, timeline[i].GroupId); err != nil {
			return timeline, err
		}
		if err = tx.Get(&timeline[i].FollowerCount, `
			SELECT COUNT(*)
			FROM asset_dependency ad
			JOIN asset_checkpoint_tag act ON act.id = ad.asset_checkpoint_tag_id
			JOIN asset_checkpoint ac ON ac.id = act.checkpoint_id
			WHERE ac.group_id = ?
		`, timeline[i].GroupId); err != nil {
			return timeline, err
		}
	}

	return timeline, nil
}

func GetLatestCheckpointsBtwTime(tx *sqlx.Tx, higherTime, lowerTime int) ([]models.Checkpoint, error) {
	checkpoints := []models.Checkpoint{}
	query := `WITH latest_checkpoints AS (
		SELECT 
		  tc.*,
		  ROW_NUMBER() OVER (PARTITION BY tc.asset_id ORDER BY tc.created_at DESC) AS rn
		FROM 
		  asset_checkpoint tc
		WHERE 
		  tc.created_at >= ? -- Replace with your lower Unix timestamp
		  AND tc.created_at <= ? -- Replace with your higher Unix timestamp
		  AND tc.trashed = 0
	  )
	  SELECT * FROM latest_checkpoints WHERE rn = 1 ORDER BY latest_checkpoints.created_at DESC;`
	err := tx.Select(&checkpoints, query, lowerTime, higherTime)
	if err != nil && err == sql.ErrNoRows {
		return checkpoints, errors.New("no checkpoints")
	} else if err != nil {
		return checkpoints, err
	}
	return checkpoints, nil
}

func GetLatestCheckpointsByTime(tx *sqlx.Tx, checkpointTime int64) ([]models.Checkpoint, error) {
	checkpoints := []models.Checkpoint{}
	query := `WITH latest_checkpoints AS (
		SELECT 
		  tc.*,
		  ROW_NUMBER() OVER (PARTITION BY tc.asset_id ORDER BY tc.created_at DESC) AS rn
		FROM 
		  asset_checkpoint tc
		WHERE tc.created_at <= ? -- Replace with your higher Unix timestamp
		  AND tc.trashed = 0
	  )
	  SELECT id, created_at, mtime, asset_id, xxhash_checksum, time_modified, file_size, chunks, comment, author_id, preview_id, trashed, synced
	  FROM latest_checkpoints WHERE rn = 1 ORDER BY latest_checkpoints.created_at DESC;`
	err := tx.Select(&checkpoints, query, checkpointTime)
	if err != nil && err == sql.ErrNoRows {
		return checkpoints, errors.New("no checkpoints")
	} else if err != nil {
		return checkpoints, err
	}
	return checkpoints, nil
}

func GetFirstAndLastCheckpointTime(tx *sqlx.Tx) (firstTime, lastTime int, err error) {
	// Create a struct to hold the result
	type TimeResult struct {
		FirstTime int `db:"first_time"`
		LastTime  int `db:"last_time"`
	}

	result := TimeResult{}

	// Query to get both times in a single query
	query := `
    SELECT 
        MIN(created_at) AS first_time,
        MAX(created_at) AS last_time
    FROM 
        asset_checkpoint
    WHERE 
        trashed = 0
    `

	err = tx.Get(&result, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, errors.New("no checkpoints found")
		}
		return 0, 0, err
	}

	return result.FirstTime, result.LastTime, nil
}

func GetDeletedCheckpoints(tx *sqlx.Tx) ([]models.Checkpoint, error) {
	checkpoints := []models.Checkpoint{}
	query := `SELECT 
		asset_checkpoint.*,
		IFNULL(preview.extension, '') AS preview_extension,
		preview.preview AS preview
	FROM 
		asset_checkpoint
	LEFT JOIN 
		preview ON asset_checkpoint.preview_id = preview.hash 
	WHERE asset_checkpoint.trashed = 1;`
	err := tx.Select(&checkpoints, query)
	if err != nil {
		return checkpoints, err
	}
	return checkpoints, nil
}

func RevertToCheckpoint(tx *sqlx.Tx, checkpointId string, filePath string, callback func(int, int, string, string)) error {
	if filePath == "" {
		return errors.New("filepath cant be empty")
	}
	checkpoint, err := GetCheckpoint(tx, checkpointId)
	if err != nil {
		return err
	}
	if checkpoint.Id == "" {
		return errors.New("checkpoint not found")
	}
	filePathParent := filepath.Dir(filePath)
	os.MkdirAll(filePathParent, os.ModePerm)
	err = RestoreFileFromChunks(tx, checkpoint.Chunks, filePath, int64(checkpoint.TimeModified), callback)
	if err != nil {
		return err
	}
	return nil
}

func RevertToLatestCheckpoint(tx *sqlx.Tx, assetId string, filePath string, callback func(int, int, string, string)) error {
	latestCheckpoint, err := GetLatestCheckpoint(tx, assetId)
	if err != nil {
		return err
	}
	err = RevertToCheckpoint(tx, latestCheckpoint.Id, filePath, callback)
	if err != nil {
		return err
	}
	return nil
}

func DeleteCheckpoint(tx *sqlx.Tx, checkpointId string, checkIfLast bool, recycle bool) error {
	checkpoint, err := GetCheckpoint(tx, checkpointId)
	if err != nil {
		return err
	}
	assetId := checkpoint.AssetId
	if checkIfLast {
		checkpoints, err := GetCheckpoints(tx, assetId, false)
		if err != nil {
			return err
		}
		if len(checkpoints) == 1 {
			return errors.New("cannot delete the only checkpoint")
		}
	}

	if checkpoint.Id == "" {
		return errors.New("checkpoint not found")
	}
	if recycle {
		err = base_service.MarkAsDeleted(tx, "asset_checkpoint", checkpointId)
	} else {
		err = base_service.Delete(tx, "asset_checkpoint", checkpointId)
	}
	if err != nil {
		return err
	}
	return nil
}

func DeleteCheckpoints(tx *sqlx.Tx, assetId string) error {
	checkpoints, err := GetCheckpoints(tx, assetId, true)
	if err != nil {
		return err
	}
	for _, checkpoint := range checkpoints {
		err := DeleteCheckpoint(tx, checkpoint.Id, false, false)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddMissingGroupIds populates missing group_id values for existing checkpoints
// Groups checkpoints by author, time windows, and similar comments
func AddMissingGroupIds(tx *sqlx.Tx) (int, int, error) {
	// Get all checkpoints with empty group_id
	type CheckpointInfo struct {
		Id        string `db:"id"`
		CreatedAt string `db:"created_at"`
		AuthorId  string `db:"author_id"`
		Comment   string `db:"comment"`
	}

	query := `
		SELECT id, created_at, author_id, comment
		FROM asset_checkpoint 
		WHERE (group_id = '' OR group_id IS NULL) AND trashed = 0
		ORDER BY author_id, created_at ASC
	`

	checkpoints := []CheckpointInfo{}
	err := tx.Select(&checkpoints, query)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch checkpoints: %w", err)
	}

	if len(checkpoints) == 0 {
		return 0, 0, nil
	}

	// Group checkpoints by time windows and author
	var currentGroupId string
	var currentAuthor string
	var currentComment string
	var currentGroupStartTime int64
	groupCount := 0
	checkpointCount := 0
	timeWindowSeconds := int64(300) // 5 minutes

	for _, checkpoint := range checkpoints {
		// Convert timestamp to epoch (simplified - just use ordering for grouping)
		checkpointEpoch := parseTimestampForGrouping(checkpoint.CreatedAt)

		// Check if we need to start a new group
		shouldStartNewGroup := false

		if currentGroupId == "" {
			// First checkpoint
			shouldStartNewGroup = true
		} else if currentAuthor != checkpoint.AuthorId {
			// Different author
			shouldStartNewGroup = true
		} else if checkpointEpoch-currentGroupStartTime > timeWindowSeconds {
			// Time window exceeded
			shouldStartNewGroup = true
		} else if currentComment != checkpoint.Comment && checkpoint.Comment != "" && currentComment != "" {
			// Different comment (only if both are non-empty)
			shouldStartNewGroup = true
		}

		if shouldStartNewGroup {
			currentGroupId = uuid.New().String()
			currentAuthor = checkpoint.AuthorId
			currentComment = checkpoint.Comment
			currentGroupStartTime = checkpointEpoch
			groupCount++
		}

		// Update the checkpoint with the current group_id and mark as not synced
		updateQuery := `UPDATE asset_checkpoint SET group_id = ?, synced = 0 WHERE id = ?`
		_, err = tx.Exec(updateQuery, currentGroupId, checkpoint.Id)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to update group_id for checkpoint %s: %w", checkpoint.Id, err)
		}

		checkpointCount++
	}

	return checkpointCount, groupCount, nil
}

// parseTimestampForGrouping converts timestamp string to int64 for grouping logic
func parseTimestampForGrouping(timestamp string) int64 {
	// For simplicity, just return 0 and rely on ordering for grouping
	// The actual time window logic can be enhanced later if needed
	return 0
}
