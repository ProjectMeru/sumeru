package orm

import (
	"context"
	"strings"
	"sync"
)

// WriteGuardFunc validates an update before it is applied. before holds the
// current row, values the prepared columns to be written. Returning an error
// aborts the write (the message is surfaced to the user by the RPC layer).
type WriteGuardFunc func(ctx context.Context, model string, before map[string]interface{}, values map[string]interface{}) error

// UnlinkGuardFunc validates deleting a record before it is removed.
type UnlinkGuardFunc func(ctx context.Context, model string, record map[string]interface{}) error

var (
	guardMu      sync.RWMutex
	writeGuards  = map[string][]WriteGuardFunc{}
	unlinkGuards = map[string][]UnlinkGuardFunc{}
)

// RegisterWriteGuard registers a per-model write validator. Guards run for every
// update of the model, including the superuser (bypass contexts are exempt so
// internal/system flows keep working).
func RegisterWriteGuard(model string, fn WriteGuardFunc) {
	model = strings.TrimSpace(model)
	if model == "" || fn == nil {
		return
	}
	guardMu.Lock()
	defer guardMu.Unlock()
	writeGuards[model] = append(writeGuards[model], fn)
}

// RegisterUnlinkGuard registers a per-model delete validator.
func RegisterUnlinkGuard(model string, fn UnlinkGuardFunc) {
	model = strings.TrimSpace(model)
	if model == "" || fn == nil {
		return
	}
	guardMu.Lock()
	defer guardMu.Unlock()
	unlinkGuards[model] = append(unlinkGuards[model], fn)
}

// RunWriteGuards invokes all write guards registered for model.
func RunWriteGuards(ctx context.Context, model string, before map[string]interface{}, values map[string]interface{}) error {
	if SecurityBypass(ctx) {
		return nil
	}
	guardMu.RLock()
	fns := append([]WriteGuardFunc(nil), writeGuards[strings.TrimSpace(model)]...)
	guardMu.RUnlock()
	for _, fn := range fns {
		if err := fn(ctx, model, before, values); err != nil {
			return err
		}
	}
	return nil
}

// RunUnlinkGuards invokes all unlink guards registered for model.
func RunUnlinkGuards(ctx context.Context, model string, record map[string]interface{}) error {
	if SecurityBypass(ctx) {
		return nil
	}
	guardMu.RLock()
	fns := append([]UnlinkGuardFunc(nil), unlinkGuards[strings.TrimSpace(model)]...)
	guardMu.RUnlock()
	for _, fn := range fns {
		if err := fn(ctx, model, record); err != nil {
			return err
		}
	}
	return nil
}
