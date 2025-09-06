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

	ErrTaskExists             = errors.New("task of same name exists")
	ErrTaskTypeNotFound       = errors.New("task type not found")
	ErrTaskTypeExists         = errors.New("task type already exist")
	ErrTaskExistsInTrash      = errors.New("task of same name exists in trash")
	ErrNotAutheticated        = errors.New("user not autheticated")
	ErrNotUnauthorized        = errors.New("user unauthorized")
	ErrMustHaveAdmin          = errors.New("must have at least one admin")
	ErrTaskNotFound           = errors.New("task not found")
	ErrTaskCheckPointNotFound = errors.New("task checkpoint not found")

	ErrCheckpointExists   = errors.New("check point already exists")
	ErrCheckpointNotFound = errors.New("check point not found")

	ErrEntityNotFound         = errors.New("entity not found")
	ErrEntityAssigneeNotFound = errors.New("entity assignee not found")
	ErrEntityTypeNotFound     = errors.New("entity type not found")
	ErrEntityTypeExists       = errors.New("entity type already exist")
	ErrEntityExists           = errors.New("entity already exists")
	ErrEntityExistsInTrash    = errors.New("entity already exists in trash")

	ErrStatusNotFound           = errors.New("task status not found")
	ErrUserNotFound             = errors.New("user not found")
	ErrRoleNotFound             = errors.New("role not found")
	ErrUserHaveTaskAssigned     = errors.New("user have task assigned")
	ErrDependencyTypeNotFound   = errors.New("task dependency type not found")
	ErrTaskDependencyNotFound   = errors.New("task dependency not found")
	ErrEntityDependencyNotFound = errors.New("entity dependency not found")

	ErrWorkflowExists   = errors.New("workflow of same name exists")
	ErrWorkflowNotFound = errors.New("workflow not found")

	ErrWorkflowEntityExists   = errors.New("workflow entity of same name exists")
	ErrWorkflowEntityNotFound = errors.New("workflow entity not found")
	ErrWorkflowTaskExists     = errors.New("workflow task of same name exists")
	ErrWorkflowTaskNotFound   = errors.New("workflow task not found")
	ErrWorkflowLinkNotFound   = errors.New("workflow link not found")
	ErrWorkflowLinkExists     = errors.New("workflow link of same name exists")

	ErrTemplateNotFound = errors.New("template not found")

	ErrTagNotFound     = errors.New("tag not found")
	ErrTaskTagNotFound = errors.New("task tag not found")

	ErrPreviewNotFound = errors.New("preview not found")

	ErrNoRows       = errors.New("sql: no rows in result set")
	ErrUnauthorized = errors.New("Unauthorized")
)

func IsConnectionResetError(err error) bool {
	return strings.Contains(err.Error(), "wsarecv: An existing connection was forcibly closed by the remote host")
}
