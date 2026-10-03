package sync_service

import (
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/repositorypb"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func uploadFixture(t *testing.T) (string, *sqlx.DB, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "project.clst")
	db := sqlx.MustOpen("sqlite3", path)
	t.Cleanup(func() { db.Close() })
	db.MustExec(repository.ProjectSchema)
	db.MustExec(`
		INSERT INTO config(name,value,mtime,synced) VALUES ('project_id','original',1,1),('working_dir','old',1,1);
		INSERT INTO role(id,mtime,name,synced) VALUES ('local-admin',1,'admin',1);
		INSERT INTO user(id,mtime,added_at,username,email,first_name,last_name,role_id,synced) VALUES ('archive-user',1,'2026-01-01','archive','archive@example.com','Archive','User','local-admin',1);
		INSERT INTO collection_type(id,mtime,name,icon,synced) VALUES ('local-folder',1,'Generic','folder',1),('shot',1,'Shot','camera',1);
		INSERT INTO asset_type(id,mtime,name,icon,synced) VALUES ('local-asset',1,'generic','generic',1),('script',1,'Script','book',1);
		INSERT INTO dependency_type(id,mtime,name,synced) VALUES ('local-linked',1,'linked',1);
		INSERT INTO status(id,mtime,name,short_name,color,synced) VALUES ('local-status',1,'todo','todo','gray',1);
		INSERT INTO collection(id,created_at,mtime,name,collection_type_id,parent_id,synced) VALUES ('shots','2026-01-01',1,'Shots','shot','',1),('root','2026-01-01',1,'Root','local-folder','',1);
		UPDATE collection SET description='';
		INSERT INTO asset(id,created_at,mtime,name,extension,asset_type_id,status_id,collection_id,synced) VALUES ('script-asset','2026-01-01',1,'launch','.py','script','local-status','shots',1),('dependency-asset','2026-01-01',1,'library','.blend','local-asset','local-status','root',1);
		INSERT INTO asset_dependency(id,mtime,asset_id,dependency_id,dependency_type_id,synced) VALUES ('asset-edge',1,'script-asset','dependency-asset','local-linked',1);
		INSERT INTO collection_dependency(id,mtime,asset_id,dependency_id,dependency_type_id,synced) VALUES ('collection-edge',1,'script-asset','root','local-linked',1);
		INSERT INTO asset_checkpoint(id,created_at,mtime,asset_id,xxhash_checksum,time_modified,file_size,chunks,comment,author_id,synced) VALUES ('checkpoint',1,1,'script-asset','checksum',1,4,'chunk','initial','author',1);
		INSERT INTO workflow_collection(id,mtime,name,collection_type_id,workflow_id,synced) VALUES ('workflow-folder',1,'Folder','local-folder','workflow',1);
		INSERT INTO workflow_asset(id,mtime,name,asset_type_id,workflow_id,template_id,synced) VALUES ('workflow-asset',1,'Asset','local-asset','workflow','template',1);
		INSERT INTO workflow_link(id,mtime,name,collection_type_id,workflow_id,linked_workflow_id,synced) VALUES ('workflow-link',1,'Link','local-folder','workflow','other',1);
	`)
	data, err := proto.Marshal(&repositorypb.ProjectData{
		CollectionTypes: []*repositorypb.CollectionType{{Id: "remote-folder", Name: "generic", Icon: "folder"}},
		AssetTypes:      []*repositorypb.AssetType{{Id: "remote-asset", Name: "generic", Icon: "generic"}},
		Statuses:        []*repositorypb.Status{{Id: "remote-status", Name: "todo"}},
		DependencyTypes: []*repositorypb.DependencyType{{Id: "remote-linked", Name: "linked"}},
		Roles:           []*repositorypb.Role{{Id: "remote-admin", Name: "admin"}},
		Users: []*repositorypb.User{{
			Id: "user", Mtime: 2, AddedAt: "2026-01-02", Username: "uploader", Email: "uploader@example.com",
			FirstName: "Upload", LastName: "User", RoleId: "remote-admin",
		}},
	})
	require.NoError(t, err)
	encoder, err := zstd.NewWriter(nil)
	require.NoError(t, err)
	compressed := encoder.EncodeAll(data, nil)
	encoder.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/data" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		_, _ = w.Write(compressed)
	}))
	t.Cleanup(server.Close)
	return path, db, server.URL
}

func TestUploadPreparationPreservesCustomTypes(t *testing.T) {
	for _, cloud := range []bool{false, true} {
		name := "personal"
		if cloud {
			name = "cloud"
		}
		t.Run(name, func(t *testing.T) {
			path, db, remote := uploadFixture(t)
			var err error
			if cloud {
				err = PrepareProjectForCloudUpload(path, "remote-project", remote, "workspace", "user")
			} else {
				err = PrepareProjectForUpload(path, repository.ProjectInfo{Id: "remote-project"}, remote, "workspace", "user")
			}
			require.NoError(t, err)
			for query, expected := range map[string]string{
				"SELECT icon FROM collection_type WHERE id='shot'":          "camera",
				"SELECT icon FROM asset_type WHERE id='script'":             "book",
				"SELECT status_id FROM asset WHERE id='script-asset'":       "remote-status",
				"SELECT collection_type_id FROM workflow_collection":        "remote-folder",
				"SELECT collection_type_id FROM workflow_link":              "remote-folder",
				"SELECT asset_type_id FROM workflow_asset":                  "remote-asset",
				"SELECT dependency_type_id FROM asset_dependency":           "remote-linked",
				"SELECT dependency_type_id FROM collection_dependency":      "remote-linked",
				"SELECT id FROM dependency_type WHERE name='linked'":        "remote-linked",
				"SELECT value FROM config WHERE name='working_dir'":         "workspace",
				"SELECT chunks FROM asset_checkpoint WHERE id='checkpoint'": "chunk",
			} {
				var actual string
				require.NoError(t, db.Get(&actual, query))
				require.Equal(t, expected, actual)
			}
			var count int
			var uploadUser models.User
			require.NoError(t, db.Get(&uploadUser, "SELECT * FROM user WHERE id='user'"))
			require.Equal(t, "remote-admin", uploadUser.RoleId)
			require.True(t, uploadUser.Synced)
			var archiveRoleId string
			require.NoError(t, db.Get(&archiveRoleId, "SELECT role_id FROM user WHERE id='archive-user'"))
			require.Equal(t, "remote-admin", archiveRoleId)
			require.NoError(t, db.Get(&count, "SELECT count(*) FROM tomb"))
			require.Zero(t, count)
			require.NoError(t, db.Get(&count, "SELECT count(*) FROM full_collection"))
			require.Equal(t, 2, count)
			tx := db.MustBegin()
			defer tx.Rollback()
			changed, err := LoadChangedData(tx)
			require.NoError(t, err)
			require.Len(t, changed.CollectionTypes, 2)
			require.Len(t, changed.AssetTypes, 2)
			require.Len(t, changed.AssetCheckpoints, 1)
			require.Len(t, changed.Assets, 2)
			require.Len(t, changed.Users, 1)
			require.Equal(t, "archive-user", changed.Users[0].Id)
			payload, err := proto.Marshal(&repositorypb.ProjectData{CollectionTypes: repository.ToPbCollectionTypes(changed.CollectionTypes), AssetTypes: repository.ToPbAssetTypes(changed.AssetTypes)})
			require.NoError(t, err)
			var decoded repositorypb.ProjectData
			require.NoError(t, proto.Unmarshal(payload, &decoded))
			require.Equal(t, changed.CollectionTypes, repository.FromPbCollectionTypes(decoded.CollectionTypes))
			require.Equal(t, changed.AssetTypes, repository.FromPbAssetTypes(decoded.AssetTypes))
		})
	}
}

func TestUploadPreparationRollsBackInvalidReferences(t *testing.T) {
	path, db, remote := uploadFixture(t)
	db.MustExec("UPDATE collection SET collection_type_id='missing' WHERE id='shots'")
	err := PrepareProjectForCloudUpload(path, "remote-project", remote, "workspace", "user")
	require.ErrorContains(t, err, "reference missing collection_type")
	var id string
	require.NoError(t, db.Get(&id, "SELECT status_id FROM asset"))
	require.Equal(t, "local-status", id)
	require.NoError(t, db.Get(&id, "SELECT value FROM config WHERE name='project_id'"))
	require.Equal(t, "original", id)
}

func TestUploadPreparationRejectsIconCollision(t *testing.T) {
	path, db, remote := uploadFixture(t)
	db.MustExec("UPDATE collection_type SET icon='tree' WHERE id='local-folder'")
	db.MustExec("UPDATE collection_type SET icon='folder' WHERE id='shot'")
	err := PrepareProjectForCloudUpload(path, "remote-project", remote, "workspace", "user")
	require.ErrorContains(t, err, "choose a different icon")
}

func TestMarkAllTablesUnsyncedReportsErrors(t *testing.T) {
	_, db, _ := uploadFixture(t)
	db.MustExec("CREATE TRIGGER reject_unsynced BEFORE UPDATE ON collection BEGIN SELECT RAISE(ABORT, 'blocked'); END")
	tx := db.MustBegin()
	defer tx.Rollback()
	require.ErrorContains(t, MarkAllTablesUnsynced(tx), "failed to mark collection unsynced")
}
