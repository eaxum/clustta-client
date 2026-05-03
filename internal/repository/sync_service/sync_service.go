package sync_service

import (
	"bytes"
	"clustta/internal/auth_service"
	"clustta/internal/chunk_service"
	"clustta/internal/constants"
	"clustta/internal/error_service"
	"clustta/internal/repository"
	"clustta/internal/repository/models"
	"clustta/internal/repository/repositorypb"
	"clustta/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	kzstd "github.com/klauspost/compress/zstd"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

type ProjectData struct {
	ProjectPreview         string                        `json:"project_preview"`
	Assets                 []models.Asset                `json:"assets"`
	AssetTypes             []models.AssetType            `json:"asset_types"`
	AssetCheckpoints       []models.Checkpoint           `json:"assets_checkpoints"`
	AssetDependencies      []models.AssetDependency      `json:"asset_dependencies"`
	CollectionDependencies []models.CollectionDependency `json:"collection_dependencies"`

	Statuses        []models.Status         `json:"statuses"`
	DependencyTypes []models.DependencyType `json:"dependency_types"`

	Users []models.User `json:"users"`
	Roles []models.Role `json:"roles"`

	CollectionTypes     []models.CollectionType     `json:"collection_types"`
	Collections         []models.Collection         `json:"collections"`
	CollectionAssignees []models.CollectionAssignee `json:"collection_assignees"`

	Templates []models.Template `json:"templates"`
	Tags      []models.Tag      `json:"tags"`
	AssetTags []models.AssetTag `json:"assets_tags"`

	Workflows           []models.Workflow           `json:"workflows"`
	WorkflowLinks       []models.WorkflowLink       `json:"workflow_links"`
	WorkflowCollections []models.WorkflowCollection `json:"workflow_collections"`
	WorkflowAssets      []models.WorkflowAsset      `json:"workflow_assets"`

	Tombs []repository.Tomb `json:"tomb"`

	IntegrationProjects           []models.IntegrationProject           `json:"integration_projects"`
	IntegrationCollectionMappings []models.IntegrationCollectionMapping `json:"integration_collection_mappings"`
	IntegrationAssetMappings      []models.IntegrationAssetMapping      `json:"integration_asset_mappings"`
}

func (d *ProjectData) IsEmpty() bool {
	return len(d.Assets) == 0 &&
		len(d.AssetTypes) == 0 &&
		len(d.AssetCheckpoints) == 0 &&
		len(d.AssetDependencies) == 0 &&
		len(d.CollectionDependencies) == 0 &&
		len(d.CollectionTypes) == 0 &&
		len(d.Collections) == 0 &&
		len(d.CollectionAssignees) == 0 &&
		len(d.Templates) == 0 &&
		len(d.Tags) == 0 &&
		len(d.AssetTags) == 0 &&
		len(d.Statuses) == 0 &&
		len(d.DependencyTypes) == 0 &&
		len(d.Users) == 0 &&
		len(d.Roles) == 0 &&
		len(d.Workflows) == 0 &&
		len(d.WorkflowLinks) == 0 &&
		len(d.WorkflowCollections) == 0 &&
		len(d.WorkflowAssets) == 0 &&
		len(d.Tombs) == 0 &&
		len(d.IntegrationProjects) == 0 &&
		len(d.IntegrationCollectionMappings) == 0 &&
		len(d.IntegrationAssetMappings) == 0 &&
		d.ProjectPreview == ""
}

func WriteProjectData(tx *sqlx.Tx, data ProjectData, strict bool) error {

	// Sort
	sortedCollections, err := repository.TopologicalSort(data.Collections)
	if err != nil {
		return err
	}
	data.Collections = sortedCollections

	tombItems := make(map[string]bool)
	tombedItems, err := repository.GetTombedItems(tx)
	if err != nil {
		return err
	}
	for _, tombItem := range tombedItems {
		tombItems[tombItem] = true
	}

	chunks := []string{}
	for _, AssetCheckpoint := range data.AssetCheckpoints {
		chunksString := AssetCheckpoint.Chunks
		chunkHashes := strings.Split(chunksString, ",")
		for _, chunkHash := range chunkHashes {
			if !utils.Contains(chunks, chunkHash) {
				chunks = append(chunks, chunkHash)
			}
		}
	}
	for _, Template := range data.Templates {
		chunksString := Template.Chunks
		chunkHashes := strings.Split(chunksString, ",")
		for _, chunkHash := range chunkHashes {
			if !utils.Contains(chunks, chunkHash) {
				chunks = append(chunks, chunkHash)
			}
		}
	}
	if strict {
		missingChunks, err := chunk_service.GetNonExistingChunks(tx, chunks)
		if err != nil {
			return err
		}
		if len(missingChunks) != 0 {
			return errors.New("data have missing chunks")
		}
	}

	previewIds := []string{}
	if data.ProjectPreview != "" && !utils.Contains(previewIds, data.ProjectPreview) {
		previewIds = append(previewIds, data.ProjectPreview)
	}

	for _, asset := range data.Assets {
		if asset.PreviewId != "" && !utils.Contains(previewIds, asset.PreviewId) {
			previewIds = append(previewIds, asset.PreviewId)
		}
	}
	for _, collection := range data.Collections {
		if collection.PreviewId != "" && !utils.Contains(previewIds, collection.PreviewId) {
			previewIds = append(previewIds, collection.PreviewId)
		}
	}
	for _, assetCheckpoint := range data.AssetCheckpoints {
		if assetCheckpoint.PreviewId != "" && !utils.Contains(previewIds, assetCheckpoint.PreviewId) {
			previewIds = append(previewIds, assetCheckpoint.PreviewId)
		}
	}

	missingPreviews, err := repository.GetNonExistingPreviews(tx, previewIds)
	if err != nil {
		return err
	}
	if len(missingPreviews) != 0 {
		return errors.New("data have missing previews")
	}

	if data.ProjectPreview != "" {
		_, err = tx.Exec(`
		INSERT INTO config (name, value, mtime, synced)
		VALUES ('project_preview', $1, $2, 1)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime, synced = 1
	`, data.ProjectPreview, utils.GetEpochTime())
	}

	for _, role := range data.Roles {
		if tombItems[role.Id] {
			continue
		}
		roleAttributes := models.RoleAttributes{
			ViewCollection:   role.ViewCollection,
			CreateCollection: role.CreateCollection,
			UpdateCollection: role.UpdateCollection,
			DeleteCollection: role.DeleteCollection,

			ViewAsset:   role.ViewAsset,
			CreateAsset: role.CreateAsset,
			UpdateAsset: role.UpdateAsset,
			DeleteAsset: role.DeleteAsset,

			ViewTemplate:   role.ViewTemplate,
			CreateTemplate: role.CreateTemplate,
			UpdateTemplate: role.UpdateTemplate,
			DeleteTemplate: role.DeleteTemplate,

			ViewCheckpoint:   role.ViewCheckpoint,
			CreateCheckpoint: role.CreateCheckpoint,
			DeleteCheckpoint: role.DeleteCheckpoint,

			PullChunk: role.PullChunk,

			AssignAsset:   role.AssignAsset,
			UnassignAsset: role.UnassignAsset,

			AddUser:    role.AddUser,
			RemoveUser: role.RemoveUser,
			ChangeRole: role.ChangeRole,

			ChangeStatus:   role.ChangeStatus,
			SetDoneAsset:   role.SetDoneAsset,
			SetRetakeAsset: role.SetRetakeAsset,

			ViewDoneAsset: role.ViewDoneAsset,

			ManageDependencies: role.ManageDependencies,
			ManageShareLinks:   role.ManageShareLinks,
		}
		localRole, err := repository.GetRole(tx, role.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrRoleNotFound) {
				_, err := repository.CreateRole(tx, role.Id, role.Name, roleAttributes)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localRole.MTime < role.MTime {
				_, err = repository.UpdateRole(tx, role.Id, role.Name, roleAttributes)
				if err != nil {
					return err
				}
			}
		}

	}
	for _, user := range data.Users {
		// if tombItems[user.Id] {
		// 	continue
		// }

		localUser, err := repository.GetUser(tx, user.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrUserNotFound) {
				_, err := repository.AddKnownUser(
					tx, user.Id, user.Email, user.Username,
					user.FirstName, user.LastName, user.RoleId, user.Photo, false)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localUser.MTime < user.MTime {
				if localUser.RoleId != user.RoleId {
					err = repository.ChangeUserRole(tx, user.Id, user.RoleId)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	for _, collectionType := range data.CollectionTypes {
		if tombItems[collectionType.Id] {
			continue
		}
		localCollectionType, err := repository.GetCollectionType(tx, collectionType.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrCollectionTypeNotFound) {
				// Check if a local type with the same name or icon already exists
				existing, nameErr := repository.GetCollectionTypeByName(tx, collectionType.Name)
				if nameErr == nil {
					if existing.MTime < collectionType.MTime {
						_, err = repository.UpdateCollectionType(tx, existing.Id, collectionType.Name, collectionType.Icon)
						if err != nil {
							return err
						}
					}
					continue
				}
				existing, iconErr := repository.GetCollectionTypeByIcon(tx, collectionType.Icon)
				if iconErr == nil {
					if existing.MTime < collectionType.MTime {
						_, err = repository.UpdateCollectionType(tx, existing.Id, collectionType.Name, collectionType.Icon)
						if err != nil {
							return err
						}
					}
					continue
				}
				_, err = repository.CreateCollectionType(
					tx, collectionType.Id, collectionType.Name, collectionType.Icon)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localCollectionType.MTime < collectionType.MTime {
				_, err = repository.UpdateCollectionType(tx, collectionType.Id, collectionType.Name, collectionType.Icon)
				if err != nil {
					return err
				}
			}
		}

	}
	for _, assetType := range data.AssetTypes {
		if tombItems[assetType.Id] {
			continue
		}
		localAssetType, err := repository.GetAssetType(tx, assetType.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrAssetTypeNotFound) {
				// Check if a local type with the same name or icon already exists
				existing, nameErr := repository.GetAssetTypeByName(tx, assetType.Name)
				if nameErr == nil {
					if existing.MTime < assetType.MTime {
						_, err = repository.UpdateAssetType(tx, existing.Id, assetType.Name, assetType.Icon)
						if err != nil {
							return err
						}
					}
					continue
				}
				existing, iconErr := repository.GetAssetTypeByIcon(tx, assetType.Icon)
				if iconErr == nil {
					if existing.MTime < assetType.MTime {
						_, err = repository.UpdateAssetType(tx, existing.Id, assetType.Name, assetType.Icon)
						if err != nil {
							return err
						}
					}
					continue
				}
				_, err = repository.CreateAssetType(
					tx, assetType.Id, assetType.Name, assetType.Icon)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {

			if localAssetType.MTime < assetType.MTime {
				_, err = repository.UpdateAssetType(tx, assetType.Id, assetType.Name, assetType.Icon)
				if err != nil {
					return err
				}
			}
		}

	}
	for _, dependencyType := range data.DependencyTypes {
		if tombItems[dependencyType.Id] {
			continue
		}
		_, err = repository.GetDependencyType(tx, dependencyType.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrDependencyTypeNotFound) {
				_, err = repository.CreateDependencyType(
					tx, dependencyType.Id, dependencyType.Name)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	for _, status := range data.Statuses {
		if tombItems[status.Id] {
			continue
		}
		_, err = repository.GetStatus(tx, status.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrStatusNotFound) {
				_, err = repository.CreateStatus(
					tx, status.Id, status.Name, status.ShortName, status.Color)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	for _, tag := range data.Tags {
		if tombItems[tag.Id] {
			continue
		}
		_, err = repository.GetTag(tx, tag.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrTagNotFound) {
				_, err = repository.CreateTag(tx, tag.Id, tag.Name)
				if err != nil {
					return err
				}
			} else {

				return err
			}
		}
	}

	localCollections, err := repository.GetSimpleCollections(tx)
	if err != nil {
		return err
	}
	localCollectionsIndex := make(map[string]int)
	for i, t := range localCollections {
		localCollectionsIndex[t.Id] = i
	}

	for _, collection := range data.Collections {
		if tombItems[collection.Id] {
			continue
		}
		i, exists := localCollectionsIndex[collection.Id]
		if !exists {
			fmt.Println("Creating: ", collection.Name)
			err = repository.AddCollection(
				tx, collection.Id, collection.Name, collection.Description, collection.CollectionTypeId, collection.ParentId, collection.PreviewId, collection.IsShared)
			if err != nil {
				if err.Error() == "parent collection not found" {
					continue
				}
				return err
			}
			continue
		}

		localCollection := localCollections[i]
		if localCollection.MTime < collection.MTime {

			parentId := collection.ParentId
			previewId := collection.PreviewId
			isShared := collection.IsShared

			collection, err = repository.RenameCollection(tx, collection.Id, collection.Name)
			if err != nil {
				return err
			}

			collection.ParentId = parentId
			collection.PreviewId = previewId
			collection.IsShared = isShared

			if localCollection.ParentId != collection.ParentId {
				err = repository.ChangeParent(tx, collection.Id, collection.ParentId)
				if err != nil {
					return err
				}
			}

			if localCollection.PreviewId != collection.PreviewId {
				err = repository.SetCollectionPreview(tx, collection.Id, "collection", collection.PreviewId)
				if err != nil {
					return err
				}
			}
			if localCollection.IsShared != collection.IsShared {
				err = repository.ChangeIsShared(tx, collection.Id, collection.IsShared)
				if err != nil {
					return err
				}
			}

		}
	}

	for _, collectionAssignee := range data.CollectionAssignees {
		if tombItems[collectionAssignee.Id] {
			continue
		}
		_, err = repository.GetAssignee(tx, collectionAssignee.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrCollectionAssigneeNotFound) {
				err = repository.AddAssignee(
					tx, collectionAssignee.Id, collectionAssignee.CollectionId, collectionAssignee.AssigneeId)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	localAssets, err := repository.GetSimpleAssets(tx)
	if err != nil {
		return err
	}
	localAssetsIndex := make(map[string]int)
	for i, t := range localAssets {
		localAssetsIndex[t.Id] = i
	}

	createAssetQuery := `
		INSERT INTO asset 
		(id, assignee_id, mtime, created_at, name, description, extension, asset_type_id, collection_id, is_resource, status_id, pointer, is_link, preview_id) 
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?);
	`
	createAssetStmt, err := tx.Prepare(createAssetQuery)
	if err != nil {
		return err
	}

	for _, asset := range data.Assets {
		if tombItems[asset.Id] {
			continue
		}

		i, exists := localAssetsIndex[asset.Id]
		if !exists {
			_, err := createAssetStmt.Exec(asset.Id, asset.AssigneeId, asset.MTime, asset.CreatedAt, asset.Name, asset.Description, asset.Extension, asset.AssetTypeId, asset.CollectionId, asset.IsResource, asset.StatusId, asset.Pointer, asset.IsLink, asset.PreviewId)
			if err != nil {
				return err
			}
			continue
		}

		localAsset := localAssets[i]
		if localAsset.MTime < asset.MTime {
			err := repository.UpdateSyncAsset(tx, asset.Id, asset.Name, asset.CollectionId, asset.AssetTypeId, asset.AssigneeId, asset.AssignerId, asset.StatusId, asset.PreviewId, asset.IsResource, asset.IsLink, asset.Pointer, []string{})
			if err != nil {
				return err
			}
		}
	}

	localAssetCheckpoints, err := repository.GetSimpleCheckpoints(tx)
	if err != nil {
		return err
	}
	localAssetCheckpointsIndex := make(map[string]int)
	for i, c := range localAssetCheckpoints {
		localAssetCheckpointsIndex[c.Id] = i
	}

	createCheckpointQuery := `
		INSERT INTO asset_checkpoint 
		(id, mtime, created_at, asset_id, xxhash_checksum, time_modified, file_size, comment, chunks, author_id, preview_id, group_id) 
		VALUES (?, ?,?,?,?,?,?,?,?,?,?,?);
	`
	createCheckpointStmt, err := tx.Prepare(createCheckpointQuery)
	if err != nil {
		return err
	}

	for _, assetCheckpoint := range data.AssetCheckpoints {
		if tombItems[assetCheckpoint.Id] {
			continue
		}

		_, exists := localAssetCheckpointsIndex[assetCheckpoint.Id]
		if !exists {
			EpochTime, err := utils.RFC3339ToEpoch(assetCheckpoint.CreatedAt)
			if err != nil {
				return err
			}

			_, err = createCheckpointStmt.Exec(assetCheckpoint.Id, assetCheckpoint.MTime, EpochTime, assetCheckpoint.AssetId, assetCheckpoint.XXHashChecksum, assetCheckpoint.TimeModified, assetCheckpoint.FileSize, assetCheckpoint.Comment, assetCheckpoint.Chunks, assetCheckpoint.AuthorUID, assetCheckpoint.PreviewId, assetCheckpoint.GroupId)
			if err != nil {
				return err
			}
			continue
		}
	}

	for _, dependency := range data.AssetDependencies {
		if tombItems[dependency.Id] {
			continue
		}
		_, err = repository.GetDependency(tx, dependency.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrAssetDependencyNotFound) {
				_, err = repository.AddDependency(
					tx, dependency.Id, dependency.AssetId, dependency.DependencyId, dependency.DependencyTypeId)
				if err != nil {
					if err.Error() == "UNIQUE constraint failed: asset_dependency.asset_id, asset_dependency.dependency_id" {
						continue
					}
					return err
				}
			} else {
				return err
			}
		}
	}

	for _, dependency := range data.CollectionDependencies {
		if tombItems[dependency.Id] {
			continue
		}
		_, err = repository.GetCollectionDependency(tx, dependency.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrCollectionDependencyNotFound) {
				_, err = repository.AddCollectionDependency(
					tx, dependency.Id, dependency.AssetId, dependency.DependencyId, dependency.DependencyTypeId)
				if err != nil {
					if err.Error() == "UNIQUE constraint failed: collection_dependency.asset_id, collection_dependency.dependency_id" {
						continue
					}
					return err
				}
			} else {
				return err
			}
		}
	}

	for _, template := range data.Templates {
		if tombItems[template.Id] {
			continue
		}
		_, err = repository.GetTemplate(tx, template.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrTemplateNotFound) {
				_, err = repository.AddTemplate(
					tx, template.Id, template.Name, template.Extension, template.Chunks, template.XxhashChecksum, template.FileSize)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	for _, workflow := range data.Workflows {
		if tombItems[workflow.Id] {
			continue
		}
		localWorkflow, err := repository.GetWorkflow(tx, workflow.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrWorkflowNotFound) {
				_, err = repository.CreateWorkflow(
					tx, workflow.Id, workflow.Name, []models.WorkflowAsset{}, []models.WorkflowCollection{}, []models.WorkflowLink{})
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localWorkflow.MTime < workflow.MTime {
				err = repository.RenameWorkflow(tx, workflow.Id, workflow.Name)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, workflowLink := range data.WorkflowLinks {
		if tombItems[workflowLink.Id] {
			continue
		}
		localWorkflowLink, err := repository.GetWorkflowLink(tx, workflowLink.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrWorkflowLinkNotFound) {
				err = repository.AddLinkWorkflow(tx, workflowLink.Id, workflowLink.Name, workflowLink.CollectionTypeId, workflowLink.WorkflowId, workflowLink.LinkedWorkflowId)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localWorkflowLink.MTime < workflowLink.MTime {
				err = repository.RenameLinkedWorkflow(tx, workflowLink.Id, workflowLink.Name)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, workflowCollection := range data.WorkflowCollections {
		if tombItems[workflowCollection.Id] {
			continue
		}
		localWorkflowCollection, err := repository.GetWorkflowCollection(tx, workflowCollection.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrWorkflowCollectionNotFound) {
				_, err = repository.CreateWorkflowCollection(tx, workflowCollection.Id, workflowCollection.Name, workflowCollection.WorkflowId, workflowCollection.CollectionTypeId)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localWorkflowCollection.MTime < workflowCollection.MTime {
				_, err = repository.UpdateWorkflowCollection(
					tx, workflowCollection.Id, workflowCollection.Name, workflowCollection.CollectionTypeId)
				if err != nil {
					return err
				}
			}
		}

	}

	for _, workflowAsset := range data.WorkflowAssets {
		if tombItems[workflowAsset.Id] {
			continue
		}
		localWorkflowAsset, err := repository.GetWorkflowAsset(tx, workflowAsset.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrWorkflowAssetNotFound) {
				_, err = repository.CreateWorkflowAsset(tx, workflowAsset.Id, workflowAsset.Name, workflowAsset.WorkflowId, workflowAsset.AssetTypeId, workflowAsset.IsResource, workflowAsset.TemplateId, workflowAsset.Pointer, workflowAsset.IsLink)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if localWorkflowAsset.MTime < workflowAsset.MTime {
				_, err = repository.UpdateWorkflowAsset(tx, workflowAsset.Id, workflowAsset.Name, workflowAsset.AssetTypeId, workflowAsset.IsResource, workflowAsset.TemplateId, workflowAsset.Pointer, workflowAsset.IsLink)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, assetTag := range data.AssetTags {
		if tombItems[assetTag.Id] {
			continue
		}
		_, err = repository.GetAssetTag(tx, assetTag.Id)
		if err != nil {
			if errors.Is(err, error_service.ErrAssetTagNotFound) {
				err = repository.AddTagToAssetById(tx, assetTag.Id, assetTag.AssetId, assetTag.TagId)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	// Integration project (only one per project)
	for _, integrationProject := range data.IntegrationProjects {
		if tombItems[integrationProject.Id] {
			continue
		}
		localIntegration, err := repository.GetIntegrationProject(tx)
		if err != nil {
			// No integration exists, create new
			_, err = repository.CreateIntegrationProject(
				tx, integrationProject.Id, integrationProject.IntegrationId, integrationProject.ExternalProjectId,
				integrationProject.ExternalProjectName, integrationProject.ApiUrl, integrationProject.SyncOptions,
				integrationProject.LinkedByUserId, integrationProject.LinkedAt)
			if err != nil {
				return err
			}
		} else {
			// Update if remote is newer
			if localIntegration.MTime < integrationProject.MTime {
				err = repository.UpdateIntegrationProject(tx, localIntegration.Id, map[string]interface{}{
					"integration_id":        integrationProject.IntegrationId,
					"external_project_id":   integrationProject.ExternalProjectId,
					"external_project_name": integrationProject.ExternalProjectName,
					"api_url":               integrationProject.ApiUrl,
					"sync_options":          integrationProject.SyncOptions,
					"enabled":               integrationProject.Enabled,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	// Integration collection mappings
	for _, mapping := range data.IntegrationCollectionMappings {
		if tombItems[mapping.Id] {
			continue
		}
		localMapping, err := repository.GetCollectionMapping(tx, mapping.Id)
		if err != nil {
			// Mapping doesn't exist, create new
			_, err = repository.CreateCollectionMapping(
				tx, mapping.Id, mapping.IntegrationId, mapping.ExternalId, mapping.ExternalType,
				mapping.ExternalName, mapping.ExternalParentId, mapping.ExternalPath,
				mapping.ExternalMetadata, mapping.CollectionId, mapping.SyncedAt)
			if err != nil {
				return err
			}
		} else {
			// Update if remote is newer
			if localMapping.MTime < mapping.MTime {
				err = repository.UpdateCollectionMapping(tx, mapping.Id, map[string]interface{}{
					"external_type":      mapping.ExternalType,
					"external_name":      mapping.ExternalName,
					"external_parent_id": mapping.ExternalParentId,
					"external_path":      mapping.ExternalPath,
					"external_metadata":  mapping.ExternalMetadata,
					"collection_id":      mapping.CollectionId,
					"synced_at":          mapping.SyncedAt,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	// Integration asset mappings
	for _, mapping := range data.IntegrationAssetMappings {
		if tombItems[mapping.Id] {
			continue
		}
		localMapping, err := repository.GetAssetMapping(tx, mapping.Id)
		if err != nil {
			// Mapping doesn't exist, create new
			_, err = repository.CreateAssetMapping(
				tx, mapping.Id, mapping.IntegrationId, mapping.ExternalId, mapping.ExternalName,
				mapping.ExternalParentId, mapping.ExternalType, mapping.ExternalStatus,
				mapping.ExternalAssignees, mapping.ExternalMetadata, mapping.AssetId, mapping.SyncedAt)
			if err != nil {
				return err
			}
		} else {
			// Update if remote is newer
			if localMapping.MTime < mapping.MTime {
				err = repository.UpdateAssetMapping(tx, mapping.Id, map[string]interface{}{
					"external_name":             mapping.ExternalName,
					"external_parent_id":        mapping.ExternalParentId,
					"external_type":             mapping.ExternalType,
					"external_status":           mapping.ExternalStatus,
					"external_assignees":        mapping.ExternalAssignees,
					"external_metadata":         mapping.ExternalMetadata,
					"asset_id":                  mapping.AssetId,
					"last_pushed_checkpoint_id": mapping.LastPushedCheckpointId,
					"synced_at":                 mapping.SyncedAt,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func OverWriteProjectData(tx *sqlx.Tx, data ProjectData) error {
	// Sort
	sortedCollections, err := repository.TopologicalSort(data.Collections)
	if err != nil {
		return err
	}
	data.Collections = sortedCollections

	previewIds := []string{}
	if data.ProjectPreview != "" && !utils.Contains(previewIds, data.ProjectPreview) {
		previewIds = append(previewIds, data.ProjectPreview)
	}

	for _, asset := range data.Assets {
		if asset.PreviewId != "" && !utils.Contains(previewIds, asset.PreviewId) {
			previewIds = append(previewIds, asset.PreviewId)
		}
	}
	for _, collection := range data.Collections {
		if collection.PreviewId != "" && !utils.Contains(previewIds, collection.PreviewId) {
			previewIds = append(previewIds, collection.PreviewId)
		}
	}
	for _, assetCheckpoint := range data.AssetCheckpoints {
		if assetCheckpoint.PreviewId != "" && !utils.Contains(previewIds, assetCheckpoint.PreviewId) {
			previewIds = append(previewIds, assetCheckpoint.PreviewId)
		}
	}

	missingPreviews, err := repository.GetNonExistingPreviews(tx, previewIds)
	if err != nil {
		return err
	}
	if len(missingPreviews) != 0 {
		return errors.New("data have missing previews")
	}

	if data.ProjectPreview != "" {
		_, err = tx.Exec(`
		INSERT INTO config (name, value, mtime, synced)
		VALUES ('project_preview', $1, $2, 1)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, mtime = EXCLUDED.mtime, synced = 1
	`, data.ProjectPreview, utils.GetEpochTime())
		if err != nil {
			return err
		}
	}

	for _, role := range data.Roles {
		roleAttributes := models.RoleAttributes{
			ViewCollection:   role.ViewCollection,
			CreateCollection: role.CreateCollection,
			UpdateCollection: role.UpdateCollection,
			DeleteCollection: role.DeleteCollection,

			ViewAsset:   role.ViewAsset,
			CreateAsset: role.CreateAsset,
			UpdateAsset: role.UpdateAsset,
			DeleteAsset: role.DeleteAsset,

			ViewTemplate:   role.ViewTemplate,
			CreateTemplate: role.CreateTemplate,
			UpdateTemplate: role.UpdateTemplate,
			DeleteTemplate: role.DeleteTemplate,

			ViewCheckpoint:   role.ViewCheckpoint,
			CreateCheckpoint: role.CreateCheckpoint,
			DeleteCheckpoint: role.DeleteCheckpoint,

			PullChunk: role.PullChunk,

			AssignAsset:   role.AssignAsset,
			UnassignAsset: role.UnassignAsset,

			AddUser:    role.AddUser,
			RemoveUser: role.RemoveUser,
			ChangeRole: role.ChangeRole,

			ChangeStatus:   role.ChangeStatus,
			SetDoneAsset:   role.SetDoneAsset,
			SetRetakeAsset: role.SetRetakeAsset,

			ViewDoneAsset: role.ViewDoneAsset,

			ManageDependencies: role.ManageDependencies,
			ManageShareLinks:   role.ManageShareLinks,
		}
		_, err := repository.CreateRole(tx, role.Id, role.Name, roleAttributes)
		if err != nil {
			return err
		}
	}

	for _, user := range data.Users {
		_, err := repository.AddKnownUser(
			tx, user.Id, user.Email, user.Username,
			user.FirstName, user.LastName, user.RoleId, user.Photo, false)
		if err != nil {
			return err
		}
	}

	for _, collectionType := range data.CollectionTypes {
		_, err = repository.CreateCollectionType(
			tx, collectionType.Id, collectionType.Name, collectionType.Icon)
		if err != nil {
			return err
		}
	}

	for _, assetType := range data.AssetTypes {
		_, err = repository.CreateAssetType(
			tx, assetType.Id, assetType.Name, assetType.Icon)
		if err != nil {
			return err
		}
	}

	for _, dependencyType := range data.DependencyTypes {
		_, err = repository.CreateDependencyType(
			tx, dependencyType.Id, dependencyType.Name)
		if err != nil {
			return err
		}
	}

	for _, status := range data.Statuses {
		_, err = repository.CreateStatus(
			tx, status.Id, status.Name, status.ShortName, status.Color)
		if err != nil {
			return err
		}
	}

	for _, tag := range data.Tags {
		_, err = repository.CreateTag(tx, tag.Id, tag.Name)
		if err != nil {
			return err
		}
	}

	for _, collection := range data.Collections {
		err = repository.AddCollection(
			tx, collection.Id, collection.Name, collection.Description, collection.CollectionTypeId, collection.ParentId, collection.PreviewId, collection.IsShared)
		if err != nil {
			if err.Error() == "parent collection not found" {
				continue
			}
			return err
		}
	}

	for _, collectionAssignee := range data.CollectionAssignees {
		err = repository.AddAssignee(
			tx, collectionAssignee.Id, collectionAssignee.CollectionId, collectionAssignee.AssigneeId)
		if err != nil {
			return err
		}
	}

	localAssets, err := repository.GetSimpleAssets(tx)
	if err != nil {
		return err
	}
	localAssetsIndex := make(map[string]int)
	for i, t := range localAssets {
		localAssetsIndex[t.Id] = i
	}

	createAssetQuery := `
		INSERT INTO asset 
		(id, assignee_id, mtime, created_at, name, description, extension, asset_type_id, collection_id, is_resource, status_id, pointer, is_link, preview_id) 
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?);
	`
	createAssetStmt, err := tx.Prepare(createAssetQuery)
	if err != nil {
		return err
	}

	for _, asset := range data.Assets {
		_, err := createAssetStmt.Exec(asset.Id, asset.AssigneeId, asset.MTime, asset.CreatedAt, asset.Name, asset.Description, asset.Extension, asset.AssetTypeId, asset.CollectionId, asset.IsResource, asset.StatusId, asset.Pointer, asset.IsLink, asset.PreviewId)
		if err != nil {
			return err
		}
	}

	createCheckpointQuery := `
		INSERT INTO asset_checkpoint 
		(id, mtime, created_at, asset_id, xxhash_checksum, time_modified, file_size, comment, chunks, author_id, preview_id, group_id) 
		VALUES (?, ?,?,?,?,?,?,?,?,?,?,?);
	`
	createCheckpointStmt, err := tx.Prepare(createCheckpointQuery)
	if err != nil {
		return err
	}

	for _, assetCheckpoint := range data.AssetCheckpoints {
		EpochTime, err := utils.RFC3339ToEpoch(assetCheckpoint.CreatedAt)
		if err != nil {
			return err
		}
		_, err = createCheckpointStmt.Exec(assetCheckpoint.Id, assetCheckpoint.MTime, EpochTime, assetCheckpoint.AssetId, assetCheckpoint.XXHashChecksum, assetCheckpoint.TimeModified, assetCheckpoint.FileSize, assetCheckpoint.Comment, assetCheckpoint.Chunks, assetCheckpoint.AuthorUID, assetCheckpoint.PreviewId, assetCheckpoint.GroupId)
		if err != nil {
			return err
		}
	}

	for _, dependency := range data.AssetDependencies {
		_, err = repository.AddDependency(
			tx, dependency.Id, dependency.AssetId, dependency.DependencyId, dependency.DependencyTypeId)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: asset_dependency.asset_id, asset_dependency.dependency_id" {
				continue
			}
			return err
		}
	}

	for _, dependency := range data.CollectionDependencies {
		_, err = repository.AddCollectionDependency(
			tx, dependency.Id, dependency.AssetId, dependency.DependencyId, dependency.DependencyTypeId)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: collection_dependency.asset_id, collection_dependency.dependency_id" {
				continue
			}
			return err
		}
	}

	for _, template := range data.Templates {
		_, err = repository.AddTemplate(
			tx, template.Id, template.Name, template.Extension, template.Chunks, template.XxhashChecksum, template.FileSize)
		if err != nil {
			return err
		}
	}

	for _, workflow := range data.Workflows {
		_, err = repository.CreateWorkflow(
			tx, workflow.Id, workflow.Name, []models.WorkflowAsset{}, []models.WorkflowCollection{}, []models.WorkflowLink{})
		if err != nil {
			return err
		}
	}

	for _, workflowLink := range data.WorkflowLinks {
		err = repository.AddLinkWorkflow(tx, workflowLink.Id, workflowLink.Name, workflowLink.CollectionTypeId, workflowLink.WorkflowId, workflowLink.LinkedWorkflowId)
		if err != nil {
			return err
		}
	}

	for _, workflowCollection := range data.WorkflowCollections {
		_, err = repository.CreateWorkflowCollection(tx, workflowCollection.Id, workflowCollection.Name, workflowCollection.WorkflowId, workflowCollection.CollectionTypeId)
		if err != nil {
			return err
		}
	}

	for _, workflowAsset := range data.WorkflowAssets {
		_, err = repository.CreateWorkflowAsset(tx, workflowAsset.Id, workflowAsset.Name, workflowAsset.WorkflowId, workflowAsset.AssetTypeId, workflowAsset.IsResource, workflowAsset.TemplateId, workflowAsset.Pointer, workflowAsset.IsLink)
		if err != nil {
			return err
		}
	}

	for _, assetTag := range data.AssetTags {
		err = repository.AddTagToAssetById(tx, assetTag.Id, assetTag.AssetId, assetTag.TagId)
		if err != nil {
			return err
		}
	}

	// Integration project
	for _, integrationProject := range data.IntegrationProjects {
		_, err = repository.CreateIntegrationProject(
			tx, integrationProject.Id, integrationProject.IntegrationId, integrationProject.ExternalProjectId,
			integrationProject.ExternalProjectName, integrationProject.ApiUrl, integrationProject.SyncOptions,
			integrationProject.LinkedByUserId, integrationProject.LinkedAt)
		if err != nil {
			return err
		}
	}

	// Integration collection mappings
	for _, mapping := range data.IntegrationCollectionMappings {
		_, err = repository.CreateCollectionMapping(
			tx, mapping.Id, mapping.IntegrationId, mapping.ExternalId, mapping.ExternalType,
			mapping.ExternalName, mapping.ExternalParentId, mapping.ExternalPath,
			mapping.ExternalMetadata, mapping.CollectionId, mapping.SyncedAt)
		if err != nil {
			return err
		}
	}

	// Integration asset mappings
	for _, mapping := range data.IntegrationAssetMappings {
		_, err = repository.CreateAssetMapping(
			tx, mapping.Id, mapping.IntegrationId, mapping.ExternalId, mapping.ExternalName,
			mapping.ExternalParentId, mapping.ExternalType, mapping.ExternalStatus,
			mapping.ExternalAssignees, mapping.ExternalMetadata, mapping.AssetId, mapping.SyncedAt)
		if err != nil {
			return err
		}
	}

	return nil
}

func FetchData(remoteUrl string, userId string) (ProjectData, error) {
	userData := ProjectData{}
	userDataPb := repositorypb.ProjectData{}
	if utils.IsValidURL(remoteUrl) {
		type userTokenStruct struct {
			UserId string `json:"user_id"`
		}
		dataUrl := remoteUrl + "/data"

		userToken := userTokenStruct{
			UserId: userId,
		}
		jsonData, err := json.Marshal(userToken)
		if err != nil {
			return userData, err
		}

		req, err := http.NewRequest("GET", dataUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			return userData, err
		}
		req.Header.Set("Clustta-Agent", constants.USER_AGENT)
		auth_service.AttachBearerToken(req)

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			return userData, err
		}
		defer response.Body.Close()

		responseCode := response.StatusCode
		if responseCode == 200 {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return userData, fmt.Errorf("error reading response body: %s", err.Error())
			}

			decoder, err := kzstd.NewReader(nil)
			if err != nil {
				return userData, err
			}
			defer decoder.Close()
			decompressedData, err := decoder.DecodeAll(body, nil)
			if err != nil {
				return userData, err
			}

			err = proto.Unmarshal(decompressedData, &userDataPb)
			if err != nil {
				return userData, err
			}

			userData = ProjectData{
				ProjectPreview:      userDataPb.ProjectPreview,
				CollectionTypes:     repository.FromPbCollectionTypes(userDataPb.CollectionTypes),
				Collections:         repository.FromPbCollections(userDataPb.Collections),
				CollectionAssignees: repository.FromPbCollectionAssignees(userDataPb.CollectionAssignees),

				AssetTypes:             repository.FromPbAssetTypes(userDataPb.AssetTypes),
				Assets:                 repository.FromPbAssets(userDataPb.Assets),
				AssetCheckpoints:       repository.FromPbCheckpoints(userDataPb.AssetCheckpoints),
				AssetDependencies:      repository.FromPbAssetDependencies(userDataPb.AssetDependencies),
				CollectionDependencies: repository.FromPbCollectionDependencies(userDataPb.CollectionDependencies),

				Statuses:        repository.FromPbStatuses(userDataPb.Statuses),
				DependencyTypes: repository.FromPbDependencyTypes(userDataPb.DependencyTypes),

				Users: repository.FromPbUsers(userDataPb.Users),
				Roles: repository.FromPbRoles(userDataPb.Roles),

				Templates: repository.FromPbTemplates(userDataPb.Templates),

				Workflows:           repository.FromPbWorkflows(userDataPb.Workflows),
				WorkflowLinks:       repository.FromPbWorkflowLinks(userDataPb.WorkflowLinks),
				WorkflowCollections: repository.FromPbWorkflowCollections(userDataPb.WorkflowCollections),
				WorkflowAssets:      repository.FromPbWorkflowAssets(userDataPb.WorkflowAssets),

				Tags:      repository.FromPbTags(userDataPb.Tags),
				AssetTags: repository.FromPbAssetTags(userDataPb.AssetTags),

				IntegrationProjects:           repository.FromPbIntegrationProjects(userDataPb.IntegrationProjects),
				IntegrationCollectionMappings: repository.FromPbIntegrationCollectionMappings(userDataPb.IntegrationCollectionMappings),
				IntegrationAssetMappings:      repository.FromPbIntegrationAssetMappings(userDataPb.IntegrationAssetMappings),
			}

			return userData, nil
		} else if responseCode == 400 {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return userData, err
			}
			return userData, errors.New(string(body))
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return userData, err
		}
		return userData, fmt.Errorf("unknown error while fetching data. url: %s, status code: %d, message: %s", dataUrl, responseCode, string(body))
	} else if utils.FileExists(remoteUrl) {
		db, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return userData, err
		}
		defer db.Close()
		remoteTx, err := db.Beginx()
		if err != nil {
			return userData, err
		}
		defer remoteTx.Rollback()
		userData, err = LoadUserData(remoteTx, userId)
		if err != nil {
			return userData, err
		}
	} else {
		return userData, fmt.Errorf("invalid url:%s", remoteUrl)
	}
	return userData, nil

}

func CalculateMissingPreviews(tx *sqlx.Tx, data ProjectData) ([]string, error) {
	previewIds := []string{}
	previewSet := make(map[string]bool)
	if data.ProjectPreview != "" && !previewSet[data.ProjectPreview] {
		previewSet[data.ProjectPreview] = true
		previewIds = append(previewIds, data.ProjectPreview)
	}

	for _, asset := range data.Assets {
		if asset.PreviewId != "" && !previewSet[asset.PreviewId] {
			previewSet[asset.PreviewId] = true
			previewIds = append(previewIds, asset.PreviewId)
		}
	}
	for _, collection := range data.Collections {
		if collection.PreviewId != "" && !previewSet[collection.PreviewId] {
			previewSet[collection.PreviewId] = true
			previewIds = append(previewIds, collection.PreviewId)
		}
	}
	for _, assetCheckpoint := range data.AssetCheckpoints {
		if assetCheckpoint.PreviewId != "" && !previewSet[assetCheckpoint.PreviewId] {
			previewSet[assetCheckpoint.PreviewId] = true
			previewIds = append(previewIds, assetCheckpoint.PreviewId)
		}
	}

	missingPreviews, err := repository.GetNonExistingPreviews(tx, previewIds)
	return missingPreviews, err
}

func CalculateMissingChunks(tx *sqlx.Tx, data ProjectData, userId string, syncOptions SyncOptions) ([]string, []string, int, error) {
	assetIdSet := make(map[string]bool)

	for _, asset := range data.Assets {
		if asset.AssigneeId == userId {
			assetIdSet[asset.Id] = true
		} else if syncOptions.AssetDependencies && asset.IsDependency {
			assetIdSet[asset.Id] = true
		} else if syncOptions.Assets {
			assetIdSet[asset.Id] = true
		}
	}

	// Collect checkpoints based on OnlyLatestCheckpoints option
	var checkpointsToProcess []models.Checkpoint

	if syncOptions.OnlyLatestCheckpoints {
		// Maps to keep track of the latest checkpoint for each collection
		latestAssetCheckpoints := make(map[string]models.Checkpoint)
		// Iterate over asset checkpoints to find the latest for each collection
		for _, assetCheckpoint := range data.AssetCheckpoints {
			if assetIdSet[assetCheckpoint.AssetId] {
				existingCheckpoint, found := latestAssetCheckpoints[assetCheckpoint.AssetId]
				if !found || assetCheckpoint.CreatedAt > existingCheckpoint.CreatedAt {
					latestAssetCheckpoints[assetCheckpoint.AssetId] = assetCheckpoint
				}
			}
		}
		// Convert map to slice
		for _, checkpoint := range latestAssetCheckpoints {
			checkpointsToProcess = append(checkpointsToProcess, checkpoint)
		}
	} else {
		// Include ALL checkpoints for selected assets
		for _, assetCheckpoint := range data.AssetCheckpoints {
			if assetIdSet[assetCheckpoint.AssetId] {
				checkpointsToProcess = append(checkpointsToProcess, assetCheckpoint)
			}
		}
	}

	// Now gather all the chunks from the checkpoints
	seenChunks := make(map[string]bool)

	// Collect all unique hashes upfront for batch preloading.
	allHashSet := make(map[string]bool)
	for _, checkpoint := range checkpointsToProcess {
		for _, h := range strings.Split(checkpoint.Chunks, ",") {
			allHashSet[h] = true
		}
	}
	for _, template := range data.Templates {
		for _, h := range strings.Split(template.Chunks, ",") {
			allHashSet[h] = true
		}
	}
	allHashes := make([]string, 0, len(allHashSet))
	for h := range allHashSet {
		allHashes = append(allHashes, h)
	}
	chunk_service.PreloadChunkExistence(tx, allHashes, seenChunks)

	missingSet := make(map[string]bool)
	missingChunks := []string{}
	allChunks := []string{}
	totalSize := 0
	for _, checkpoint := range checkpointsToProcess {
		chunkHashes := strings.Split(checkpoint.Chunks, ",")
		checkpointFullyDownloaded := true
		for _, chunkHash := range chunkHashes {
			if chunk_service.ChunkExists(chunkHash, tx, seenChunks) {
				continue
			}
			if !missingSet[chunkHash] {
				checkpointFullyDownloaded = false
				missingSet[chunkHash] = true
				missingChunks = append(missingChunks, chunkHash)
			}
		}
		if !checkpointFullyDownloaded {
			totalSize += checkpoint.FileSize
			allChunks = append(allChunks, chunkHashes...)
		}
	}

	for _, template := range data.Templates {
		chunkHashes := strings.Split(template.Chunks, ",")
		templateFullyDownloaded := true
		for _, chunkHash := range chunkHashes {
			if chunk_service.ChunkExists(chunkHash, tx, seenChunks) {
				continue
			}
			if !missingSet[chunkHash] {
				templateFullyDownloaded = false
				missingSet[chunkHash] = true
				missingChunks = append(missingChunks, chunkHash)
			}
		}
		if !templateFullyDownloaded {
			totalSize += template.FileSize
			allChunks = append(allChunks, chunkHashes...)
		}
	}

	return missingChunks, allChunks, totalSize, nil
}
func CalculateCheckpointsMissingChunks(tx *sqlx.Tx, checkpoints []models.Checkpoint) ([]string, []string, int, error) {
	// Now gather all the chunks from the latest checkpoints
	seenChunks := make(map[string]bool)

	// Collect all unique hashes upfront for batch preloading.
	allHashSet := make(map[string]bool)
	for _, checkpoint := range checkpoints {
		for _, h := range strings.Split(checkpoint.Chunks, ",") {
			allHashSet[h] = true
		}
	}
	allHashes := make([]string, 0, len(allHashSet))
	for h := range allHashSet {
		allHashes = append(allHashes, h)
	}
	chunk_service.PreloadChunkExistence(tx, allHashes, seenChunks)

	missingSet := make(map[string]bool)
	missingChunks := []string{}
	allChunks := []string{}
	totalSize := 0
	for _, checkpoint := range checkpoints {
		chunkHashes := strings.Split(checkpoint.Chunks, ",")
		checkpointFullyDownloaded := true
		for _, chunkHash := range chunkHashes {
			if chunk_service.ChunkExists(chunkHash, tx, seenChunks) {
				continue
			}
			if !missingSet[chunkHash] {
				checkpointFullyDownloaded = false
				missingSet[chunkHash] = true
				missingChunks = append(missingChunks, chunkHash)
			}
		}
		if !checkpointFullyDownloaded {
			totalSize += checkpoint.FileSize
			allChunks = append(allChunks, chunkHashes...)
		}
	}

	return missingChunks, allChunks, totalSize, nil
}

func DownloadCheckpoint(ctx context.Context, projectPath, remoteUrl string, checkpointId string, userId string, callback func(int, int, string, string)) error {
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

	checkpoint, err := repository.GetCheckpoint(tx, checkpointId)
	if err != nil {
		return err
	}
	missingChunks, allChunks, totalSize, err := CalculateCheckpointsMissingChunks(tx, []models.Checkpoint{checkpoint})
	if err != nil {
		return err
	}
	tx.Rollback()

	if len(missingChunks) > 0 {
		err = chunk_service.PullStreamChunks(ctx, projectPath, remoteUrl, missingChunks, allChunks, totalSize, callback)
		if err != nil {
			return err
		}
	}
	return nil
}

func DownloadCheckpoints(ctx context.Context, projectPath, remoteUrl string, checkpointIds []string, userId string, callback func(int, int, string, string)) error {
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

	quotedcheckpointIds := make([]string, len(checkpointIds))
	for i, id := range checkpointIds {
		quotedcheckpointIds[i] = fmt.Sprintf("\"%s\"", id)
	}

	checkpoints := []models.Checkpoint{}
	err = tx.Select(&checkpoints, fmt.Sprintf("SELECT * FROM asset_checkpoint WHERE id IN (%s)", strings.Join(quotedcheckpointIds, ",")))
	if err != nil {
		return err
	}
	missingChunks, allChunks, totalSize, err := CalculateCheckpointsMissingChunks(tx, checkpoints)
	if err != nil {
		return err
	}
	tx.Rollback()

	if len(missingChunks) > 0 {
		err = chunk_service.PullStreamChunks(ctx, projectPath, remoteUrl, missingChunks, allChunks, totalSize, callback)
		if err != nil {
			return err
		}
	}
	return nil
}

func SyncData(workingData, remoteUrl string, user auth_service.User) error {
	dbConn, err := utils.OpenDb(workingData)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	tx, err := dbConn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	data, err := FetchData(remoteUrl, user.Id)
	if err != nil {
		return err
	}
	println("recieve")
	err = ClearLocalData(tx)
	if err != nil {
		return err
	}
	err = WriteProjectData(tx, data, false)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func IsUnsynced(tx *sqlx.Tx) (bool, error) {
	data, err := LoadChangedData(tx)
	if err != nil {
		return false, err
	}
	if data.IsEmpty() {
		return false, nil
	}
	return true, nil
}

func FetchChunksInfo(remoteUrl string, userId string, chunks []string) ([]chunk_service.ChunkInfo, error) {
	if utils.IsValidURL(remoteUrl) {

		dataUrl := remoteUrl + "/chunks-info"

		jsonData, err := json.Marshal(chunks)
		if err != nil {
			return []chunk_service.ChunkInfo{}, err
		}

		req, err := http.NewRequest("GET", dataUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			return []chunk_service.ChunkInfo{}, err
		}
		req.Header.Set("Clustta-Agent", constants.USER_AGENT)
		auth_service.AttachBearerToken(req)

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			return []chunk_service.ChunkInfo{}, err
		}
		defer response.Body.Close()

		responseCode := response.StatusCode
		responseData := []chunk_service.ChunkInfo{}
		if responseCode == 200 {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return []chunk_service.ChunkInfo{}, fmt.Errorf("error reading response body: %s", err.Error())
			}
			err = json.Unmarshal(body, &responseData)
			if err != nil {
				return []chunk_service.ChunkInfo{}, err
			}
			return responseData, nil
		} else {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return []chunk_service.ChunkInfo{}, err
			}
			println(string(body))
			return []chunk_service.ChunkInfo{}, errors.New(string(body))
		}
	} else if utils.FileExists(remoteUrl) {
		dbConn, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return []chunk_service.ChunkInfo{}, err
		}
		defer dbConn.Close()
		remoteTx, err := dbConn.Beginx()
		if err != nil {
			return []chunk_service.ChunkInfo{}, err
		}
		defer remoteTx.Rollback()

		ChunksInfo, err := chunk_service.GetChunksInfo(remoteTx, chunks)
		if err != nil {
			return []chunk_service.ChunkInfo{}, err
		}
		return ChunksInfo, nil
	} else {
		return []chunk_service.ChunkInfo{}, fmt.Errorf("invalid url:%s", remoteUrl)
	}
}

func FetchMissingChunks(ctx context.Context, remoteUrl string, userId string, chunks []string) ([]string, error) {
	if utils.IsValidURL(remoteUrl) {

		dataUrl := remoteUrl + "/chunks-missing"

		jsonData, err := json.Marshal(chunks)
		if err != nil {
			return []string{}, err
		}

		req, err := http.NewRequestWithContext(ctx, "GET", dataUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			return []string{}, err
		}
		req.Header.Set("Clustta-Agent", constants.USER_AGENT)
		auth_service.AttachBearerToken(req)

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			return []string{}, err
		}
		defer response.Body.Close()

		responseCode := response.StatusCode
		responseData := []string{}
		if responseCode == 200 {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return []string{}, fmt.Errorf("error reading response body: %s", err.Error())
			}
			err = json.Unmarshal(body, &responseData)
			if err != nil {
				return []string{}, err
			}
			return responseData, nil
		} else {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return []string{}, err
			}
			return []string{}, errors.New(string(body))
		}
	} else if utils.FileExists(remoteUrl) {
		dbConn, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return []string{}, err
		}
		defer dbConn.Close()
		remoteTx, err := dbConn.Beginx()
		if err != nil {
			return []string{}, err
		}
		defer remoteTx.Rollback()

		missingChunks := []string{}
		seenChunks := make(map[string]bool)
		chunk_service.PreloadChunkExistence(remoteTx, chunks, seenChunks)
		for _, chunkHash := range chunks {
			if chunk_service.ChunkExists(chunkHash, remoteTx, seenChunks) {
				continue
			}
			missingChunks = append(missingChunks, chunkHash)
		}
		return missingChunks, nil
	} else {
		return []string{}, fmt.Errorf("invalid url:%s", remoteUrl)
	}
}

func FetchMissingPreviews(ctx context.Context, remoteUrl string, userId string, previews []string) ([]string, error) {
	if utils.IsValidURL(remoteUrl) {

		dataUrl := remoteUrl + "/previews-exist"

		jsonData, err := json.Marshal(previews)
		if err != nil {
			return []string{}, err
		}

		req, err := http.NewRequestWithContext(ctx, "GET", dataUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			return []string{}, err
		}
		req.Header.Set("Clustta-Agent", constants.USER_AGENT)
		auth_service.AttachBearerToken(req)

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			return []string{}, err
		}
		defer response.Body.Close()

		responseCode := response.StatusCode
		responseData := []string{}
		if responseCode == 200 {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return []string{}, fmt.Errorf("error reading response body: %s", err.Error())
			}
			err = json.Unmarshal(body, &responseData)
			if err != nil {
				return []string{}, err
			}
			return responseData, nil
		} else {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return []string{}, err
			}
			return []string{}, errors.New(string(body))
		}
	} else if utils.FileExists(remoteUrl) {
		dbConn, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return []string{}, err
		}
		defer dbConn.Close()
		remoteTx, err := dbConn.Beginx()
		if err != nil {
			return []string{}, err
		}
		defer remoteTx.Rollback()

		missingPreviews := []string{}
		for _, previewHash := range previews {
			if repository.PreviewExists(previewHash, remoteTx) {
				continue
			}
			missingPreviews = append(missingPreviews, previewHash)
		}
		return missingPreviews, nil
	} else {
		return []string{}, fmt.Errorf("invalid url:%s", remoteUrl)
	}
}
