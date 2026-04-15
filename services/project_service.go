package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/constants"
	"clustta/internal/error_service"
	"clustta/internal/ignore"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/sync_service"
	"clustta/internal/settings"
	"clustta/internal/utils"
	"clustta/output"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type ProjectService struct {
}

type UntrackedItems struct {
	Files   []models.UntrackedAsset      `json:"assets"`
	Folders []models.UntrackedCollection `json:"collections"`
}

func (p *ProjectService) CreateProject(projectUri, studioName, workingDir, templateName, hostingMode, studioId string) (repository.ProjectInfo, error) {
	if studioName == "" {
		return repository.ProjectInfo{}, errors.New("studio name can't be empty")
	}

	// For cloud studios, construct the proper API URL
	if hostingMode == "cloud" && studioId != "" {
		projectName := filepath.Base(projectUri)
		projectUri = constants.HOST + "/studio/" + studioId + "/" + projectName
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return repository.ProjectInfo{}, err
	}

	fmt.Println(user)
	projectInfo, err := repository.CreateProject(projectUri, studioName, workingDir, templateName, "", user)
	if err != nil {
		if !utils.IsValidURL(projectUri) &&
			utils.FileExists(projectUri) &&
			err != error_service.ErrInvalidProjectExists &&
			err != error_service.ErrProjectExists {
			journal := projectUri + "-journal"
			err := os.Remove(projectUri)
			if err != nil {
				return projectInfo, err
			}
			if utils.FileExists(journal) {
				err = os.Remove(journal)
				if err != nil {
					return projectInfo, err
				}
			}

		}
		return projectInfo, err
	}
	projectInfo.WorkingDirectory = workingDir
	return projectInfo, nil
}

// MakeProjectRemote uploads a local project to Clustta Cloud as a remote project.
// Creates the remote project, remaps IDs to match remote, pushes all data, and stores the remote URL locally.
func (p *ProjectService) MakeProjectRemote(projectPath string) error {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}

	authHost := auth_service.GetAuthHost()
	if authHost == "" {
		return errors.New("not connected to Clustta Cloud")
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

	projectName, err := utils.GetProjectName(tx)
	if err != nil {
		return err
	}
	workingDir, _ := utils.GetProjectWorkingDir(tx)
	projectId, _ := utils.GetProjectId(tx)
	tx.Rollback()
	dbConn.Close()

	remoteURL := fmt.Sprintf("%s/user/%s/%s", authHost, user.Id, projectName)

	// Ensure the local project schema is up to date before reading data
	err = repository.UpdateProject(projectPath)
	if err != nil {
		return fmt.Errorf("failed to update local project schema: %w", err)
	}

	// Create the remote project on the server
	remoteProjectInfo, err := repository.CreateProject(remoteURL, "", "", "No Template", projectId, user)
	if err != nil {
		return fmt.Errorf("failed to create remote project: %w", err)
	}

	// Remap local IDs (project_id, roles, statuses, types) to match the remote project
	err = sync_service.PrepareProjectForUpload(projectPath, remoteProjectInfo, remoteURL, workingDir, user.Id)
	if err != nil {
		return fmt.Errorf("failed to remap project IDs: %w", err)
	}

	app := application.Get()
	progressCallback := func(current int, total int, message string, extraMessage string) {
		app.Event.Emit("progress-update", output.ProgressReport{
			Title:        "Uploading to Cloud",
			Message:      message,
			Percentage:   float64(current) / float64(total) * 100,
			Current:      1,
			Total:        1,
			ExtraMessage: extraMessage,
		})
	}

	// Push all data (metadata, chunks, previews)
	err = sync_service.PushData(context.Background(), projectPath, remoteURL, user.Id, progressCallback)
	if err != nil {
		return fmt.Errorf("failed to upload project data: %w", err)
	}

	InvalidateRemoteCache(projectPath)
	return nil
}

// RemoveProjectFromRemote deletes the remote copy and clears the local remote URL.
// The local project data is preserved.
func (p *ProjectService) RemoveProjectFromRemote(projectPath string) error {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
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

	remoteURL, err := utils.GetRemoteUrl(tx)
	if err != nil || remoteURL == "" {
		return errors.New("project is not remote")
	}

	// Delete the remote project from the server
	err = repository.DeleteRemoteProject(remoteURL, "", user)
	if err != nil {
		return fmt.Errorf("failed to delete remote project: %w", err)
	}

	// Clear the remote URL locally
	err = utils.SetRemoteUrl(tx, "")
	if err != nil {
		return fmt.Errorf("failed to clear remote URL: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	InvalidateRemoteCache(projectPath)
	return nil
}

func (p *ProjectService) ApplyTemplate(projectPath, templateName string) error {
	if templateName == "" || templateName == "No Template" {
		return nil
	}

	if !utils.FileExists(projectPath) {
		return errors.New("project file not found")
	}

	templatesPath, err := settings.GetUserProjectTemplatesPath()
	if err != nil {
		return err
	}

	templatePath := filepath.Join(templatesPath, templateName+".clst")

	err = repository.LoadProjectTemplateData(projectPath, templatePath)
	if err != nil {
		return err
	}

	return nil
}

func (p *ProjectService) ResetDefaultTemplates() error {
	return repository.ResetDefaultTemplates()
}

func (p *ProjectService) ToggleCloseProject(projectUri, studioName string) error {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	err = repository.ToggleCloseProject(projectUri, studioName, user)
	if err != nil {
		return err
	}
	return nil
}

func (p *ProjectService) GetIsClose(projectPath string) (bool, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return false, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	return utils.GetIsClosed(tx)
}

func (p *ProjectService) GetIgnoreList(projectPath string) ([]string, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []string{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []string{}, err
	}
	defer tx.Rollback()
	return repository.GetIgnoreList(tx)
}

func (p *ProjectService) SetIgnoreList(projectUri, studioName string, ignoreList []string) error {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	return repository.SetIgnoreList(projectUri, studioName, ignoreList, user)
}

func (p *ProjectService) CloseProject(projectPath string) error {
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

	err = utils.SetIsClosed(tx, false)
	if err != nil {
		return err
	}

	return nil
}

func (p *ProjectService) ProjectInfo(projectPath string) (repository.ProjectInfo, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return repository.ProjectInfo{}, err
	}
	info, err := repository.GetProjectInfo(projectPath, user)
	if err != nil {
		return repository.ProjectInfo{}, err
	}
	return info, nil
}

func (p *ProjectService) GetSyncToken(projectUri string) (string, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return "", err
	}
	syncToken, err := repository.GetSyncToken(projectUri, user)
	if err != nil {
		return "", err
	}
	return syncToken, nil
}

func (p *ProjectService) ProjectsInfo(projectPaths []string) ([]repository.ProjectInfo, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []repository.ProjectInfo{}, err
	}
	infos := []repository.ProjectInfo{}
	for _, projectPath := range projectPaths {
		info, err := repository.GetProjectInfo(projectPath, user)
		if err != nil {
			return infos, err
		}
		infos = append(infos, info)
	}

	return infos, nil
}

func (p *ProjectService) Purge(projectPath string) error {
	app := application.Get()

	if !utils.FileExists(projectPath) {
		return error_service.ErrProjectNotFound
	}

	progress := output.ProgressReport{
		Title:      "Clearing Trash",
		Message:    filepath.Base(projectPath),
		Percentage: 0,
		Current:    1,
		Total:      2,
	}
	app.Event.Emit("progress-update", progress)

	err := repository.Purge(projectPath)
	if err != nil {
		return err
	}

	progress = output.ProgressReport{
		Title:      "Cleaning Up",
		Message:    filepath.Base(projectPath),
		Percentage: 0,
		Current:    2,
		Total:      2,
	}
	app.Event.Emit("progress-update", progress)
	err = repository.Vacuum(projectPath)
	if err != nil {
		return err
	}
	progress = output.ProgressReport{
		Title:      "Cleaning Up. This may take some time",
		Message:    filepath.Base(projectPath),
		Percentage: 100,
		Current:    2,
		Total:      2,
	}
	app.Event.Emit("progress-update", progress)
	return nil
}

// TrimProject removes all chunks and previews from the project database to reduce file size.
// The data can be re-fetched from the remote when needed.
func (p *ProjectService) TrimProject(projectPath string) error {
	app := application.Get()

	if !utils.FileExists(projectPath) {
		return error_service.ErrProjectNotFound
	}

	progress := output.ProgressReport{
		Title:      "Trimming Project",
		Message:    filepath.Base(projectPath),
		Percentage: 0,
		Current:    1,
		Total:      2,
	}
	app.Event.Emit("progress-update", progress)

	err := repository.TrimProject(projectPath)
	if err != nil {
		return err
	}

	progress = output.ProgressReport{
		Title:      "Compacting Database",
		Message:    filepath.Base(projectPath),
		Percentage: 50,
		Current:    2,
		Total:      2,
	}
	app.Event.Emit("progress-update", progress)

	err = repository.Vacuum(projectPath)
	if err != nil {
		return err
	}

	progress = output.ProgressReport{
		Title:      "Project Trimmed",
		Message:    filepath.Base(projectPath),
		Percentage: 100,
		Current:    2,
		Total:      2,
	}
	app.Event.Emit("progress-update", progress)

	return nil
}

// DeleteRemoteProject permanently deletes a project from the studio server.
// This requires admin permissions and cannot be undone.
func (p *ProjectService) DeleteRemoteProject(projectUri, studioName string) error {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	return repository.DeleteRemoteProject(projectUri, studioName, user)
}

// LeaveProject removes the current user as a collaborator from a remote project.
// The project remote URL is used to construct the leave endpoint.
func (p *ProjectService) LeaveProject(remoteUrl string) error {
	return repository.LeaveProject(remoteUrl)
}

func (p *ProjectService) AddUser(projectPath, email, roleName string) (models.User, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.User{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.User{}, err
	}
	defer tx.Rollback()

	user, err := repository.AddUser(tx, email, roleName)
	if err != nil {
		return models.User{}, err
	}
	tx.Commit()
	enqueueUserWriteThrough(projectPath, user)
	return user, nil
}

// AddUserSynced adds a user to the local project and marks them as synced.
// Used when the server already has the user data via write-through.
func (p *ProjectService) AddUserSynced(projectPath, email, roleName string) (models.User, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return models.User{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return models.User{}, err
	}
	defer tx.Rollback()

	user, err := repository.AddUser(tx, email, roleName)
	if err != nil {
		return models.User{}, err
	}

	err = utils.SetRowsSynced(tx, "user", []string{user.Id})
	if err != nil {
		return models.User{}, err
	}

	tx.Commit()
	return user, nil
}

func (p *ProjectService) ChangeRole(projectPath, userId, roleName string) error {
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

	err = repository.ChangeUserRoleByName(tx, userId, roleName)
	if err != nil {
		if errors.Is(err, error_service.ErrMustHaveAdmin) {
			return error_service.ErrMustHaveAdmin
		}
		return err
	}
	user, err := repository.GetUser(tx, userId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	enqueueUserWriteThrough(projectPath, user)
	return nil
}

func (p *ProjectService) RemoveUser(projectPath, userId string) error {
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

	err = repository.RemoveUser(tx, userId)
	if err != nil {
		if errors.Is(err, error_service.ErrUserHaveAssetAssigned) {
			return error_service.ErrUserHaveAssetAssigned
		}
		return err
	}
	tomb, err := repository.GetTomb(tx, userId)
	if err != nil {
		return err
	}
	tx.Commit()
	enqueueTombWriteThrough(projectPath, tomb)
	return nil
}

// RemoveUserSynced removes a user from the local project and marks the tomb as synced.
// Used when the server already has the deletion via its own endpoint.
func (p *ProjectService) RemoveUserSynced(projectPath, userId string) error {
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

	err = repository.RemoveUser(tx, userId)
	if err != nil {
		if errors.Is(err, error_service.ErrUserHaveAssetAssigned) {
			return error_service.ErrUserHaveAssetAssigned
		}
		return err
	}

	err = utils.SetRowsSynced(tx, "tomb", []string{userId})
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (p *ProjectService) GetStudioProjects(url, name, hostingMode, studioId string) ([]repository.ProjectInfo, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []repository.ProjectInfo{}, err
	}

	projects, err := sync_service.GetStudioProjects(user, url, name, hostingMode, studioId)
	if err != nil {
		return projects, err
	}
	return projects, nil
}

func (p *ProjectService) GetTemplates() ([]repository.ProjectInfo, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []repository.ProjectInfo{}, err
	}
	templateProjects := []repository.ProjectInfo{}

	templatesDir, err := settings.GetUserProjectTemplatesPath()
	if err != nil {
		return templateProjects, err
	}

	extension := "clst"
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return templateProjects, err
	}

	// Iterate over the directory entries
	for _, entry := range entries {
		// Check if the entry is a file and has the specified extension
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), extension) {
			projectPath := filepath.Join(templatesDir, entry.Name())

			err := repository.UpdateProject(projectPath)
			if err != nil {
				return templateProjects, err
			}

			userInProject, err := repository.UserInProject(projectPath, user.Id)
			if err != nil {
				return templateProjects, err
			}
			if userInProject {
				projectInfo, err := repository.GetProjectInfo(projectPath, user)
				if err != nil {
					return templateProjects, err
				}
				projectInfo.Uri = projectPath
				projectInfo.Remote = projectPath
				projectInfo.IsDownloaded = true

				templateProjects = append(templateProjects, projectInfo)
			}

		}
	}

	return templateProjects, nil
}

func (p *ProjectService) UserInProject(projectPath, userId string) (bool, error) {
	if !utils.FileExists(projectPath) {
		return false, error_service.ErrProjectNotFound
	}
	userInProject, err := repository.UserInProject(projectPath, userId)
	if err != nil {
		return false, err
	}
	return userInProject, nil
}

func (p *ProjectService) UpdatePreview(projectPath, previewPath string) error {
	if !utils.FileExists(projectPath) {
		return error_service.ErrProjectNotFound
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

	err = repository.SetProjectPreview(tx, previewPath)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (p *ProjectService) UpdateIcon(projectUri, studioName, iconValue string) error {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	return repository.SetIcon(projectUri, studioName, iconValue, user)
}

func (p *ProjectService) GetPreview(projectPath string) (string, error) {
	if !utils.FileExists(projectPath) {
		return "", error_service.ErrProjectNotFound
	}

	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return "", err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	preview, err := repository.GetProjectPreview(tx)
	if err != nil {
		if err.Error() == "no preview" {
			return "", nil
		}
		return "", err
	}

	base64Str := base64.StdEncoding.EncodeToString(preview.Preview)

	return base64Str, nil
}

func (p *ProjectService) UpdateWorkingDirectory(projectUri, studioName, newWorkingDir string) error {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	return repository.UpdateProjectWorkingDirectory(projectUri, studioName, newWorkingDir, user)
}

func (p *ProjectService) Rename(projectUri, studioName, newName string) error {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return err
	}
	return repository.RenameProject(projectUri, studioName, newName, user)
}

func normalizeExtensions(extensions []string) []string {
	normalized := make([]string, len(extensions))
	for i, ext := range extensions {
		normalized[i] = strings.TrimPrefix(ext, ".")
	}
	return normalized
}

func isIgnoredExtension(path string, ignoredExtensions []string) bool {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	for _, ignored := range ignoredExtensions {
		if ext == ignored {
			return true
		}
	}
	return false
}

func (p *ProjectService) GetFolderUntrackedItems(
	projectWorkingDir string,
	directory string,
	ignoreList []string,
	tracked []string) (UntrackedItems, error) {

	// Initialize result structure with thread-safe slices
	untracked := UntrackedItems{
		Files:   make([]models.UntrackedAsset, 0),
		Folders: make([]models.UntrackedCollection, 0),
	}

	ignoreObject := ignore.CompileIgnoreLines(ignoreList...)

	// list items in directory
	entries, err := os.ReadDir(directory)
	if err != nil {
		return untracked, err
	}

	for _, entry := range entries {
		println(entry.Name())
		path := filepath.Join(directory, entry.Name())
		relativePath, err := filepath.Rel(projectWorkingDir, path)
		if err != nil {
			return untracked, err
		}
		println(relativePath)
		if ignoreObject.MatchesPath(relativePath) {
			continue
		}
		// no recursion for now
		if entry.IsDir() {
			// check if directory is tracked
			if !utils.Contains(tracked, path) {
				collection := models.UntrackedCollection{
					Id:             utils.GetMD5Hash(path),
					Name:           entry.Name(),
					FilePath:       path,
					CollectionPath: relativePath,
				}
				untracked.Folders = append(untracked.Folders, collection)
			}
		} else {
			// check if file is tracked
			if !utils.Contains(tracked, path) {
				assetName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				asset := models.UntrackedAsset{
					Id:             utils.GetMD5Hash(path),
					Name:           assetName,
					FilePath:       path,
					AssetPath:      relativePath,
					CollectionPath: filepath.ToSlash(filepath.Dir(relativePath)),
					Extension:      filepath.Ext(entry.Name()),
				}
				untracked.Files = append(untracked.Files, asset)
			}
		}
	}

	return untracked, nil
}

func processDirectory(
	dir string,
	projectWorkingDir string,
	trackedFiles map[string]bool,
	trackedFolders map[string]bool,
	clusttaIgnore *ignore.GitIgnore,
	filesChan chan<- string,
	foldersChan chan<- string,
	errorsChan chan<- error) {

	relativePath, err := filepath.Rel(projectWorkingDir, dir)
	if err != nil {
		errorsChan <- err
		return
	}
	relativePath = utils.NormalizePath(relativePath)

	if strings.HasPrefix(filepath.Base(dir), ".") {
		return
	}
	if !trackedFolders[dir] && !clusttaIgnore.MatchesPath(relativePath) {
		foldersChan <- relativePath
		// return // Skip processing contents of untracked directories
	}

	// Process directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		errorsChan <- err
		return
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			// Recursive call for subdirectories
			processDirectory(path, projectWorkingDir, trackedFiles, trackedFolders,
				clusttaIgnore, filesChan, foldersChan, errorsChan)
		} else {
			processFile(path, projectWorkingDir, trackedFiles, clusttaIgnore, filesChan)
		}
	}
}

func processFile(
	path string,
	projectWorkingDir string,
	trackedFiles map[string]bool,
	clusttaIgnore *ignore.GitIgnore,
	filesChan chan<- string) {

	if strings.HasPrefix(filepath.Base(path), ".") {
		return
	}
	relativePath, _ := filepath.Rel(projectWorkingDir, path)
	relativePath = utils.NormalizePath(relativePath)
	if !trackedFiles[path] && !clusttaIgnore.MatchesPath(relativePath) {
		// ext := filepath.Ext(path)
		// relativePathWithoutExt := strings.TrimSuffix(relativePath, ext)
		filesChan <- relativePath
	}
}

func (p *ProjectService) IsIgnored(
	itemPath string,
	ignoreList []string) bool {
	ignoreObject := ignore.CompileIgnoreLines(ignoreList...)
	return ignoreObject.MatchesPath(itemPath)
}

// ValidateProjectFile checks if a .clst file is a valid Clustta project.
// Returns true if the file is valid, false otherwise.
func (p *ProjectService) ValidateProjectFile(filePath string) (bool, error) {
	if !utils.FileExists(filePath) {
		return false, errors.New("file not found")
	}
	return repository.VerifyProjectIntegrity(filePath)
}

// ClusttaFileInfo contains lightweight metadata extracted from a .clst file.
// Used when opening a project file from an arbitrary location.
type ClusttaFileInfo struct {
	Id                    string `json:"id"`
	Name                  string `json:"name"`
	StudioName            string `json:"studio_name"`
	Icon                  string `json:"icon"`
	WorkingDirectory      string `json:"working_directory"`
	LocalWorkingDirectory string `json:"local_working_directory"`
	FilePath              string `json:"file_path"`
	Valid                 bool   `json:"valid"`
}

// InspectClusttaFile opens a .clst file and extracts its metadata.
// If the file is outside Clustta's known project directories, updates the
// working directory in the database to a sibling folder of the .clst file.
func (p *ProjectService) InspectClusttaFile(filePath string) (ClusttaFileInfo, error) {
	if !utils.FileExists(filePath) {
		return ClusttaFileInfo{}, errors.New("file not found")
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return ClusttaFileInfo{}, err
	}

	valid, err := repository.VerifyProjectIntegrity(absPath)
	if err != nil || !valid {
		return ClusttaFileInfo{Valid: false, FilePath: absPath}, errors.New("invalid project file")
	}

	dbConn, err := utils.OpenDb(absPath)
	if err != nil {
		return ClusttaFileInfo{Valid: false, FilePath: absPath}, err
	}
	defer dbConn.Close()

	tx, err := dbConn.Beginx()
	if err != nil {
		return ClusttaFileInfo{Valid: false, FilePath: absPath}, err
	}
	defer tx.Rollback()

	projectId, err := utils.GetProjectId(tx)
	if err != nil {
		return ClusttaFileInfo{Valid: false, FilePath: absPath}, err
	}

	projectName, err := utils.GetProjectName(tx)
	if err != nil {
		return ClusttaFileInfo{Valid: false, FilePath: absPath}, err
	}

	studioName, err := utils.GetStudioName(tx)
	if err != nil {
		studioName = "Personal"
	}

	icon, err := utils.GetProjectIcon(tx)
	if err != nil {
		icon = ""
	}

	workingDir, err := utils.GetProjectWorkingDir(tx)
	if err != nil {
		workingDir = ""
	}

	// If the file is outside Clustta's known project directories,
	// update the working directory in the database to a sibling folder of the .clst file.
	localWorkingDir := ""
	if !isInProjectsDirectory(absPath) {
		localWorkingDir = strings.TrimSuffix(absPath, filepath.Ext(absPath))
		err = utils.SetProjectWorkingDir(tx, filepath.ToSlash(localWorkingDir))
		if err != nil {
			return ClusttaFileInfo{Valid: false, FilePath: absPath}, err
		}
		err = tx.Commit()
		if err != nil {
			return ClusttaFileInfo{Valid: false, FilePath: absPath}, err
		}
		workingDir = localWorkingDir
	}

	return ClusttaFileInfo{
		Id:                    projectId,
		Name:                  projectName,
		StudioName:            studioName,
		Icon:                  icon,
		WorkingDirectory:      workingDir,
		LocalWorkingDirectory: localWorkingDir,
		FilePath:              absPath,
		Valid:                 true,
	}, nil
}

// isInProjectsDirectory checks if a file path is inside Clustta's known project directories.
func isInProjectsDirectory(absPath string) bool {
	projectsDir, err := settings.GetProjectDirectory()
	if err == nil && projectsDir != "" {
		normed, _ := filepath.Abs(projectsDir)
		if strings.HasPrefix(strings.ToLower(absPath), strings.ToLower(normed)+string(filepath.Separator)) {
			return true
		}
	}

	sharedDir, err := settings.GetSharedProjectDirectory()
	if err == nil && sharedDir != "" {
		normed, _ := filepath.Abs(sharedDir)
		if strings.HasPrefix(strings.ToLower(absPath), strings.ToLower(normed)+string(filepath.Separator)) {
			return true
		}
	}

	return false
}

// UploadProject uploads a local .clst project to a remote studio.
// It creates the project on the remote, copies the file, remaps IDs, and prepares for sync.
func (p *ProjectService) UploadProject(sourceClstPath, studioName, workingDir, projectName, remoteProjectUrl string) (repository.ProjectInfo, error) {
	if studioName == "" {
		return repository.ProjectInfo{}, errors.New("studio name can't be empty")
	}
	if studioName == "Personal" {
		return repository.ProjectInfo{}, errors.New("cannot upload to Personal studio")
	}

	user, err := auth_service.GetActiveUser()
	if err != nil {
		return repository.ProjectInfo{}, err
	}

	// Validate source file
	valid, err := repository.VerifyProjectIntegrity(sourceClstPath)
	if err != nil || !valid {
		return repository.ProjectInfo{}, errors.New("invalid project file")
	}

	// Create empty project on remote
	remoteProjectInfo, err := repository.CreateProject(remoteProjectUrl, studioName, workingDir, "", "", user)
	if err != nil {
		return repository.ProjectInfo{}, err
	}

	// Determine destination path for the .clst file
	sharedProjectsDir, err := settings.GetSharedProjectDirectory()
	if err != nil {
		return repository.ProjectInfo{}, err
	}
	studioProjectsDir := filepath.Join(sharedProjectsDir, studioName)
	os.MkdirAll(studioProjectsDir, os.ModePerm)
	destClstPath := filepath.Join(studioProjectsDir, projectName+".clst")

	// Copy the .clst file
	err = utils.CopyFile(sourceClstPath, destClstPath)
	if err != nil {
		return repository.ProjectInfo{}, errors.New("failed to copy project file: " + err.Error())
	}

	// Prepare the project for upload (remap IDs, update config)
	err = sync_service.PrepareProjectForUpload(destClstPath, remoteProjectInfo, remoteProjectUrl, workingDir, user.Id)
	if err != nil {
		// Clean up copied file on error
		os.Remove(destClstPath)
		return repository.ProjectInfo{}, errors.New("failed to prepare project for upload: " + err.Error())
	}

	// Get updated project info
	projectInfo, err := repository.GetProjectInfo(destClstPath, user)
	if err != nil {
		return repository.ProjectInfo{}, err
	}

	projectInfo.Uri = destClstPath
	projectInfo.Remote = remoteProjectUrl
	projectInfo.IsDownloaded = true
	projectInfo.WorkingDirectory = workingDir

	return projectInfo, nil
}

// GetWriteThroughEnabled returns whether write-through sync is enabled for the project.
func (p *ProjectService) GetWriteThroughEnabled(projectPath string) (bool, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return false, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	return utils.GetWriteThroughEnabled(tx)
}

// SetWriteThroughEnabled enables or disables write-through sync for the project.
func (p *ProjectService) SetWriteThroughEnabled(projectPath string, enabled bool) error {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}

	err = utils.SetWriteThroughEnabled(tx, enabled)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	InvalidateEnabledCache(projectPath)
	return nil
}
