package repository

import (
	"clustta/internal/repository"
	"clustta/internal/testutils"
	"path/filepath"
	"testing"
)

func TestCreateTemplate(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.CreateTemplate(testutils.Tx, "Template 1", testutils.TestFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestGetTemplates(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()

	_, err := repository.GetTemplates(testutils.Tx, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRenameTemplate(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	template, err := repository.CreateTemplate(testutils.Tx, "Template 1", testutils.TestFile)
	if err != nil {
		t.Errorf(err.Error())
	}
	err = repository.RenameTemplate(testutils.Tx, template.Id, "renamed")
	if err != nil {
		t.Errorf(err.Error())
	}
	renameTemplate, err := repository.GetTemplateByName(testutils.Tx, "renamed")
	if err != nil {
		t.Errorf(err.Error())
	} else {
		if renameTemplate.Name != "renamed" {
			t.Errorf("Unexpected template name: %s", renameTemplate.Name)
		}
	}

}

func TestUpdateTemplateFile(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	template, err := repository.CreateTemplate(testutils.Tx, "Template 1", testutils.TestFile)
	if err != nil {
		t.Errorf(err.Error())
	}
	updatedTemplate, err := repository.UpdateTemplateFile(testutils.Tx, template.Id, testutils.TestPreviewFile)
	if err != nil {
		t.Errorf(err.Error())
	} else {
		if updatedTemplate.Extension != filepath.Ext(testutils.TestPreviewFile) {
			t.Errorf("Unexpected template file extension: %s", updatedTemplate.Extension)
		}
	}
}

func TestGetOrCreateTemplate(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	_, err := repository.GetOrCreateTemplate(testutils.Tx, "New Template", testutils.TestFile)
	if err != nil {
		t.Errorf(err.Error())
	}

}

func TestDeleteTemplate(t *testing.T) {
	testutils.Setup()
	defer testutils.Teardown()
	template, err := repository.CreateTemplate(testutils.Tx, "Template 1", testutils.TestFile)
	if err != nil {
		t.Errorf(err.Error())
	}
	err = repository.DeleteTemplate(testutils.Tx, template.Id, true)
	if err != nil {
		t.Errorf(err.Error())
	}
	deletedTemplates, err := repository.GetDeletedTemplates(testutils.Tx)
	if err != nil {
		t.Errorf(err.Error())
	} else {
		if len(deletedTemplates) != 1 {
			t.Errorf("expected 1 deleted template, but got %d", len(deletedTemplates))
		}

	}

}
