package sync_service

import (
	"clustta/internal/base_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/repositorypb"
	"clustta/internal/utils"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

func OLDLoadUserData(tx *sqlx.Tx, userId string) (ProjectData, error) {
	userData := ProjectData{}

	user, err := repository.GetUser(tx, userId)
	if err != nil {
		return ProjectData{}, err
	}
	userRole, err := repository.GetRole(tx, user.RoleId)
	if err != nil {
		return ProjectData{}, err
	}

	// assets, err := repository.GetUserAssets(tx, userId)
	// if err != nil {
	// 	return ProjectData{}, err
	// }
	assets := []models.Asset{}
	if userRole.ViewAsset {
		assets, err = repository.GetAssets(tx, true)
		if err != nil {
			return ProjectData{}, err
		}
	} else {
		assets, err = repository.OLDGetUserAssets(tx, user.Id)
		if err != nil {
			return ProjectData{}, err
		}
	}

	var assetCollectionIds []string
	var assetIds []string
	for _, asset := range assets {
		assetIds = append(assetIds, asset.Id)
		if !utils.Contains(assetCollectionIds, asset.CollectionId) {
			assetCollectionIds = append(assetCollectionIds, asset.CollectionId)
		}
	}
	quotedAssetIds := make([]string, len(assetIds))
	for i, id := range assetIds {
		quotedAssetIds[i] = fmt.Sprintf("\"%s\"", id)
	}
	quotedAssetCollectionIds := make([]string, len(assetCollectionIds))
	for i, id := range assetCollectionIds {
		quotedAssetCollectionIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	// dependenciesQuery := fmt.Sprintf("SELECT * FROM asset_dependencies WHERE asset_id IN (%s) AND dependency_id NOT IN (%s)", strings.Join(quotedAssetIds, ","), strings.Join(quotedAssetIds, ","))
	assetDependenciesQuery := fmt.Sprintf("SELECT * FROM asset_dependency WHERE asset_id IN (%s)", strings.Join(quotedAssetIds, ","))
	assetDependencies := []models.AssetDependency{}
	err = tx.Select(&assetDependencies, assetDependenciesQuery)
	if err != nil {
		return ProjectData{}, err
	}

	collectionDependenciesQuery := fmt.Sprintf("SELECT * FROM collection_dependency WHERE asset_id IN (%s)", strings.Join(quotedAssetIds, ","))
	collectionDependencies := []models.CollectionDependency{}
	err = tx.Select(&collectionDependencies, collectionDependenciesQuery)
	if err != nil {
		return ProjectData{}, err
	}

	var uniqueDependencyIds []string
	for _, dependency := range assetDependencies {
		if !utils.Contains(uniqueDependencyIds, dependency.DependencyId) && !utils.Contains(assetIds, dependency.DependencyId) {
			uniqueDependencyIds = append(uniqueDependencyIds, dependency.DependencyId)
		}
	}
	quotedUniqueDependencyIds := make([]string, len(uniqueDependencyIds))
	for i, id := range uniqueDependencyIds {
		quotedUniqueDependencyIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	uniqueDependenciesQuery := fmt.Sprintf("SELECT * FROM asset WHERE trashed = 0 AND id IN (%s)", strings.Join(quotedUniqueDependencyIds, ","))
	uniqueDependencies := []models.Asset{}
	err = tx.Select(&uniqueDependencies, uniqueDependenciesQuery)
	if err != nil {
		return ProjectData{}, err
	}

	assets = append(assets, uniqueDependencies...)
	assetIds = append(assetIds, uniqueDependencyIds...)

	quotedAssetIds = append(quotedAssetIds, quotedUniqueDependencyIds...)

	checkpointQuery := fmt.Sprintf("SELECT * FROM asset_checkpoint WHERE trashed = 0 AND asset_id IN (%s)", strings.Join(quotedAssetIds, ","))
	assetsCheckpoints := []models.Checkpoint{}
	err = tx.Select(&assetsCheckpoints, checkpointQuery)
	if err != nil {
		return ProjectData{}, err
	}

	statuses, err := repository.GetStatuses(tx)
	if err != nil {
		return ProjectData{}, err
	}
	assetTypes, err := repository.GetAssetTypes(tx)
	if err != nil {
		return ProjectData{}, err
	}

	users, err := repository.GetUsers(tx)
	if err != nil {
		return ProjectData{}, err
	}
	roles, err := repository.GetRoles(tx)
	if err != nil {
		return ProjectData{}, err
	}

	dependencyTypes, err := repository.GetDependencyTypes(tx)
	if err != nil {
		return ProjectData{}, err
	}

	collectionTypes, err := repository.GetCollectionTypes(tx)
	if err != nil {
		return ProjectData{}, err
	}
	collections := []models.Collection{}
	collectionAssignees := []models.CollectionAssignee{}
	if userRole.ViewAsset {
		//TODO remove with trashed
		collections, err = repository.GetCollections(tx, true)
		if err != nil {
			return ProjectData{}, err
		}
		err = tx.Select(&collectionAssignees, "SELECT * FROM collection_assignee")
		if err != nil {
			return ProjectData{}, err
		}
	} else {
		collections, err = repository.OLDGetUserCollections(tx, user.Id)
		if err != nil {
			return ProjectData{}, err
		}
		qoutedCollectionIds := make([]string, len(collections))
		for i, collection := range collections {
			qoutedCollectionIds[i] = fmt.Sprintf("\"%s\"", collection.Id)
		}
		collectionAssigneesQuery := fmt.Sprintf("SELECT * FROM collection_assignee WHERE collection_id IN (%s)", strings.Join(qoutedCollectionIds, ","))
		err = tx.Select(&collectionAssignees, collectionAssigneesQuery)
		if err != nil {
			return ProjectData{}, err
		}
	}

	templates := []models.Template{}
	if userRole.CreateAsset {
		templates, err = repository.GetTemplates(tx, false)
		if err != nil {
			return ProjectData{}, err
		}
	}

	workflows := []models.Workflow{}
	if userRole.CreateAsset {
		workflows, err = repository.GetWorkflows(tx)
		if err != nil {
			return ProjectData{}, err
		}
	}
	workflowLinks := []models.WorkflowLink{}
	if userRole.CreateAsset {
		err = base_service.GetAll(tx, "workflow_link", &workflowLinks)
		if err != nil {
			return ProjectData{}, err
		}
	}
	workflowCollections := []models.WorkflowCollection{}
	if userRole.CreateAsset {
		err = base_service.GetAll(tx, "workflow_collection", &workflowCollections)
		if err != nil {
			return ProjectData{}, err
		}
	}
	workflowAssets := []models.WorkflowAsset{}
	if userRole.CreateAsset {
		err = base_service.GetAll(tx, "workflow_asset", &workflowAssets)
		if err != nil {
			return ProjectData{}, err
		}
	}

	tags, err := repository.GetTags(tx)
	if err != nil {
		return ProjectData{}, err
	}

	assetstagsQuery := fmt.Sprintf("SELECT * FROM asset_tag WHERE asset_id IN (%s)", strings.Join(quotedAssetIds, ","))
	assetsTags := []models.AssetTag{}
	err = tx.Select(&assetsTags, assetstagsQuery)
	if err != nil {
		return ProjectData{}, err
	}

	projectPreview, err := repository.GetProjectPreview(tx)
	if err != nil {
		if err.Error() != "no preview" {
			return ProjectData{}, err
		}
	}

	userData.ProjectPreview = projectPreview.Hash
	userData.CollectionTypes = collectionTypes
	userData.Collections = collections
	userData.CollectionAssignees = collectionAssignees

	userData.AssetTypes = assetTypes
	userData.Assets = assets
	userData.AssetCheckpoints = assetsCheckpoints
	userData.AssetDependencies = assetDependencies
	userData.CollectionDependencies = collectionDependencies

	userData.Statuses = statuses
	userData.DependencyTypes = dependencyTypes

	userData.Users = users
	userData.Roles = roles

	userData.Templates = templates

	userData.Workflows = workflows
	userData.WorkflowLinks = workflowLinks
	userData.WorkflowCollections = workflowCollections
	userData.WorkflowAssets = workflowAssets

	userData.Tags = tags
	userData.AssetTags = assetsTags
	return userData, nil
}

func LoadUserData(tx *sqlx.Tx, userId string) (ProjectData, error) {
	userData := ProjectData{}

	user, err := repository.GetUser(tx, userId)
	if err != nil {
		return ProjectData{}, err
	}
	userRole, err := repository.GetRole(tx, user.RoleId)
	if err != nil {
		return ProjectData{}, err
	}

	// assets, err := repository.GetUserAssets(tx, userId)
	// if err != nil {
	// 	return ProjectData{}, err
	// }
	assets := []models.Asset{}
	if userRole.ViewAsset {
		assets, err = repository.GetAssets(tx, true)
		if err != nil {
			return ProjectData{}, err
		}
	} else {
		assets, err = repository.GetUserAssets(tx, user.Id)
		if err != nil {
			return ProjectData{}, err
		}
	}

	var assetCollectionIds []string
	var assetIds []string
	for _, asset := range assets {
		assetIds = append(assetIds, asset.Id)
		if !utils.Contains(assetCollectionIds, asset.CollectionId) {
			assetCollectionIds = append(assetCollectionIds, asset.CollectionId)
		}
	}
	quotedAssetIds := make([]string, len(assetIds))
	for i, id := range assetIds {
		quotedAssetIds[i] = fmt.Sprintf("\"%s\"", id)
	}
	quotedAssetCollectionIds := make([]string, len(assetCollectionIds))
	for i, id := range assetCollectionIds {
		quotedAssetCollectionIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	// dependenciesQuery := fmt.Sprintf("SELECT * FROM asset_dependencies WHERE asset_id IN (%s) AND dependency_id NOT IN (%s)", strings.Join(quotedAssetIds, ","), strings.Join(quotedAssetIds, ","))
	assetDependenciesQuery := fmt.Sprintf("SELECT * FROM asset_dependency WHERE asset_id IN (%s)", strings.Join(quotedAssetIds, ","))
	assetDependencies := []models.AssetDependency{}
	err = tx.Select(&assetDependencies, assetDependenciesQuery)
	if err != nil {
		return ProjectData{}, err
	}

	collectionDependenciesQuery := fmt.Sprintf("SELECT * FROM collection_dependency WHERE asset_id IN (%s)", strings.Join(quotedAssetIds, ","))
	collectionDependencies := []models.CollectionDependency{}
	err = tx.Select(&collectionDependencies, collectionDependenciesQuery)
	if err != nil {
		return ProjectData{}, err
	}

	var uniqueDependencyIds []string
	for _, dependency := range assetDependencies {
		if !utils.Contains(uniqueDependencyIds, dependency.DependencyId) && !utils.Contains(assetIds, dependency.DependencyId) {
			uniqueDependencyIds = append(uniqueDependencyIds, dependency.DependencyId)
		}
	}
	quotedUniqueDependencyIds := make([]string, len(uniqueDependencyIds))
	for i, id := range uniqueDependencyIds {
		quotedUniqueDependencyIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	uniqueDependenciesQuery := fmt.Sprintf("SELECT * FROM asset WHERE trashed = 0 AND id IN (%s)", strings.Join(quotedUniqueDependencyIds, ","))
	uniqueDependencies := []models.Asset{}
	err = tx.Select(&uniqueDependencies, uniqueDependenciesQuery)
	if err != nil {
		return ProjectData{}, err
	}

	assets = append(assets, uniqueDependencies...)
	assetIds = append(assetIds, uniqueDependencyIds...)

	quotedAssetIds = append(quotedAssetIds, quotedUniqueDependencyIds...)

	checkpointQuery := fmt.Sprintf("SELECT * FROM asset_checkpoint WHERE trashed = 0 AND asset_id IN (%s)", strings.Join(quotedAssetIds, ","))
	assetsCheckpoints := []models.Checkpoint{}
	err = tx.Select(&assetsCheckpoints, checkpointQuery)
	if err != nil {
		return ProjectData{}, err
	}

	statuses, err := repository.GetStatuses(tx)
	if err != nil {
		return ProjectData{}, err
	}
	assetTypes, err := repository.GetAssetTypes(tx)
	if err != nil {
		return ProjectData{}, err
	}

	users, err := repository.GetUsers(tx)
	if err != nil {
		return ProjectData{}, err
	}
	roles, err := repository.GetRoles(tx)
	if err != nil {
		return ProjectData{}, err
	}

	dependencyTypes, err := repository.GetDependencyTypes(tx)
	if err != nil {
		return ProjectData{}, err
	}

	collectionTypes, err := repository.GetCollectionTypes(tx)
	if err != nil {
		return ProjectData{}, err
	}

	collections := []models.Collection{}
	collectionAssignees := []models.CollectionAssignee{}
	if userRole.ViewAsset {
		//TODO remove with trashed
		collections, err = repository.GetCollections(tx, true)
		if err != nil {
			return ProjectData{}, err
		}
		err = tx.Select(&collectionAssignees, "SELECT * FROM collection_assignee")
		if err != nil {
			return ProjectData{}, err
		}
	} else {
		collections, err = repository.GetUserCollections(tx, assets, user.Id)
		if err != nil {
			return ProjectData{}, err
		}

		qoutedCollectionIds := make([]string, len(collections))
		for i, collection := range collections {
			qoutedCollectionIds[i] = fmt.Sprintf("\"%s\"", collection.Id)
		}
		collectionAssigneesQuery := fmt.Sprintf("SELECT * FROM collection_assignee WHERE collection_id IN (%s)", strings.Join(qoutedCollectionIds, ","))
		err = tx.Select(&collectionAssignees, collectionAssigneesQuery)
		if err != nil {
			return ProjectData{}, err
		}
	}

	templates := []models.Template{}
	if userRole.CreateAsset {
		templates, err = repository.GetTemplates(tx, false)
		if err != nil {
			return ProjectData{}, err
		}
	}

	workflows := []models.Workflow{}
	if userRole.CreateAsset {
		workflows, err = repository.GetWorkflows(tx)
		if err != nil {
			return ProjectData{}, err
		}
	}
	workflowLinks := []models.WorkflowLink{}
	if userRole.CreateAsset {
		err = base_service.GetAll(tx, "workflow_link", &workflowLinks)
		if err != nil {
			return ProjectData{}, err
		}
	}
	workflowCollections := []models.WorkflowCollection{}
	if userRole.CreateAsset {
		err = base_service.GetAll(tx, "workflow_collection", &workflowCollections)
		if err != nil {
			return ProjectData{}, err
		}
	}
	workflowAssets := []models.WorkflowAsset{}
	if userRole.CreateAsset {
		err = base_service.GetAll(tx, "workflow_asset", &workflowAssets)
		if err != nil {
			return ProjectData{}, err
		}
	}

	tags, err := repository.GetTags(tx)
	if err != nil {
		return ProjectData{}, err
	}

	assetstagsQuery := fmt.Sprintf("SELECT * FROM asset_tag WHERE asset_id IN (%s)", strings.Join(quotedAssetIds, ","))
	assetsTags := []models.AssetTag{}
	err = tx.Select(&assetsTags, assetstagsQuery)
	if err != nil {
		return ProjectData{}, err
	}

	projectPreview, err := repository.GetProjectPreview(tx)
	if err != nil {
		if err.Error() != "no preview" {
			return ProjectData{}, err
		}
	}

	userData.ProjectPreview = projectPreview.Hash
	userData.CollectionTypes = collectionTypes
	userData.Collections = collections
	userData.CollectionAssignees = collectionAssignees

	userData.AssetTypes = assetTypes
	userData.Assets = assets
	userData.AssetCheckpoints = assetsCheckpoints
	userData.AssetDependencies = assetDependencies
	userData.CollectionDependencies = collectionDependencies

	userData.Statuses = statuses
	userData.DependencyTypes = dependencyTypes

	userData.Users = users
	userData.Roles = roles

	userData.Templates = templates

	userData.Workflows = workflows
	userData.WorkflowLinks = workflowLinks
	userData.WorkflowCollections = workflowCollections
	userData.WorkflowAssets = workflowAssets

	userData.Tags = tags
	userData.AssetTags = assetsTags
	return userData, nil
}

func LoadUserDataPb(tx *sqlx.Tx, userId string) ([]byte, error) {
	user, err := repository.GetUser(tx, userId)
	if err != nil {
		return []byte{}, err
	}
	userRole, err := repository.GetRole(tx, user.RoleId)
	if err != nil {
		return []byte{}, err
	}

	// assets, err := repository.GetUserAssets(tx, userId)
	// if err != nil {
	// 	return ProjectData{}, err
	// }
	assets := []models.Asset{}
	if userRole.ViewAsset {
		assets, err = repository.GetAssets(tx, true)
		if err != nil {
			return []byte{}, err
		}
	} else {
		assets, err = repository.GetUserAssets(tx, user.Id)
		if err != nil {
			return []byte{}, err
		}
	}

	var assetCollectionIds []string
	var assetIds []string
	for _, asset := range assets {
		assetIds = append(assetIds, asset.Id)
		if !utils.Contains(assetCollectionIds, asset.CollectionId) {
			assetCollectionIds = append(assetCollectionIds, asset.CollectionId)
		}
	}
	quotedAssetIds := make([]string, len(assetIds))
	for i, id := range assetIds {
		quotedAssetIds[i] = fmt.Sprintf("\"%s\"", id)
	}
	quotedAssetCollectionIds := make([]string, len(assetCollectionIds))
	for i, id := range assetCollectionIds {
		quotedAssetCollectionIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	// dependenciesQuery := fmt.Sprintf("SELECT * FROM asset_dependencies WHERE asset_id IN (%s) AND dependency_id NOT IN (%s)", strings.Join(quotedAssetIds, ","), strings.Join(quotedAssetIds, ","))
	assetDependenciesQuery := fmt.Sprintf("SELECT * FROM asset_dependency WHERE asset_id IN (%s)", strings.Join(quotedAssetIds, ","))
	assetDependencies := []models.AssetDependency{}
	err = tx.Select(&assetDependencies, assetDependenciesQuery)
	if err != nil {
		return []byte{}, err
	}

	collectionDependenciesQuery := fmt.Sprintf("SELECT * FROM collection_dependency WHERE asset_id IN (%s)", strings.Join(quotedAssetIds, ","))
	collectionDependencies := []models.CollectionDependency{}
	err = tx.Select(&collectionDependencies, collectionDependenciesQuery)
	if err != nil {
		return []byte{}, err
	}

	var uniqueDependencyIds []string
	for _, dependency := range assetDependencies {
		if !utils.Contains(uniqueDependencyIds, dependency.DependencyId) && !utils.Contains(assetIds, dependency.DependencyId) {
			uniqueDependencyIds = append(uniqueDependencyIds, dependency.DependencyId)
		}
	}
	quotedUniqueDependencyIds := make([]string, len(uniqueDependencyIds))
	for i, id := range uniqueDependencyIds {
		quotedUniqueDependencyIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	uniqueDependenciesQuery := fmt.Sprintf("SELECT * FROM asset WHERE trashed = 0 AND id IN (%s)", strings.Join(quotedUniqueDependencyIds, ","))
	uniqueDependencies := []models.Asset{}
	err = tx.Select(&uniqueDependencies, uniqueDependenciesQuery)
	if err != nil {
		return []byte{}, err
	}

	assets = append(assets, uniqueDependencies...)
	assetIds = append(assetIds, uniqueDependencyIds...)

	quotedAssetIds = append(quotedAssetIds, quotedUniqueDependencyIds...)

	checkpointQuery := fmt.Sprintf("SELECT * FROM asset_checkpoint WHERE trashed = 0 AND asset_id IN (%s)", strings.Join(quotedAssetIds, ","))
	assetsCheckpoints := []models.Checkpoint{}
	err = tx.Select(&assetsCheckpoints, checkpointQuery)
	if err != nil {
		return []byte{}, err
	}

	statuses, err := repository.GetStatuses(tx)
	if err != nil {
		return []byte{}, err
	}
	assetTypes, err := repository.GetAssetTypes(tx)
	if err != nil {
		return []byte{}, err
	}

	users, err := repository.GetUsers(tx)
	if err != nil {
		return []byte{}, err
	}
	roles, err := repository.GetRoles(tx)
	if err != nil {
		return []byte{}, err
	}

	dependencyTypes, err := repository.GetDependencyTypes(tx)
	if err != nil {
		return []byte{}, err
	}

	collectionTypes, err := repository.GetCollectionTypes(tx)
	if err != nil {
		return []byte{}, err
	}

	collections := []models.Collection{}
	collectionAssignees := []models.CollectionAssignee{}
	if userRole.ViewAsset {
		//TODO remove with trashed
		collections, err = repository.GetCollections(tx, true)
		if err != nil {
			return []byte{}, err
		}
		err = tx.Select(&collectionAssignees, "SELECT * FROM collection_assignee")
		if err != nil {
			return []byte{}, err
		}
	} else {
		collections, err = repository.GetUserCollections(tx, assets, user.Id)
		if err != nil {
			return []byte{}, err
		}

		qoutedCollectionIds := make([]string, len(collections))
		for i, collection := range collections {
			qoutedCollectionIds[i] = fmt.Sprintf("\"%s\"", collection.Id)
		}
		collectionAssigneesQuery := fmt.Sprintf("SELECT * FROM collection_assignee WHERE collection_id IN (%s)", strings.Join(qoutedCollectionIds, ","))
		err = tx.Select(&collectionAssignees, collectionAssigneesQuery)
		if err != nil {
			return []byte{}, err
		}
	}

	templates := []models.Template{}
	if userRole.CreateAsset {
		templates, err = repository.GetTemplates(tx, false)
		if err != nil {
			return []byte{}, err
		}
	}

	workflows := []models.Workflow{}
	if userRole.CreateAsset {
		workflows, err = repository.GetWorkflows(tx)
		if err != nil {
			return []byte{}, err
		}
	}
	workflowLinks := []models.WorkflowLink{}
	if userRole.CreateAsset {
		err = base_service.GetAll(tx, "workflow_link", &workflowLinks)
		if err != nil {
			return []byte{}, err
		}
	}
	workflowCollections := []models.WorkflowCollection{}
	if userRole.CreateAsset {
		err = base_service.GetAll(tx, "workflow_collection", &workflowCollections)
		if err != nil {
			return []byte{}, err
		}
	}
	workflowAssets := []models.WorkflowAsset{}
	if userRole.CreateAsset {
		err = base_service.GetAll(tx, "workflow_asset", &workflowAssets)
		if err != nil {
			return []byte{}, err
		}
	}

	tags, err := repository.GetTags(tx)
	if err != nil {
		return []byte{}, err
	}

	assetstagsQuery := fmt.Sprintf("SELECT * FROM asset_tag WHERE asset_id IN (%s)", strings.Join(quotedAssetIds, ","))
	assetsTags := []models.AssetTag{}
	err = tx.Select(&assetsTags, assetstagsQuery)
	if err != nil {
		return []byte{}, err
	}

	projectPreview, err := repository.GetProjectPreview(tx)
	if err != nil {
		if err.Error() != "no preview" {
			return []byte{}, err
		}
	}

	// Load integration data
	integrationProjects, err := repository.GetIntegrationProjects(tx)
	if err != nil {
		return []byte{}, err
	}

	integrationCollectionMappings, err := repository.GetAllCollectionMappings(tx)
	if err != nil {
		return []byte{}, err
	}

	integrationAssetMappings, err := repository.GetAllAssetMappings(tx)
	if err != nil {
		return []byte{}, err
	}

	userData := &repositorypb.ProjectData{
		ProjectPreview:  projectPreview.Hash,
		CollectionTypes:     repository.ToPbCollectionTypes(collectionTypes),
		Collections:        repository.ToPbCollections(collections),
		CollectionAssignees: repository.ToPbCollectionAssignees(collectionAssignees),

		AssetTypes:          repository.ToPbAssetTypes(assetTypes),
		Assets:              repository.ToPbAssets(assets),
		AssetCheckpoints:   repository.ToPbCheckpoints(assetsCheckpoints),
		AssetDependencies:   repository.ToPbAssetDependencies(assetDependencies),
		CollectionDependencies: repository.ToPbCollectionDependencies(collectionDependencies),

		Statuses:        repository.ToPbStatuses(statuses),
		DependencyTypes: repository.ToPbDependencyTypes(dependencyTypes),

		Users: repository.ToPbUsers(users),
		Roles: repository.ToPbRoles(roles),

		Templates: repository.ToPbTemplates(templates),

		Workflows:        repository.ToPbWorkflows(workflows),
		WorkflowLinks:    repository.ToPbWorkflowLinks(workflowLinks),
		WorkflowCollections: repository.ToPbWorkflowCollections(workflowCollections),
		WorkflowAssets:    repository.ToPbWorkflowAssets(workflowAssets),

		Tags:      repository.ToPbTags(tags),
		AssetTags: repository.ToPbAssetTags(assetsTags),

		IntegrationProjects:           repository.ToPbIntegrationProjects(integrationProjects),
		IntegrationCollectionMappings: repository.ToPbIntegrationCollectionMappings(integrationCollectionMappings),
		IntegrationAssetMappings:      repository.ToPbIntegrationAssetMappings(integrationAssetMappings),
	}
	userDataBytes, err := proto.Marshal(userData)
	if err != nil {
		return []byte{}, err
	}

	return userDataBytes, nil

	// userData.ProjectPreview = projectPreview.Hash
	// userData.CollectionTypes = repository.ToPbCollectionTypes(collectionTypes)
	// userData.Collections = repository.ToPbCollections(collections)
	// userData.CollectionAssignees = repository.ToPbCollectionAssignees(collectionAssignees)

	// userData.AssetTypes = repository.ToPbAssetTypes(assetTypes)
	// userData.Assets = repository.ToPbAssets(assets)
	// userData.AssetCheckpoints = repository.ToPbCheckpoints(assetsCheckpoints)
	// userData.AssetDependencies = repository.ToPbAssetDependencies(assetDependencies)
	// userData.CollectionDependencies = repository.ToPbCollectionDependencies(collectionDependencies)

	// userData.Statuses = repository.ToPbStatuses(statuses)
	// userData.DependencyTypes = repository.ToPbDependencyTypes(dependencyTypes)

	// userData.Users = repository.ToPbUsers(users)
	// userData.Roles = repository.ToPbRoles(roles)

	// userData.Templates = repository.ToPbTemplates(templates)

	// userData.Workflows = repository.ToPbWorkflows(workflows)
	// userData.WorkflowLinks = repository.ToPbWorkflowLinks(workflowLinks)
	// userData.WorkflowCollections = repository.ToPbWorkflowCollections(workflowCollections)
	// userData.WorkflowAssets = repository.ToPbWorkflowAssets(workflowAssets)

	// userData.Tags = repository.ToPbTags(tags)
	// userData.AssetTags = repository.ToPbAssetTags(assetsTags)
	// return userData, nil
}

func LoadChangedData(tx *sqlx.Tx) (ProjectData, error) {
	userData := ProjectData{}

	assetQuery := "SELECT * FROM asset WHERE synced = 0"
	assets := []models.Asset{}
	err := tx.Select(&assets, assetQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	checkpointQuery := "SELECT * FROM asset_checkpoint WHERE synced = 0"
	assetsCheckpoints := []models.Checkpoint{}
	err = tx.Select(&assetsCheckpoints, checkpointQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	assetDependenciesQuery := "SELECT * FROM asset_dependency WHERE synced = 0"
	assetDependencies := []models.AssetDependency{}
	err = tx.Select(&assetDependencies, assetDependenciesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	collectionDependenciesQuery := "SELECT * FROM collection_dependency WHERE synced = 0"
	collectionDependencies := []models.CollectionDependency{}
	err = tx.Select(&collectionDependencies, collectionDependenciesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	statusesQuery := "SELECT * FROM status WHERE synced = 0"
	statuses := []models.Status{}
	err = tx.Select(&statuses, statusesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	dependencyTypesQuery := "SELECT * FROM dependency_type WHERE synced = 0"
	dependencyTypes := []models.DependencyType{}
	err = tx.Select(&dependencyTypes, dependencyTypesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	assetTypeQuery := "SELECT * FROM asset_type WHERE synced = 0"
	assetTypes := []models.AssetType{}
	err = tx.Select(&assetTypes, assetTypeQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	collectionTypesQuery := "SELECT * FROM collection_type WHERE synced = 0"
	collectionTypes := []models.CollectionType{}
	err = tx.Select(&collectionTypes, collectionTypesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	collectionQuery := "SELECT * FROM collection WHERE synced = 0"
	collections := []models.Collection{}
	err = tx.Select(&collections, collectionQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	collectionAssigneeQuery := "SELECT * FROM collection_assignee WHERE synced = 0"
	collectionAssignees := []models.CollectionAssignee{}
	err = tx.Select(&collectionAssignees, collectionAssigneeQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	userQuery := "SELECT * FROM user WHERE synced = 0"
	users := []models.User{}
	err = tx.Select(&users, userQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	roleQuery := "SELECT * FROM role WHERE synced = 0"
	roles := []models.Role{}
	err = tx.Select(&roles, roleQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	templatesQuery := "SELECT * FROM template WHERE synced = 0"
	templates := []models.Template{}
	err = tx.Select(&templates, templatesQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	workflowsQuery := "SELECT * FROM workflow WHERE synced = 0"
	workflows := []models.Workflow{}
	err = tx.Select(&workflows, workflowsQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	workflowLinksQuery := "SELECT * FROM workflow_link WHERE synced = 0"
	workflowLinks := []models.WorkflowLink{}
	err = tx.Select(&workflowLinks, workflowLinksQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	workflowCollectionsQuery := "SELECT * FROM workflow_collection WHERE synced = 0"
	workflowCollections := []models.WorkflowCollection{}
	err = tx.Select(&workflowCollections, workflowCollectionsQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	workflowAssetsQuery := "SELECT * FROM workflow_asset WHERE synced = 0"
	workflowAssets := []models.WorkflowAsset{}
	err = tx.Select(&workflowAssets, workflowAssetsQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	tagsQuery := "SELECT * FROM tag WHERE synced = 0"
	tags := []models.Tag{}
	err = tx.Select(&tags, tagsQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	assetsTagsQuery := "SELECT * FROM asset_tag WHERE synced = 0"
	assetsTags := []models.AssetTag{}
	err = tx.Select(&assetsTags, assetsTagsQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	tombs, err := repository.GetTombs(tx)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	isProjectPreviewSynced, err := repository.IsProjectPreviewSynced(tx)
	if err != nil {
		return userData, err
	}
	if !isProjectPreviewSynced {
		projectPreview, err := repository.GetProjectPreview(tx)
		if err != nil {
			return ProjectData{}, err
		}
		userData.ProjectPreview = projectPreview.Hash
	}
	userData.AssetTypes = assetTypes
	userData.Assets = assets
	userData.AssetCheckpoints = assetsCheckpoints
	userData.AssetDependencies = assetDependencies
	userData.CollectionDependencies = collectionDependencies

	userData.Statuses = statuses
	userData.DependencyTypes = dependencyTypes

	userData.Users = users
	userData.Roles = roles

	userData.CollectionTypes = collectionTypes
	userData.Collections = collections
	userData.CollectionAssignees = collectionAssignees

	userData.Templates = templates

	userData.Workflows = workflows
	userData.WorkflowLinks = workflowLinks
	userData.WorkflowCollections = workflowCollections
	userData.WorkflowAssets = workflowAssets

	userData.Tags = tags
	userData.AssetTags = assetsTags

	userData.Tombs = tombs

	// Integration tables
	integrationProjectsQuery := "SELECT * FROM integration_project WHERE synced = 0"
	integrationProjects := []models.IntegrationProject{}
	err = tx.Select(&integrationProjects, integrationProjectsQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	userData.IntegrationProjects = integrationProjects

	integrationCollectionMappingsQuery := "SELECT * FROM integration_collection_mapping WHERE synced = 0"
	integrationCollectionMappings := []models.IntegrationCollectionMapping{}
	err = tx.Select(&integrationCollectionMappings, integrationCollectionMappingsQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	userData.IntegrationCollectionMappings = integrationCollectionMappings

	integrationAssetMappingsQuery := "SELECT * FROM integration_asset_mapping WHERE synced = 0"
	integrationAssetMappings := []models.IntegrationAssetMapping{}
	err = tx.Select(&integrationAssetMappings, integrationAssetMappingsQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}
	userData.IntegrationAssetMappings = integrationAssetMappings

	return userData, nil
}

// LoadAssetData loads a single asset (asset) and its unsynced checkpoints from the database.
// Only includes rows where synced = 0. Does not include dependencies.
func LoadAssetData(tx *sqlx.Tx, assetId string) (ProjectData, error) {
	data := ProjectData{}

	asset := models.Asset{}
	err := tx.Get(&asset, "SELECT * FROM asset WHERE id = ? AND synced = 0", assetId)
	if err != nil && err != sql.ErrNoRows {
		return data, err
	}
	if err == nil {
		data.Assets = []models.Asset{asset}
	}

	checkpoints := []models.Checkpoint{}
	err = tx.Select(&checkpoints, "SELECT * FROM asset_checkpoint WHERE asset_id = ? AND synced = 0", assetId)
	if err != nil && err != sql.ErrNoRows {
		return data, err
	}
	data.AssetCheckpoints = checkpoints

	return data, nil
}

func LoadChangedDataPb(tx *sqlx.Tx) ([]byte, error) {

	assetQuery := "SELECT * FROM asset WHERE synced = 0"
	assets := []models.Asset{}
	err := tx.Select(&assets, assetQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	checkpointQuery := "SELECT * FROM asset_checkpoint WHERE synced = 0"
	assetsCheckpoints := []models.Checkpoint{}
	err = tx.Select(&assetsCheckpoints, checkpointQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	assetDependenciesQuery := "SELECT * FROM asset_dependency WHERE synced = 0"
	assetDependencies := []models.AssetDependency{}
	err = tx.Select(&assetDependencies, assetDependenciesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	collectionDependenciesQuery := "SELECT * FROM collection_dependency WHERE synced = 0"
	collectionDependencies := []models.CollectionDependency{}
	err = tx.Select(&collectionDependencies, collectionDependenciesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	statusesQuery := "SELECT * FROM status WHERE synced = 0"
	statuses := []models.Status{}
	err = tx.Select(&statuses, statusesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	dependencyTypesQuery := "SELECT * FROM dependency_type WHERE synced = 0"
	dependencyTypes := []models.DependencyType{}
	err = tx.Select(&dependencyTypes, dependencyTypesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	assetTypeQuery := "SELECT * FROM asset_type WHERE synced = 0"
	assetTypes := []models.AssetType{}
	err = tx.Select(&assetTypes, assetTypeQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	collectionTypesQuery := "SELECT * FROM collection_type WHERE synced = 0"
	collectionTypes := []models.CollectionType{}
	err = tx.Select(&collectionTypes, collectionTypesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	collectionQuery := "SELECT * FROM collection WHERE synced = 0"
	collections := []models.Collection{}
	err = tx.Select(&collections, collectionQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	collectionAssigneeQuery := "SELECT * FROM collection_assignee WHERE synced = 0"
	collectionAssignees := []models.CollectionAssignee{}
	err = tx.Select(&collectionAssignees, collectionAssigneeQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	userQuery := "SELECT * FROM user WHERE synced = 0"
	users := []models.User{}
	err = tx.Select(&users, userQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	roleQuery := "SELECT * FROM role WHERE synced = 0"
	roles := []models.Role{}
	err = tx.Select(&roles, roleQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	templatesQuery := "SELECT * FROM template WHERE synced = 0"
	templates := []models.Template{}
	err = tx.Select(&templates, templatesQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	workflowsQuery := "SELECT * FROM workflow WHERE synced = 0"
	workflows := []models.Workflow{}
	err = tx.Select(&workflows, workflowsQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}
	workflowLinksQuery := "SELECT * FROM workflow_link WHERE synced = 0"
	workflowLinks := []models.WorkflowLink{}
	err = tx.Select(&workflowLinks, workflowLinksQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}
	workflowCollectionsQuery := "SELECT * FROM workflow_collection WHERE synced = 0"
	workflowCollections := []models.WorkflowCollection{}
	err = tx.Select(&workflowCollections, workflowCollectionsQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}
	workflowAssetsQuery := "SELECT * FROM workflow_asset WHERE synced = 0"
	workflowAssets := []models.WorkflowAsset{}
	err = tx.Select(&workflowAssets, workflowAssetsQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	tagsQuery := "SELECT * FROM tag WHERE synced = 0"
	tags := []models.Tag{}
	err = tx.Select(&tags, tagsQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}
	assetsTagsQuery := "SELECT * FROM asset_tag WHERE synced = 0"
	assetsTags := []models.AssetTag{}
	err = tx.Select(&assetsTags, assetsTagsQuery)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}

	tombs, err := repository.GetTombs(tx)
	if err != nil && err != sql.ErrNoRows {
		return []byte{}, err
	}
	isProjectPreviewSynced, err := repository.IsProjectPreviewSynced(tx)
	if err != nil {
		return []byte{}, err
	}
	projectPreview := models.Preview{}
	if !isProjectPreviewSynced {
		projectPreview, err = repository.GetProjectPreview(tx)
		if err != nil {
			return []byte{}, err
		}
	}

	userData := &repositorypb.ProjectData{
		ProjectPreview:  projectPreview.Hash,
		CollectionTypes:     repository.ToPbCollectionTypes(collectionTypes),
		Collections:        repository.ToPbCollections(collections),
		CollectionAssignees: repository.ToPbCollectionAssignees(collectionAssignees),

		AssetTypes:          repository.ToPbAssetTypes(assetTypes),
		Assets:              repository.ToPbAssets(assets),
		AssetCheckpoints:   repository.ToPbCheckpoints(assetsCheckpoints),
		AssetDependencies:   repository.ToPbAssetDependencies(assetDependencies),
		CollectionDependencies: repository.ToPbCollectionDependencies(collectionDependencies),

		Statuses:        repository.ToPbStatuses(statuses),
		DependencyTypes: repository.ToPbDependencyTypes(dependencyTypes),

		Users: repository.ToPbUsers(users),
		Roles: repository.ToPbRoles(roles),

		Templates: repository.ToPbTemplates(templates),

		Workflows:        repository.ToPbWorkflows(workflows),
		WorkflowLinks:    repository.ToPbWorkflowLinks(workflowLinks),
		WorkflowCollections: repository.ToPbWorkflowCollections(workflowCollections),
		WorkflowAssets:    repository.ToPbWorkflowAssets(workflowAssets),

		Tags:      repository.ToPbTags(tags),
		AssetTags: repository.ToPbAssetTags(assetsTags),

		Tomb: repository.ToPbTombs(tombs),
	}
	userDataBytes, err := proto.Marshal(userData)
	if err != nil {
		return []byte{}, err
	}
	return userDataBytes, nil
}

func LoadCheckpointData(tx *sqlx.Tx) (ProjectData, error) {
	userData := ProjectData{}

	assetCheckpointQuery := "SELECT * FROM asset_checkpoint WHERE synced = 0"
	assetsCheckpoints := []models.Checkpoint{}
	err := tx.Select(&assetsCheckpoints, assetCheckpointQuery)
	if err != nil && err != sql.ErrNoRows {
		return userData, err
	}

	userData.AssetCheckpoints = assetsCheckpoints
	return userData, nil
}
