package transfer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type projectGate struct {
	users     int
	exclusive bool
	write     chan struct{}
	cleanup   func()
}

var projectMu sync.Mutex
var projects = map[string]*projectGate{}

type destinationLease struct {
	project string
	users   int
}

var destinations = map[string]*destinationLease{}
var removals = map[string]int{}

func ReserveDestination(project, directory string) (func(), error) {
	root, err := canonicalPath(directory)
	if err != nil {
		return nil, err
	}
	projectMu.Lock()
	defer projectMu.Unlock()
	for path := range removals {
		if pathsOverlap(path, root) {
			return nil, errors.New("files in this working directory are being removed")
		}
	}
	for path, lease := range destinations {
		if lease.project != project && pathsOverlap(path, root) {
			return nil, errors.New("another project is downloading into this working directory")
		}
	}
	lease := destinations[root]
	if lease == nil {
		lease = &destinationLease{project: project}
		destinations[root] = lease
	}
	lease.users++
	return func() {
		projectMu.Lock()
		defer projectMu.Unlock()
		lease.users--
		if lease.users == 0 {
			delete(destinations, root)
		}
	}, nil
}

func pathsOverlap(left, right string) bool {
	separator := string(filepath.Separator)
	return left == right || strings.HasPrefix(left, strings.TrimRight(right, separator)+separator) || strings.HasPrefix(right, strings.TrimRight(left, separator)+separator)
}

func Removal(path string) (func(), error) {
	resolved, err := canonicalPath(path)
	if err != nil {
		return nil, err
	}
	projectMu.Lock()
	defer projectMu.Unlock()
	for root := range destinations {
		if pathsOverlap(root, resolved) {
			return nil, errors.New("cannot remove files while this project has active transfers")
		}
	}
	for project, gate := range projects {
		if (gate.users > 0 || gate.exclusive) && pathsOverlap(project, resolved) {
			return nil, errors.New("cannot remove a project with active transfers")
		}
	}
	removals[resolved]++
	return func() {
		projectMu.Lock()
		defer projectMu.Unlock()
		removals[resolved]--
		if removals[resolved] == 0 {
			delete(removals, resolved)
		}
	}, nil
}

func CanonicalProject(path string) (string, error) { return canonicalPath(path) }

func canonicalPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	suffix := []string{}
	for {
		resolved, err := filepath.EvalSymlinks(absolute)
		if err == nil {
			for index := len(suffix) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, suffix[index])
			}
			if runtime.GOOS == "windows" {
				resolved = strings.ToLower(resolved)
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) || filepath.Dir(absolute) == absolute {
			return "", err
		}
		suffix = append(suffix, filepath.Base(absolute))
		absolute = filepath.Dir(absolute)
	}
}

func gateFor(path string) *projectGate {
	gate := projects[path]
	if gate == nil {
		gate = &projectGate{write: make(chan struct{}, 1)}
		projects[path] = gate
	}
	return gate
}

// Enter keeps cleanup/exclusive writes out while an active or queued action reserves the project.
func Enter(path string) (func(), error) {
	projectMu.Lock()
	defer projectMu.Unlock()
	for removing := range removals {
		if pathsOverlap(removing, path) {
			return nil, errors.New("project files are being removed")
		}
	}
	gate := gateFor(path)
	if gate.exclusive {
		return nil, errors.New("project is busy; retry after the current operation finishes")
	}
	gate.users++
	return func() {
		projectMu.Lock()
		gate.users--
		startCleanup(gate)
		projectMu.Unlock()
	}, nil
}

func CleanupWhenIdle(path string, cleanup func()) {
	projectMu.Lock()
	defer projectMu.Unlock()
	gate := gateFor(path)
	gate.cleanup = cleanup
	startCleanup(gate)
}

func startCleanup(gate *projectGate) {
	if gate.users != 0 || gate.exclusive || gate.cleanup == nil {
		return
	}
	cleanup := gate.cleanup
	gate.cleanup = nil
	gate.exclusive = true
	go func() {
		cleanup()
		projectMu.Lock()
		gate.exclusive = false
		startCleanup(gate)
		projectMu.Unlock()
	}()
}

func Exclusive(path string) (func(), error) {
	canonical, err := CanonicalProject(path)
	if err != nil {
		return nil, err
	}
	projectMu.Lock()
	defer projectMu.Unlock()
	gate := gateFor(canonical)
	if gate.exclusive || gate.users != 0 {
		return nil, errors.New("project has active transfers; retry when they finish")
	}
	gate.exclusive = true
	return func() { projectMu.Lock(); gate.exclusive = false; projectMu.Unlock() }, nil
}

func Write(ctx context.Context, path string) (func(), error) {
	canonical, err := CanonicalProject(path)
	if err != nil {
		return nil, err
	}
	projectMu.Lock()
	gate := gateFor(canonical)
	projectMu.Unlock()
	select {
	case gate.write <- struct{}{}:
		if err := ctx.Err(); err != nil {
			<-gate.write
			return nil, err
		}
		return func() { <-gate.write }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
