package error_service

import (
	"errors"
	"strings"
)

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectExists        = errors.New("project already exists")
	ErrInvalidProject       = errors.New("uri not a valid project")
	ErrInvalidProjectExists = errors.New("uri not a valid project, but exists")

	ErrAssetExists             = errors.New("asset of same name exists")
	ErrAssetTypeNotFound       = errors.New("asset type not found")
	ErrAssetTypeExists         = errors.New("asset type already exist")
	ErrAssetExistsInTrash      = errors.New("asset of same name exists in trash")
	ErrNotAutheticated         = errors.New("user not autheticated")
	ErrNotUnauthorized         = errors.New("user unauthorized")
	ErrMustHaveAdmin           = errors.New("must have at least one admin")
	ErrAssetNotFound           = errors.New("asset not found")
	ErrAssetCheckPointNotFound = errors.New("asset checkpoint not found")

	ErrCheckpointExists   = errors.New("checkpoint already exists")
	ErrCheckpointNotFound = errors.New("checkpoint not found")

	ErrCollectionNotFound         = errors.New("collection not found")
	ErrCollectionAssigneeNotFound = errors.New("collection assignee not found")
	ErrCollectionTypeNotFound     = errors.New("collection type not found")
	ErrCollectionTypeExists       = errors.New("collection type already exist")
	ErrCollectionExists           = errors.New("collection already exists")
	ErrCollectionExistsInTrash    = errors.New("collection already exists in trash")

	ErrStatusNotFound               = errors.New("asset status not found")
	ErrUserNotFound                 = errors.New("user not found")
	ErrRoleNotFound                 = errors.New("role not found")
	ErrUserHaveAssetAssigned        = errors.New("user have asset assigned")
	ErrDependencyTypeNotFound       = errors.New("asset dependency type not found")
	ErrAssetDependencyNotFound      = errors.New("asset dependency not found")
	ErrCollectionDependencyNotFound = errors.New("collection dependency not found")

	ErrWorkflowExists   = errors.New("workflow of same name exists")
	ErrWorkflowNotFound = errors.New("workflow not found")

	ErrWorkflowCollectionExists   = errors.New("workflow collection of same name exists")
	ErrWorkflowCollectionNotFound = errors.New("workflow collection not found")
	ErrWorkflowAssetExists        = errors.New("workflow asset of same name exists")
	ErrWorkflowAssetNotFound      = errors.New("workflow asset not found")
	ErrWorkflowLinkNotFound       = errors.New("workflow link not found")
	ErrWorkflowLinkExists         = errors.New("workflow link of same name exists")

	ErrTemplateNotFound = errors.New("template not found")

	ErrTagNotFound      = errors.New("tag not found")
	ErrTagExists        = errors.New("tag already exists")
	ErrAssetTagNotFound = errors.New("asset tag not found")

	ErrPreviewNotFound = errors.New("preview not found")

	ErrNoRows       = errors.New("sql: no rows in result set")
	ErrUnauthorized = errors.New("Unauthorized")
)

func IsConnectionResetError(err error) bool {
	return strings.Contains(err.Error(), "wsarecv: An existing connection was forcibly closed by the remote host")
}
