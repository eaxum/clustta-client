package services

import (
	"clustta/internal/activity"
	"clustta/internal/repository/models"
	"clustta/output"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestActivityRetainsDownloadDetailsDuringRestoration(t *testing.T) {
	id, ctx := activities.Start(context.Background(), activity.Operation{Title: "Fetching assets"})
	t.Cleanup(func() { activities.Finish(id, nil); activities.Dismiss(id) })
	action := &transferAction{id: id, ctx: ctx}
	action.assets([]models.Asset{{Id: "asset", Name: "Scene", Extension: ".blend", CollectionId: "collection"}})
	action.downloaded(800, 1000, "Receiving 800 B/1 KB", "Data saved: 400 B (50.00%)")
	action.report(output.ProgressReport{Message: "Restoring Scene.blend", Current: 1, Total: 2, Percentage: 50})
	for _, operation := range activities.List().Operations {
		if operation.ID != id {
			continue
		}
		if operation.Title != "Fetching Scene" {
			t.Fatalf("generic fetch title retained: %s", operation.Title)
		}
		if operation.DownloadMessage != "Receiving 800 B/1 KB" || operation.ExtraMessage != "Data saved: 400 B (50.00%)" {
			t.Fatalf("download details lost during restoration: %+v", operation)
		}
		if len(operation.Assets) != 1 || operation.Assets[0].Extension != ".blend" || operation.Assets[0].Name != "Scene" || operation.Assets[0].CollectionID != "collection" {
			t.Fatalf("asset display details missing: %+v", operation.Assets)
		}
		return
	}
	t.Fatal("activity missing")
}

func TestTransferRejectsFileCreatedDuringDownload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "asset.blend")
	state, err := captureTransferFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.verify(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("local changes"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := state.verify(path); err == nil {
		t.Fatal("would overwrite a file created during download")
	}
}

func TestActivityRebuildingProgressSpansAllAssets(t *testing.T) {
	id, ctx := activities.Start(context.Background(), activity.Operation{})
	t.Cleanup(func() { activities.Finish(id, nil); activities.Dismiss(id) })
	action := &transferAction{id: id, ctx: ctx}
	for _, step := range []struct {
		index, current, total int
		want                  float64
	}{
		{0, 50, 100, 25}, {0, 100, 100, 50}, {1, 0, 100, 50}, {1, 50, 100, 75}, {1, 100, 100, 99},
	} {
		action.rebuilding(step.index, 2, step.current, step.total, "Asset")
		for _, operation := range activities.List().Operations {
			if operation.ID == id && operation.Percentage != step.want {
				t.Fatalf("progress = %v, want %v", operation.Percentage, step.want)
			}
		}
	}
}
