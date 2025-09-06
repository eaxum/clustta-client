package services

import (
	"clustta/internal/auth_service"
	"clustta/internal/error_service"
	"clustta/internal/ignore"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/sync_service"
	"clustta/internal/settings"
	"clustta/internal/utils"
	"clustta/output"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type ProjectService struct {
}

type UntrackedItems struct {
	Files   []models.UntrackedTask   `json:"tasks"`
	Folders []models.UntrackedEntity `json:"entities"`
}

func (p *ProjectService) CreateProject(projectUri, studioName, workingDir, templateName string) (repository.ProjectInfo, error) {
	if studioName == "" {
		return repository.ProjectInfo{}, errors.New("studio name can't be empty")
	}
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return repository.ProjectInfo{}, err
	}
	projectInfo, err := repository.CreateProject(projectUri, studioName, workingDir, templateName, user)
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
	app.EmitEvent("progress-update", progress)

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
	app.EmitEvent("progress-update", progress)
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
	app.EmitEvent("progress-update", progress)
	return nil
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
	err = tx.Commit()
	if err != nil {
		return err
	}
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
		if errors.Is(err, error_service.ErrUserHaveTaskAssigned) {
			return error_service.ErrUserHaveTaskAssigned
		}
		return err
	}
	tx.Commit()
	return nil
}

func (p *ProjectService) GetStudioProjects(url, name string) ([]repository.ProjectInfo, error) {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		return []repository.ProjectInfo{}, err
	}

	projects, err := sync_service.GetStudioProjects(user, url, name)
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
		Files:   make([]models.UntrackedTask, 0),
		Folders: make([]models.UntrackedEntity, 0),
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
				entity := models.UntrackedEntity{
					Id:         utils.GetMD5Hash(path),
					Name:       entry.Name(),
					FilePath:   path,
					EntityPath: relativePath,
				}
				untracked.Folders = append(untracked.Folders, entity)
			}
		} else {
			// check if file is tracked
			if !utils.Contains(tracked, path) {
				taskName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				task := models.UntrackedTask{
					Id:         utils.GetMD5Hash(path),
					Name:       taskName,
					FilePath:   path,
					TaskPath:   relativePath,
					EntityPath: filepath.ToSlash(filepath.Dir(relativePath)),
					Extension:  filepath.Ext(entry.Name()),
				}
				untracked.Files = append(untracked.Files, task)
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
