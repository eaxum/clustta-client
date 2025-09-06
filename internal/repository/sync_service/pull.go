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
		defaultWorkingDirRoot, err := settings.GetWorkingDirectory()
		if err != nil {
			return err
		}
		projectName := strings.TrimSuffix(filepath.Base(projectUri), filepath.Ext(projectUri))
		workingDir = filepath.Join(defaultWorkingDirRoot, studioDisplayName, projectName)
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
		extension := "clst"
		entries, err := os.ReadDir(projectsDir)
		if err != nil {
			return studioProjects, err
		}

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
					studioProjects = append(studioProjects, projectInfo)
				}

			}
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
			studioProjects[i].SyncToken = syncToken
		}

		return studioProjects, nil
	}
}
