package services

import (
	"context"
	"sync"
)

// ctx, cancel := context.WithCancel(context.Background())
// mu := sync.Mutex{}

// create a new context and cancel function

var mu sync.Mutex
var cancel context.CancelFunc
var ctx context.Context

func init() {
	ctx, cancel = context.WithCancel(context.Background())
}

func cancelSync() {
	mu.Lock()
	if cancel != nil {
		cancel()
	}
	mu.Unlock()
}

func reset() {
	mu.Lock()
	defer mu.Unlock()
	if cancel != nil {
		cancel()
	}
	ctx, cancel = context.WithCancel(context.Background())
}

// getContext returns the current context in a thread-safe way
func getContext() context.Context {
	mu.Lock()
	defer mu.Unlock()
	return ctx
}
