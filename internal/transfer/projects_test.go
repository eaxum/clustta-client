package transfer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func receive[T any](t *testing.T, channel <-chan T) T {
	t.Helper()
	select {
	case value := <-channel:
		return value
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for cleanup")
		var zero T
		return zero
	}
}

func TestProjectAdmissionAndDeferredCleanup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "project.clst")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalProject(path)
	if err != nil {
		t.Fatal(err)
	}
	first, err := Enter(canonical)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Enter(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if release, err := Exclusive(path); err == nil {
		release()
		t.Fatal("exclusive write admitted during downloads")
	}
	cleaned := make(chan bool, 1)
	CleanupWhenIdle(canonical, func() { cleaned <- true })
	first()
	select {
	case <-cleaned:
		t.Fatal("cleaned chunks while a queued action remained")
	default:
	}
	second()
	receive(t, cleaned)
}

func TestWorkingDirectoriesCannotOverlapAcrossProjects(t *testing.T) {
	directory := t.TempDir()
	release, err := ReserveDestination("project-a", directory)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if other, err := ReserveDestination("project-b", filepath.Join(directory, "nested")); err == nil {
		other()
		t.Fatal("overlapping destination admitted")
	}
	if other, err := ReserveDestination("project-a", directory); err != nil {
		t.Fatal(err)
	} else {
		other()
	}
	if remove, err := Removal(directory); err == nil {
		remove()
		t.Fatal("active working directory can be deleted")
	}
}

func TestWaitingWriteCanBeCancelled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "project.clst")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	release, err := Write(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if acquired, err := Write(ctx, path); err == nil {
		acquired()
		t.Fatal("cancelled writer acquired the lock")
	}
}
