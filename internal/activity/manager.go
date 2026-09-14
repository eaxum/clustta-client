package activity

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	Queued           = "queued"
	Running          = "running"
	Completed        = "completed"
	Cancelled        = "cancelled"
	Failed           = "failed"
	progressInterval = 100 * time.Millisecond
)

type Asset struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Extension      string `json:"extension"`
	CollectionID   string `json:"collection_id"`
	CollectionPath string `json:"collection_path"`
}

type Collection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"collection_path"`
}

type Operation struct {
	ID               string       `json:"operation_id"`
	ProjectID        string       `json:"project_id"`
	ProjectURI       string       `json:"project_uri"`
	ProjectName      string       `json:"project_name"`
	Kind             string       `json:"kind"`
	Title            string       `json:"title"`
	Status           string       `json:"status"`
	CancelRequested  bool         `json:"cancel_requested"`
	Message          string       `json:"message"`
	ExtraMessage     string       `json:"extra_message"`
	DownloadMessage  string       `json:"download_message"`
	Percentage       float64      `json:"percentage"`
	Current          int          `json:"current"`
	Total            int          `json:"total"`
	Assets           []Asset      `json:"assets"`
	Collections      []Collection `json:"collections"`
	Error            string       `json:"error"`
	RestoredAssetIDs []string     `json:"restored_asset_ids"`
}

type Snapshot struct {
	Revision   uint64      `json:"revision"`
	Operations []Operation `json:"operations"`
}

type entry struct {
	Operation
	cancel context.CancelFunc
	ready  chan struct{}
}

type Manager struct {
	mu       sync.Mutex
	entries  []*entry
	revision uint64
	emit     func(Snapshot)
	timer    *time.Timer
}

func NewManager(emit func(Snapshot)) *Manager { return &Manager{emit: emit} }

func (m *Manager) Start(parent context.Context, operation Operation) (string, context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	operation.ID = uuid.NewString()
	operation.Status = Running
	operation.Assets = append([]Asset{}, operation.Assets...)
	operation.Collections = append([]Collection{}, operation.Collections...)
	for _, item := range m.entries {
		if item.Status == Running {
			operation.Status = Queued
			break
		}
	}
	ctx, cancel := context.WithCancel(parent)
	item := &entry{Operation: operation, cancel: cancel, ready: make(chan struct{})}
	if operation.Status == Running {
		close(item.ready)
	}
	m.entries = append(m.entries, item)
	m.publish()
	return operation.ID, ctx
}

// Wait admits one action at a time, in registration order, across all projects.
func (m *Manager) Wait(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	var ready <-chan struct{}
	for _, item := range m.entries {
		if item.ID == id {
			ready = item.ready
			break
		}
	}
	m.mu.Unlock()
	if ready == nil {
		return errors.New("activity not found")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ready:
		return ctx.Err()
	}
}

func (m *Manager) Update(id string, update func(*Operation)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range m.entries {
		if item.ID != id || (item.Status != Running && item.Status != Queued) {
			continue
		}
		update(&item.Operation)
		if m.timer == nil {
			m.timer = time.AfterFunc(progressInterval, func() { m.mu.Lock(); defer m.mu.Unlock(); m.publish() })
		}
		return
	}
}

func (m *Manager) Finish(id string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range m.entries {
		if item.ID != id || (item.Status != Running && item.Status != Queued) {
			continue
		}
		wasRunning := item.Status == Running
		item.cancel()
		item.Status = Completed
		if errors.Is(err, context.Canceled) {
			item.Status = Cancelled
		} else if err != nil {
			item.Status = Failed
			var requestError *url.Error
			if errors.As(err, &requestError) {
				item.Error = requestError.Err.Error()
			} else {
				item.Error = err.Error()
			}
		} else {
			item.Percentage = 100
		}
		if wasRunning {
			for _, next := range m.entries {
				if next.Status == Queued {
					next.Status = Running
					close(next.ready)
					break
				}
			}
		}
		m.publish()
		return
	}
}

func (m *Manager) Cancel(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range m.entries {
		if item.ID != id {
			continue
		}
		if item.Status == Running || item.Status == Queued {
			if item.Status == Queued {
				item.Status = Cancelled
			}
			item.CancelRequested = true
			item.cancel()
			m.publish()
		}
		return nil
	}
	return errors.New("activity not found")
}

func (m *Manager) Dismiss(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	remaining := m.entries[:0]
	for _, item := range m.entries {
		if item.Status == Running || item.Status == Queued || (id != "" && item.ID != id) {
			remaining = append(remaining, item)
		}
	}
	m.entries = remaining
	m.publish()
}

func (m *Manager) List() Snapshot { m.mu.Lock(); defer m.mu.Unlock(); return m.snapshot() }

func (m *Manager) snapshot() Snapshot {
	result := Snapshot{Revision: m.revision, Operations: []Operation{}}
	for _, item := range m.entries {
		operation := item.Operation
		operation.Assets = append([]Asset{}, item.Assets...)
		operation.Collections = append([]Collection{}, item.Collections...)
		operation.RestoredAssetIDs = append([]string{}, item.RestoredAssetIDs...)
		result.Operations = append(result.Operations, operation)
	}
	return result
}

func (m *Manager) publish() {
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
	m.revision++
	if m.emit != nil {
		m.emit(m.snapshot())
	}
}
