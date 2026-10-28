package services

import (
	"log"
	"path"
	"path/filepath"
	"strings"

	"clustta/internal/project_index"
	"clustta/internal/repository"
	"clustta/internal/settings"
	"clustta/internal/utils"
)

func registerProjectInIndex(project repository.ProjectInfo, studioName string) {
	registerProjectsInIndex([]repository.ProjectInfo{project}, studioName)
}

func localProjectDatabasePath(projectURI, studioName string) (string, error) {
	if !utils.IsValidURL(projectURI) {
		return projectURI, nil
	}
	projectDirectory, err := settings.GetSharedProjectDirectory()
	if studioName == "Personal" {
		projectDirectory, err = settings.GetProjectDirectory()
	}
	if err != nil {
		return "", err
	}
	if studioName != "Personal" {
		projectDirectory = filepath.Join(projectDirectory, studioName)
	}
	return filepath.Join(projectDirectory, path.Base(projectURI)+".clst"), nil
}

func registerProjectsInIndex(projects []repository.ProjectInfo, studioName string) {
	entries := make([]projectindex.Entry, 0, len(projects))
	for _, project := range projects {
		if utils.IsValidURL(project.Uri) || strings.TrimSpace(project.WorkingDirectory) == "" {
			continue
		}
		entries = append(entries, projectindex.Entry{
			ProjectID:        project.Id,
			ProjectName:      project.Name,
			StudioName:       studioName,
			DatabasePath:     project.Uri,
			WorkingDirectory: project.WorkingDirectory,
		})
	}
	if len(entries) == 0 {
		return
	}
	if err := projectindex.UpsertMany(entries); err != nil {
		log.Printf("Failed to update project index for %s: %v", studioName, err)
	}
}
