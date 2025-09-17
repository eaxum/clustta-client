package auth_service_test

import (
	"clustta/internal/auth_service"
	"testing"
)

func TestGetActiveUser(t *testing.T) {
	_, err := auth_service.GetActiveUser()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
