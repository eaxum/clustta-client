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

	ErrTaskExists             = errors.New("asset of same name exists")
	ErrTaskTypeNotFound       = errors.New("asset type not found")
	ErrTaskTypeExists         = errors.New("asset type already exist")
	ErrTaskExistsInTrash      = errors.New("asset of same name exists in trash")
	ErrNotAutheticated        = errors.New("user not autheticated")
	ErrNotUnauthorized        = errors.New("user unauthorized")
	ErrMustHaveAdmin          = errors.New("must have at least one admin")
	ErrTaskNotFound           = errors.New("asset not found")
	ErrTaskCheckPointNotFound = errors.New("asset checkpoint not found")

	ErrCheckpointExists   = errors.New("checkpoint already exists")
	ErrCheckpointNotFound = errors.New("checkpoint not found")

	ErrEntityNotFound         = errors.New("collection not found")
	ErrEntityAssigneeNotFound = errors.New("collection assignee not found")
	ErrEntityTypeNotFound     = errors.New("collection type not found")
	ErrEntityTypeExists       = errors.New("collection type already exist")
	ErrEntityExists           = errors.New("collection already exists")
	ErrEntityExistsInTrash    = errors.New("collection already exists in trash")

	ErrStatusNotFound           = errors.New("asset status not found")
	ErrUserNotFound             = errors.New("user not found")
	ErrRoleNotFound             = errors.New("role not found")
	ErrUserHaveTaskAssigned     = errors.New("user have asset assigned")
	ErrDependencyTypeNotFound   = errors.New("asset dependency type not found")
	ErrTaskDependencyNotFound   = errors.New("asset dependency not found")
	ErrEntityDependencyNotFound = errors.New("collection dependency not found")

	ErrWorkflowExists   = errors.New("workflow of same name exists")
	ErrWorkflowNotFound = errors.New("workflow not found")

	ErrWorkflowEntityExists   = errors.New("workflow collection of same name exists")
	ErrWorkflowEntityNotFound = errors.New("workflow collection not found")
	ErrWorkflowTaskExists     = errors.New("workflow asset of same name exists")
	ErrWorkflowTaskNotFound   = errors.New("workflow asset not found")
	ErrWorkflowLinkNotFound   = errors.New("workflow link not found")
	ErrWorkflowLinkExists     = errors.New("workflow link of same name exists")

	ErrTemplateNotFound = errors.New("template not found")

	ErrTagNotFound     = errors.New("tag not found")
	ErrTaskTagNotFound = errors.New("asset tag not found")

	ErrPreviewNotFound = errors.New("preview not found")

	ErrNoRows       = errors.New("sql: no rows in result set")
	ErrUnauthorized = errors.New("Unauthorized")
)

func IsConnectionResetError(err error) bool {
	return strings.Contains(err.Error(), "wsarecv: An existing connection was forcibly closed by the remote host")
}
