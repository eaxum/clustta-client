package sync_service

import (
	"clustta/internal/auth_service"
	"clustta/internal/chunk_service"
	"clustta/internal/constants"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/settings"
	"clustta/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

func PullData(ctx context.Context, projectPath, remoteUrl string, userId string, pullChunk bool, syncOptions SyncOptions, callback func(int, int, string, string)) error {
	fmt.Printf("start pull for %s\n", projectPath)
	trueStart := time.Now()
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

	err = utils.SetIsClosed(tx, projectInfo.IsClosed)
	if err != nil {
		return err
	}
	err = utils.SetProjectIcon(tx, projectInfo.Icon)
	if err != nil {
		return err
	}
	err = utils.SetProjectIgnoreList(tx, projectInfo.IgnoreList)
	if err != nil {
		return err
	}

	isUpToDate := false

	if !syncOptions.Force && projectInfo.SyncToken != "" && projectInfo.SyncToken == syncToken {
		isUpToDate = true
		println("Project is up to date")
	}

	start := time.Now()
	data := ProjectData{}
	if isUpToDate {
		data, err = LoadUserData(tx, userId)
		if err != nil {
			return err
		}
	} else {
		data, err = FetchData(remoteUrl, userId)
		if err != nil {
			return err
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("data transfer took %s\n", elapsed)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	userRole := models.Role{}
	userRoleId := ""
	for _, user := range data.Users {
		if user.Id == userId {
			userRoleId = user.RoleId
			break
		}
	}
	for _, role := range data.Roles {
		if role.Id == userRoleId {
			userRole = role
			break
		}
	}

	start = time.Now()
	missingPreviews, err := CalculateMissingPreviews(tx, data)
	if err != nil {
		return err
	}
	elapsed = time.Since(start)
	fmt.Printf("preview processing took %s\n", elapsed)

	start = time.Now()
	if len(missingPreviews) > 0 {
		err = repository.PullPreviews(tx, remoteUrl, missingPreviews, callback)
		if err != nil {
			return err
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("%d preview download took %s\n", len(missingPreviews), elapsed)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if !isUpToDate {
		start = time.Now()
		err = ClearLocalDataDrop(tx)
		if err != nil {
			return err
		}
		elapsed = time.Since(start)
		fmt.Printf("clear data took %s\n", elapsed)

		err = utils.SetProjectSyncToken(tx, projectInfo.SyncToken)
		if err != nil {
			return err
		}

		start = time.Now()
		err = OverWriteProjectData(tx, data)
		if err != nil {
			return err
		}
		elapsed = time.Since(start)
		fmt.Printf("writing transfered data took %s\n", elapsed)

		err = utils.SetLastSyncTime(tx, utils.GetEpochTime())
		if err != nil {
			return err
		}

		err = utils.SetTablesToSynced(tx, ProjectTables)
		if err != nil {
			return err
		}
	}

	start = time.Now()
	missingChunks := []string{}
	allChunks := []string{}
	totalSize := 0
	if pullChunk && userRole.PullChunk {
		missingChunks, allChunks, totalSize, err = CalculateMissingChunks(tx, data, userId, syncOptions)
		if err != nil {
			return err
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("missing chunks took %s\n", elapsed)

	err = tx.Commit()
	if err != nil {
		return err
	}

	if pullChunk && userRole.PullChunk {
		if len(missingChunks) > 0 {
			err = chunk_service.PullStreamChunks(ctx, projectPath, remoteUrl, missingChunks, allChunks, totalSize, callback)
			if err != nil {
				return err
			}
		}
	}
	trueElapsed := time.Since(trueStart)
	fmt.Printf("total took %s\n", trueElapsed)
	return nil
}

func PullLatestCheckpoints(ctx context.Context, projectPath, remoteUrl string, userId string, callback func(int, int, string, string)) error {
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

	data, err := LoadUserData(tx, userId)
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	syncOptions := SyncOptions{
		OnlyLatestCheckpoints: true,
		Tasks:                 true,
		TaskDependencies:      true,
		Resources:             true,
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	missingChunks, allChunks, totalSize, err := CalculateMissingChunks(tx, data, userId, syncOptions)
	if err != nil {
		return err
	}

	err = tx.Rollback()
	if err != nil {
		return err
	}

	if len(missingChunks) > 0 {
		err = chunk_service.PullStreamChunks(ctx, projectPath, remoteUrl, missingChunks, allChunks, totalSize, callback)
		if err != nil {
			return err
		}
	}
	return nil
}

func CloneProject(ctx context.Context, remoteProjectUri string, projectUri string, studioDisplayName, workingDir string, user auth_service.User, syncOptions SyncOptions, callback func(int, int, string, string)) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	db, err := utils.OpenDb(projectUri)
	if err != nil {
		return err
	}
	defer db.Close()

	// statements := strings.Split(projectSchema, ";")
	err = utils.CreateSchema(db, repository.ProjectSchema)
	if err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	projectInfo, err := repository.GetProjectInfo(remoteProjectUri, user)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO config (name, value, mtime) VALUES ('project_id', ?, ?)", projectInfo.Id, utils.GetEpochTime())
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("INSERT INTO config (name, value, mtime) VALUES ('remote', ?, ?)", remoteProjectUri, utils.GetEpochTime())
	if err != nil {
		tx.Rollback()
		return err
	}
	err = utils.SetIsClosed(tx, projectInfo.IsClosed)
	if err != nil {
		return err
	}
	err = utils.SetProjectVersion(tx, projectInfo.Version)
	if err != nil {
		return err
	}
	err = utils.SetStudioName(tx, studioDisplayName)
	if err != nil {
		return err
	}

	if workingDir == "" {
		// Fallback to default location if no working directory specified
		defaultLocation, err := settings.GetDefaultLocation()
		if err != nil {
			return err
		}
		projectName := strings.TrimSuffix(filepath.Base(projectUri), filepath.Ext(projectUri))
		workingDir = filepath.Join(defaultLocation.Path, studioDisplayName, projectName)
	}

	err = utils.SetProjectWorkingDir(tx, workingDir)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	err = PullData(ctx, projectUri, remoteProjectUri, user.Id, true, syncOptions, callback)
	if err != nil {
		return err
	}
	return nil
}

func GetStudioProjects(user auth_service.User, url string, studioName string) ([]repository.ProjectInfo, error) {
	isLocal := studioName == "Personal"
	studioProjects := []repository.ProjectInfo{}
	projectsDir, err := settings.GetProjectDirectory()
	if err != nil {
		return studioProjects, err
	}
	studioProjectsDir := ""
	if isLocal {
		studioProjectsDir = projectsDir
	} else {
		sharedProjectsDir, err := settings.GetSharedProjectDirectory()
		if err != nil {
			return studioProjects, err
		}
		studioProjectsDir = filepath.Join(sharedProjectsDir, studioName)
	}
	os.MkdirAll(studioProjectsDir, os.ModePerm)
	if isLocal {

		//this finds all existing personal projects by checking in the projectsDir
		//it also checks in the user working directories for folders without a
		//corresponding clst file and appends them as untracked projects for the user to easily add later

		extension := "clst"
		entries, err := os.ReadDir(projectsDir)
		if err != nil {
			return studioProjects, err
		}

		// Track project names that have .clst files
		trackedProjectNames := make(map[string]bool)

		// Iterate over the directory entries
		for _, entry := range entries {
			// Check if the entry is a file and has the specified extension
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), extension) {
				projectPath := filepath.Join(projectsDir, entry.Name())

				fileInfo, err := entry.Info()
				if err != nil {
					return studioProjects, err
				}
				if fileInfo.Size() == 0 {
					repository.InitDB(projectPath, studioName, "", user, false)
				}

				valid, err := repository.VerifyProjectIntegrity(projectPath)
				if !valid || err != nil {
					continue
				}

				err = repository.UpdateProject(projectPath)
				if err != nil {
					return studioProjects, err
				}

				userInProject, err := repository.UserInProject(projectPath, user.Id)
				if err != nil {
					return studioProjects, err
				}
				if userInProject {
					projectInfo, err := repository.GetProjectInfo(projectPath, user)
					if err != nil {
						return studioProjects, err
					}
					projectInfo.Uri = projectPath
					projectInfo.Remote = projectPath
					projectInfo.IsDownloaded = true
					projectInfo.IsTracked = true
					studioProjects = append(studioProjects, projectInfo)

					// Track this project name (without .clst extension)
					projectName := strings.TrimSuffix(entry.Name(), "."+extension)
					trackedProjectNames[projectName] = true
				}

			}
		}

		// Get list of studio folders to exclude from untracked scanning
		studioFolders := make(map[string]bool)
		sharedProjectsDir, err := settings.GetSharedProjectDirectory()
		if err == nil {
			// Read all studio folders from shared_projects directory
			if entries, err := os.ReadDir(sharedProjectsDir); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						studioFolders[entry.Name()] = true
						fmt.Printf("Excluding studio folder from untracked scan: %s\n", entry.Name())
					}
				}
			}
		}

		// Scan all configured working locations for untracked projects
		projectLocations, err := settings.GetAllLocationPaths()
		if err != nil {
			return studioProjects, err
		}

		for _, location := range projectLocations {
			// For Personal projects, scan directly in the location path
			// No studio subdirectory needed
			locationScanPath := location.Path

			// Skip if path doesn't exist yet
			if _, err := os.Stat(locationScanPath); os.IsNotExist(err) {
				fmt.Printf("Location path does not exist: %s\n", locationScanPath)
				continue
			}

			untrackedInLocation, err := GetUntrackedProjects(locationScanPath, trackedProjectNames, studioFolders, location.ID)
			if err != nil {
				// Log but don't fail entire operation
				fmt.Printf("Warning: Failed to scan location %s: %v\n", location.Name, err)
				continue
			}

			studioProjects = append(studioProjects, untrackedInLocation...)
		}

		return studioProjects, nil
		// projects := []ProjectInfo{}
		// totalProjects := len(studioProjects)
		// for i, project := range studioProjects {
		// 	innerCallBack := func(current int, total int, message string, extraMessage string) {
		// 		callback(project.Name, current, total, i+1, totalProjects)
		// 	}
		// 	projectPath := filepath.Join(studioProjectsDir, project.Name) + ".clst"
		// 	projectUri := project.Uri
		// 	err = CreateProject(projectPath, project.Name, projectUrl, innerCallBack)
		// 	if err != nil {
		// 		return projects, err
		// 	}
		// 	projectInfo, err := GetProjectInfo(projectPath)
		// 	if err != nil {
		// 		return projects, err
		// 	}
		// 	projects = append(projects, projectInfo)
		// }

		// return projects, nil
	} else {
		studioProjectUrl := url + "/projects"
		req, err := http.NewRequest("GET", studioProjectUrl, nil)
		if err != nil {
			return studioProjects, err
		}
		userJson, err := json.Marshal(user)
		if err != nil {
			return studioProjects, err
		}
		req.Header.Set("Clustta-Agent", constants.USER_AGENT)
		req.Header.Set("UserData", string(userJson))
		req.Header.Set("UserId", user.Id)

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			return studioProjects, err
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return studioProjects, err
			}
			return studioProjects, errors.New(string(body))
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return studioProjects, err
		}

		err = json.Unmarshal(body, &studioProjects)
		if err != nil {
			return studioProjects, err
		}

		for i, studioProject := range studioProjects {
			workingDir := ""
			projectPath := filepath.Join(studioProjectsDir, studioProject.Name) + ".clst"
			isDownloaded := utils.FileExists(projectPath)
			projectUrl := url + "/" + studioProject.Name
			syncToken := ""
			if isDownloaded {
				valid, err := repository.VerifyProjectIntegrity(projectPath)
				if !valid || err != nil {
					err := os.Remove(projectPath)
					if err != nil {
						return studioProjects, err
					}
					isDownloaded = false
				} else {
					repository.UpdateProject(projectPath)
					dbConn, err := utils.OpenDb(projectPath)
					if err != nil {
						return studioProjects, err
					}
					defer dbConn.Close()
					tx, err := dbConn.Beginx()
					if err != nil {
						return studioProjects, err
					}
					defer tx.Rollback()
					isSynced, err := repository.IsProjectPreviewSynced(tx)
					if err != nil {
						return studioProjects, err
					}
					if isSynced && studioProject.PreviewId != "" {
						projectPreviewId := studioProject.PreviewId
						missingPreviews, err := CalculateMissingPreviews(tx, ProjectData{
							ProjectPreview: projectPreviewId,
						})
						if err != nil {
							return studioProjects, err
						}
						if len(missingPreviews) > 0 {
							err = repository.PullPreviews(tx, projectUrl, missingPreviews, func(i1, i2 int, s1, s2 string) {})
							if err != nil {
								return studioProjects, err
							}
						}
						_, err = tx.Exec(`
							INSERT INTO config (name, value, mtime, synced)
							VALUES ('project_preview', $1, $2, 1)
							ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime, synced = 1
						`, projectPreviewId, utils.GetEpochTime())
						if err != nil {
							return studioProjects, err
						}

					}

					workingDir, err = utils.GetProjectWorkingDir(tx)
					if err != nil {
						return studioProjects, err
					}

					syncToken, err = utils.GetProjectSyncToken(tx)
					if err != nil {
						return studioProjects, err
					}

					err = tx.Commit()
					if err != nil {
						return studioProjects, err
					}
				}

			}
			studioProjects[i].HasRemote = true
			studioProjects[i].Uri = projectPath
			studioProjects[i].Remote = projectUrl
			studioProjects[i].WorkingDirectory = workingDir
			studioProjects[i].IsDownloaded = isDownloaded
			studioProjects[i].IsTracked = true
			studioProjects[i].SyncToken = syncToken
		}

		return studioProjects, nil
	}
}

func GetUntrackedProjects(projectsDir string, trackedProjectNames map[string]bool, studioFolders map[string]bool, locationID string) ([]repository.ProjectInfo, error) {
	untrackedProjects := []repository.ProjectInfo{}

	// Read all entries in the projects directory
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return untrackedProjects, err
	}

	fmt.Printf("Scanning for untracked projects in: %s (location: %s)\n", projectsDir, locationID)
	fmt.Printf("Tracked project names: %v\n", trackedProjectNames)
	fmt.Printf("Studio folders to exclude: %v\n", studioFolders)

	// Iterate over directory entries
	for _, entry := range entries {
		// Only process directories, skip files
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories (starting with .)
		if strings.HasPrefix(entry.Name(), ".") {
			fmt.Printf("Skipping hidden directory: %s\n", entry.Name())
			continue
		}

		// Skip studio folders (these are for non-Personal projects)
		if studioFolders[entry.Name()] {
			fmt.Printf("Skipping studio folder: %s\n", entry.Name())
			continue
		}

		// Check if this directory already has a corresponding .clst project
		if trackedProjectNames[entry.Name()] {
			fmt.Printf("Skipping tracked project directory: %s\n", entry.Name())
			continue
		}

		// Create minimal ProjectInfo for untracked project
		dirPath := filepath.Join(projectsDir, entry.Name())
		projectId := uuid.New().String()
		untrackedProject := repository.ProjectInfo{
			Id:               projectId,
			Name:             entry.Name(),
			WorkingDirectory: dirPath,
			LocationID:       locationID,
			IsTracked:        false,
			IsDownloaded:     false,
			Valid:            true,
		}

		fmt.Printf("Found untracked project: %s (ID: %s) at %s\n", entry.Name(), projectId, dirPath)
		untrackedProjects = append(untrackedProjects, untrackedProject)
	}

	fmt.Printf("Total untracked projects found: %d\n", len(untrackedProjects))
	for i, proj := range untrackedProjects {
		fmt.Printf("  [%d] Name: %s, WorkingDir: %s, LocationID: %s, IsTracked: %v\n", i+1, proj.Name, proj.WorkingDirectory, proj.LocationID, proj.IsTracked)
	}

	return untrackedProjects, nil
}
