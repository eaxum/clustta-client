package agent

import (
	"clustta/internal/utils"
	"sort"
	"strings"
)

const entityReferenceRootID = ""

type EntityReference struct {
	ID        string `db:"id" json:"id"`
	Type      string `db:"type" json:"type"`
	Name      string `db:"name" json:"name"`
	Extension string `db:"extension" json:"extension,omitempty"`
	ParentID  string `db:"parent_id" json:"parent_id,omitempty"`
	Path      string `db:"path" json:"path"`
}

func ListEntityReferenceChildren(projectPath, parentID string) ([]EntityReference, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	return listEntityReferenceChildren(tx, parentID)
}

func listEntityReferenceChildren(queryer interface {
	Select(interface{}, string, ...interface{}) error
}, parentID string) ([]EntityReference, error) {
	references := []EntityReference{}
	query := `
		SELECT id, 'collection' AS type, name, '' AS extension,
			parent_id, COALESCE(collection_path, '') AS path
		FROM collection
		WHERE parent_id = ? AND trashed = 0
		UNION ALL
		SELECT a.id, 'asset' AS type, a.name, a.extension,
			a.collection_id AS parent_id,
			COALESCE(c.collection_path, '/') || a.name || a.extension AS path
		FROM asset a
		LEFT JOIN collection c ON c.id = a.collection_id
		WHERE a.collection_id = ? AND a.trashed = 0
	`
	if err := queryer.Select(&references, query, parentID, parentID); err != nil {
		return nil, err
	}

	for index := range references {
		references[index].Path = normalizeReferencePath(references[index].Path)
	}
	sort.SliceStable(references, func(i, j int) bool {
		if references[i].Type != references[j].Type {
			return references[i].Type == "collection"
		}
		return strings.ToLower(references[i].Name) < strings.ToLower(references[j].Name)
	})
	return references, nil
}

func normalizeReferencePath(path string) string {
	normalized := strings.ReplaceAll(path, "\\", "/")
	normalized = strings.Trim(normalized, "/")
	return normalized
}
