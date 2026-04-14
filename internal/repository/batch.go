package repository

import (
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func BatchCreateCollections(tx *sqlx.Tx, collections []models.Collection) error {
	if len(collections) == 0 {
		return nil
	}

	// Build bulk insert query
	valueStrings := make([]string, 0, len(collections))
	valueArgs := make([]interface{}, 0, len(collections)*8)

	for _, collection := range collections {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			collection.Id,
			collection.Name,
			collection.Description,
			collection.CollectionTypeId,
			collection.ParentId,
			collection.PreviewId,
			collection.IsLibrary,
			utils.GetEpochTime(), // mtime
		)
	}

	stmt := fmt.Sprintf(`
		INSERT INTO collection (id, name, description, collection_type_id, parent_id, preview_id, is_library, mtime)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := tx.Exec(stmt, valueArgs...)
	return err
}

// BatchUpdateCollections updates multiple collections in a single transaction
func BatchUpdateCollections(tx *sqlx.Tx, collections []models.Collection) error {
	if len(collections) == 0 {
		return nil
	}

	// Use CASE statements for bulk updates
	ids := make([]string, len(collections))
	nameMap := make(map[string]string)
	parentMap := make(map[string]string)
	previewMap := make(map[string]string)
	libraryMap := make(map[string]bool)

	for i, collection := range collections {
		ids[i] = collection.Id
		nameMap[collection.Id] = collection.Name
		parentMap[collection.Id] = collection.ParentId
		previewMap[collection.Id] = collection.PreviewId
		libraryMap[collection.Id] = collection.IsLibrary
	}

	// Build CASE statements
	nameCases := make([]string, 0, len(collections))
	parentCases := make([]string, 0, len(collections))
	previewCases := make([]string, 0, len(collections))
	libraryCases := make([]string, 0, len(collections))

	for _, collection := range collections {
		nameCases = append(nameCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", collection.Id, collection.Name))
		parentCases = append(parentCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", collection.Id, collection.ParentId))
		previewCases = append(previewCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", collection.Id, collection.PreviewId))
		libraryCases = append(libraryCases, fmt.Sprintf("WHEN id = '%s' THEN %t", collection.Id, collection.IsLibrary))
	}

	query := fmt.Sprintf(`
		UPDATE collection SET 
			name = CASE %s END,
			parent_id = CASE %s END,
			preview_id = CASE %s END,
			is_library = CASE %s END,
			mtime = %d
		WHERE id IN ('%s')
	`,
		strings.Join(nameCases, " "),
		strings.Join(parentCases, " "),
		strings.Join(previewCases, " "),
		strings.Join(libraryCases, " "),
		utils.GetEpochTime(),
		strings.Join(ids, "','"),
	)

	_, err := tx.Exec(query)
	return err
}

// BatchCreateAssets creates multiple assets in a single transaction
func BatchCreateAssets(tx *sqlx.Tx, assets []models.Asset) error {
	if len(assets) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(assets))
	valueArgs := make([]interface{}, 0, len(assets)*12)

	for _, asset := range assets {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			asset.Id,
			asset.CreatedAt,
			asset.Name,
			asset.AssetTypeId,
			asset.CollectionId,
			asset.StatusId,
			asset.Extension,
			asset.Description,
			asset.Pointer,
			asset.IsLink,
			asset.AssigneeId,
			asset.PreviewId,
		)
	}

	stmt := fmt.Sprintf(`
		INSERT INTO assets (id, created_at, name, asset_type_id, collection_id, status_id, extension, description, pointer, is_link, assignee_id, preview_id)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := tx.Exec(stmt, valueArgs...)
	return err
}

// BatchUpdateAssets updates multiple assets in a single transaction
func BatchUpdateAssets(tx *sqlx.Tx, assets []models.Asset) error {
	if len(assets) == 0 {
		return nil
	}

	ids := make([]string, len(assets))
	for i, asset := range assets {
		ids[i] = asset.Id
	}

	// Build CASE statements for each field that needs updating
	nameCases := make([]string, 0, len(assets))
	collectionCases := make([]string, 0, len(assets))
	typeCases := make([]string, 0, len(assets))
	assigneeCases := make([]string, 0, len(assets))
	statusCases := make([]string, 0, len(assets))
	previewCases := make([]string, 0, len(assets))
	resourceCases := make([]string, 0, len(assets))
	linkCases := make([]string, 0, len(assets))
	pointerCases := make([]string, 0, len(assets))

	for _, asset := range assets {
		nameCases = append(nameCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", asset.Id, asset.Name))
		collectionCases = append(collectionCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", asset.Id, asset.CollectionId))
		typeCases = append(typeCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", asset.Id, asset.AssetTypeId))
		assigneeCases = append(assigneeCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", asset.Id, asset.AssigneeId))
		statusCases = append(statusCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", asset.Id, asset.StatusId))
		previewCases = append(previewCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", asset.Id, asset.PreviewId))
		resourceCases = append(resourceCases, fmt.Sprintf("WHEN id = '%s' THEN %t", asset.Id, asset.IsResource))
		linkCases = append(linkCases, fmt.Sprintf("WHEN id = '%s' THEN %t", asset.Id, asset.IsLink))
		pointerCases = append(pointerCases, fmt.Sprintf("WHEN id = '%s' THEN '%s'", asset.Id, asset.Pointer))
	}

	query := fmt.Sprintf(`
		UPDATE assets SET 
			name = CASE %s END,
			collection_id = CASE %s END,
			asset_type_id = CASE %s END,
			assignee_id = CASE %s END,
			status_id = CASE %s END,
			preview_id = CASE %s END,
			is_resource = CASE %s END,
			is_link = CASE %s END,
			pointer = CASE %s END,
			mtime = %d
		WHERE id IN ('%s')
	`,
		strings.Join(nameCases, " "),
		strings.Join(collectionCases, " "),
		strings.Join(typeCases, " "),
		strings.Join(assigneeCases, " "),
		strings.Join(statusCases, " "),
		strings.Join(previewCases, " "),
		strings.Join(resourceCases, " "),
		strings.Join(linkCases, " "),
		strings.Join(pointerCases, " "),
		utils.GetEpochTime(),
		strings.Join(ids, "','"),
	)

	_, err := tx.Exec(query)
	return err
}

// BatchCreateCheckpoints creates multiple asset checkpoints in a single transaction
func BatchCreateCheckpoints(tx *sqlx.Tx, checkpoints []models.Checkpoint) error {
	if len(checkpoints) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(checkpoints))
	valueArgs := make([]interface{}, 0, len(checkpoints)*10)

	for _, checkpoint := range checkpoints {
		epochTime, err := utils.RFC3339ToEpoch(checkpoint.CreatedAt)
		if err != nil {
			return err
		}

		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			checkpoint.Id,
			epochTime,
			checkpoint.AssetId,
			checkpoint.XXHashChecksum,
			checkpoint.TimeModified,
			checkpoint.FileSize,
			checkpoint.Comment,
			checkpoint.Chunks,
			checkpoint.AuthorUID,
			checkpoint.PreviewId,
		)
	}

	stmt := fmt.Sprintf(`
		INSERT INTO asset_checkpoints (id, created_at, asset_id, xxhash_checksum, time_modified, file_size, comment, chunks, author_uid, preview_id)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := tx.Exec(stmt, valueArgs...)
	return err
}

// GetAllCollections retrieves all collections for batch processing
func GetAllCollections(tx *sqlx.Tx) ([]models.Collection, error) {
	var collections []models.Collection
	query := `
		SELECT id, name, description, collection_type_id, parent_id, preview_id, is_library, mtime
		FROM collection
	`
	err := tx.Select(&collections, query)
	return collections, err
}

// Alternative approach using prepared statements for very large datasets
func BatchCreateCollectionsWithPreparedStmt(tx *sqlx.Tx, collections []models.Collection) error {
	if len(collections) == 0 {
		return nil
	}

	stmt, err := tx.Preparex(`
		INSERT INTO collection (id, name, description, collection_type_id, parent_id, preview_id, is_library, mtime)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, collection := range collections {
		_, err := stmt.Exec(
			collection.Id,
			collection.Name,
			collection.Description,
			collection.CollectionTypeId,
			collection.ParentId,
			collection.PreviewId,
			collection.IsLibrary,
			utils.GetEpochTime(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
