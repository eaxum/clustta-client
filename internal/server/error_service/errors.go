package error_service

import (
	"errors"
	"strings"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrNoRows       = errors.New("sql: no rows in result set")
	ErrUnauthorized = errors.New("Unauthorized")
)

func IsConnectionResetError(err error) bool {
	return strings.Contains(err.Error(), "wsarecv: An existing connection was forcibly closed by the remote host")
}
