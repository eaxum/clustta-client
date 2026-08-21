package agent

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestListEntityReferenceChildrenReturnsMinimalSortedChildren(t *testing.T) {
	db, err := sqlx.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE collection (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, parent_id TEXT NOT NULL,
			collection_path TEXT NOT NULL, trashed BOOLEAN NOT NULL
		);
		CREATE TABLE asset (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, extension TEXT NOT NULL,
			collection_id TEXT NOT NULL, trashed BOOLEAN NOT NULL
		);
		INSERT INTO collection VALUES
			('characters', 'Characters', '', '/Characters/', 0),
			('props', 'Props', '', '/Props/', 0),
			('hidden', 'Hidden', '', '/Hidden/', 1),
			('heroes', 'Heroes', 'characters', '/Characters/Heroes/', 0);
		INSERT INTO asset VALUES
			('root-file', 'Readme', '.txt', '', 0),
			('hero', 'Zeus', '.blend', 'characters', 0),
			('old-hero', 'Old', '.blend', 'characters', 1);
	`)
	require.NoError(t, err)
	tx, err := db.Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	root, err := listEntityReferenceChildren(tx, entityReferenceRootID)
	require.NoError(t, err)
	require.Equal(t, []EntityReference{
		{ID: "characters", Type: "collection", Name: "Characters", Path: "Characters"},
		{ID: "props", Type: "collection", Name: "Props", Path: "Props"},
		{ID: "root-file", Type: "asset", Name: "Readme", Extension: ".txt", Path: "Readme.txt"},
	}, root)

	children, err := listEntityReferenceChildren(tx, "characters")
	require.NoError(t, err)
	require.Equal(t, []EntityReference{
		{ID: "heroes", Type: "collection", Name: "Heroes", ParentID: "characters", Path: "Characters/Heroes"},
		{ID: "hero", Type: "asset", Name: "Zeus", Extension: ".blend", ParentID: "characters", Path: "Characters/Zeus.blend"},
	}, children)
}
