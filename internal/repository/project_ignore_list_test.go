package repository

import (
	"clustta/internal/auth_service"
	"clustta/internal/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetIgnoreListUpdatesRemoteAndExactLocalProject(t *testing.T) {
	expectedIgnoreList := []string{"*.blend1", "renders"}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodPut, request.Method)
		require.Equal(t, "/ignore-list", request.URL.Path)
		var ignoreList []string
		require.NoError(t, json.NewDecoder(request.Body).Decode(&ignoreList))
		require.Equal(t, expectedIgnoreList, ignoreList)
		writer.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	projectPath := filepath.Join(t.TempDir(), "project.clst")
	db, err := utils.OpenDb(projectPath)
	require.NoError(t, err)
	_, err = db.Exec(`
		CREATE TABLE config (
			name TEXT PRIMARY KEY,
			value TEXT,
			mtime INTEGER,
			synced BOOLEAN
		)
	`)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	err = SetIgnoreList(projectPath, server.URL, expectedIgnoreList, auth_service.User{Id: "user-id"})

	require.NoError(t, err)
	db, err = utils.OpenDb(projectPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	var storedIgnoreList string
	require.NoError(t, db.Get(&storedIgnoreList, "SELECT value FROM config WHERE name = 'ignore_list'"))
	require.JSONEq(t, `["*.blend1", "renders"]`, storedIgnoreList)
}
