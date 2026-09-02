package sdk

import (
	"context"
	"sync"
)

var (
	startupMu sync.RWMutex
	startups  []func(context.Context) error
)

// RegisterStartup registers a function called after DB init and module load, before HTTP listen.
func RegisterStartup(fn func(context.Context) error) {
	if fn == nil {
		return
	}
	startupMu.Lock()
	defer startupMu.Unlock()
	startups = append(startups, fn)
}

// ClearStartups removes registered startup hooks (tests).
func ClearStartups() {
	startupMu.Lock()
	defer startupMu.Unlock()
	startups = nil
}

// RunStartups invokes all registered startup hooks in registration order.
func RunStartups(ctx context.Context) error {
	startupMu.RLock()
	list := append([]func(context.Context) error(nil), startups...)
	startupMu.RUnlock()
	for _, fn := range list {
		if err := fn(ctx); err != nil {
			return err
		}
	}
	return nil
}
