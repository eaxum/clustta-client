// testutils/setup.go
package testutils

import (
	"clustta/internal/auth_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/utils"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var TestFolder = getAbsolutePath("../../tests")
var TestFile = getAbsolutePath("../testutils/place_holder.blend")
var TestFileModified = getAbsolutePath("../testutils/place_holder_modified.blend")
var TestPreviewFile = getAbsolutePath("../testutils/image.png")

var DbConn *sqlx.DB
var Tx *sqlx.Tx

var Project repository.ProjectInfo
var Template models.Template
var DependencyType models.DependencyType
var CollectionType models.CollectionType
var Collection models.Collection
var AssetType models.AssetType
var Asset models.Asset

func getAbsolutePath(file_path string) string {
	file, err := filepath.Abs(file_path)
	if err != nil {
		panic(err)
	}
	return file
}

func GenerateFixtureProject() {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		panic(err)
	}
	projectUri := filepath.Join(TestFolder, "test.clst")
	projectInfo, err := repository.CreateProject(projectUri, "test", "", "No Template", "", user)
	if err != nil {
		panic(err)
	}
	Project = projectInfo
}
func GenerateFixtureTemplate() {
	template, err := repository.CreateTemplate(Tx, "Basic", TestFile)
	if err != nil {
		panic(err)
	}
	Template = template
}
func GenerateFixtureCollectionType() {
	collectionType, err := repository.CreateCollectionType(Tx, "", "Test Collection Type", "test")
	if err != nil {
		panic(err)
	}
	CollectionType = collectionType
}
func GenerateFixtureCollection() {
	collection, err := repository.CreateCollection(Tx, "", "Test Collection", "", CollectionType.Id, "", "", true)
	if err != nil {
		panic(err)
	}
	Collection = collection
}
func GenerateFixtureAssetType() {
	assetType, err := repository.CreateAssetType(Tx, "", "Test Asset Type", "test")
	if err != nil {
		panic(err)
	}
	AssetType = assetType
}
func GenerateFixtureAsset() {
	user, err := auth_service.GetActiveUser()
	if err != nil {
		panic(err)
	}
	asset, err := repository.CreateAsset(Tx, "", "Test Asset", AssetType.Id, Collection.Id, false, Template.Id, "", "", []string{}, "", false, "", user.Id, "Asset created", uuid.New().String(), func(int, int, string, string) {})
	if err != nil {
		panic(err)
	}
	Asset = asset
}
func GenerateFixtureDependencyType() {
	dependencyType, err := repository.CreateDependencyType(Tx, "", "Test")
	if err != nil {
		panic(err)
	}
	DependencyType = dependencyType
}

func Setup() {
	err := os.RemoveAll(TestFolder)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(TestFolder, os.ModePerm)
	if err != nil {
		panic(err)
	}
	GenerateFixtureProject()
	DbConn, err = utils.OpenDb(Project.Uri)
	if err != nil {
		panic(err)
	}
	Tx, err = DbConn.Beginx()
	if err != nil {
		panic(err)
	}
}

func Teardown() {
	Tx.Rollback()
	DbConn.Close()

	err := os.RemoveAll(TestFolder)
	if err != nil {
		panic(err)
	}
}
