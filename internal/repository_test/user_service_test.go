package repository

import (
	"clustta/internal/repository"
	"clustta/internal/testutils"
	"testing"
)

func TestGetUsers(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetUsers(testutils.Tx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
