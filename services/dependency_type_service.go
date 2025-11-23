package services

import (
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
)

type DependencyTypeService struct{}

//GetDependencyTypes retrieves all dependency types in the project.
//Returns the list of dependency types or an error if the operation fails.
func (d *DependencyTypeService) GetDependencyTypes(projectPath string) ([]models.DependencyType, error) {
	dbConn, err := utils.OpenDb(projectPath)
	if err != nil {
		return []models.DependencyType{}, err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return []models.DependencyType{}, err
	}
	defer tx.Rollback()

	dependency_types, err := repository.GetDependencyTypes(tx)
	if err != nil {
		return dependency_types, err
	}

	return dependency_types, nil
}
