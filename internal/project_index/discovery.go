package projectindex

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"clustta/internal/auth_service"
	"clustta/internal/settings"
	"clustta/internal/utils"
)

var discoveryState = struct {
	sync.Mutex
	accountID string
	complete  bool
}{}

func EnsureLocalProjects() error {
	user, err := activeUser()
	if err != nil {
		return err
	}
	discoveryState.Lock()
	defer discoveryState.Unlock()
	if discoveryState.accountID == user.ID && discoveryState.complete {
		return nil
	}
	complete, err := isBackfillComplete()
	if err != nil {
		return err
	}
	if complete {
		discoveryState.accountID = user.ID
		discoveryState.complete = true
		return nil
	}

	paths, err := localProjectFiles()
	if err != nil {
		return err
	}
	entries := make([]Entry, 0, len(paths))
	for _, path := range paths {
		entry, inspectErr := inspectProject(path)
		if inspectErr != nil {
			log.Printf("Failed to inspect local project %s: %v", path, inspectErr)
			continue
		}
		entries = append(entries, entry)
	}
	if err = UpsertMany(entries); err != nil {
		return err
	}
	if err = markBackfillComplete(); err != nil {
		return err
	}
	discoveryState.accountID = user.ID
	discoveryState.complete = true
	return nil
}

func resetDiscovery() {
	discoveryState.Lock()
	defer discoveryState.Unlock()
	discoveryState.accountID = ""
	discoveryState.complete = false
}

type activeUserIdentity struct {
	ID string
}

func activeUser() (activeUserIdentity, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return activeUserIdentity{}, err
	}
	return activeUserIdentity{ID: user.Id}, nil
}

func localProjectFiles() ([]string, error) {
	personalDirectory, err := settings.GetProjectDirectory()
	if err != nil {
		return nil, err
	}
	sharedDirectory, err := settings.GetSharedProjectDirectory()
	if err != nil {
		return nil, err
	}

	files := []string{}
	for _, root := range []string{personalDirectory, sharedDirectory} {
		if strings.TrimSpace(root) == "" {
			continue
		}
		err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				if os.IsNotExist(walkErr) {
					return nil
				}
				return walkErr
			}
			if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".clst") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	return files, nil
}

func inspectProject(databasePath string) (Entry, error) {
	database, err := utils.OpenDb(databasePath)
	if err != nil {
		return Entry{}, err
	}
	defer database.Close()
	transaction, err := database.Beginx()
	if err != nil {
		return Entry{}, err
	}
	defer transaction.Rollback()

	projectID, err := utils.GetProjectId(transaction)
	if err != nil {
		return Entry{}, err
	}
	projectName, err := utils.GetProjectName(transaction)
	if err != nil {
		return Entry{}, err
	}
	studioName, err := utils.GetStudioName(transaction)
	if err != nil {
		studioName = "Personal"
	}
	workingDirectory, err := utils.GetProjectWorkingDir(transaction)
	if err != nil {
		return Entry{}, err
	}
	return Entry{
		ProjectID:        projectID,
		ProjectName:      projectName,
		StudioName:       studioName,
		DatabasePath:     databasePath,
		WorkingDirectory: workingDirectory,
	}, nil
}
