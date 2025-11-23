package services

import (
	"context"
	"sync"
)

var mu sync.Mutex
var cancel context.CancelFunc
var ctx context.Context

func init() {
	ctx, cancel = context.WithCancel(context.Background())
}

//cancelSync cancels the current context in a thread-safe manner.
//Acquires a lock before canceling to ensure thread safety.
func cancelSync() {
	mu.Lock()
	if cancel != nil {
		cancel()
	}
	mu.Unlock()
}

//reset cancels the current context and creates a new one.
//Used to reset the context state after an operation completes or is cancelled.
func reset() {
	mu.Lock()
	defer mu.Unlock()
	if cancel != nil {
		cancel()
	}
	ctx, cancel = context.WithCancel(context.Background())
}

//getContext returns the current context in a thread-safe way.
//Acquires a lock to safely read the context variable.
func getContext() context.Context {
	mu.Lock()
	defer mu.Unlock()
	return ctx
}
