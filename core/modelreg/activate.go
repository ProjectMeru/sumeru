package modelreg

import (
	"fmt"
	"sort"
	"sync"

	"sumeru/core/orm"
)

type pendingModel struct {
	name   string
	extend bool
	fields []orm.FieldDefinition
}

var (
	pendingMu       sync.Mutex
	pendingByModule = map[string][]pendingModel{}
	activated       bool
)

func queueRegistration(moduleName string, models []pendingModel) {
	if len(models) == 0 {
		return
	}
	pendingMu.Lock()
	defer pendingMu.Unlock()
	pendingByModule[moduleName] = append(pendingByModule[moduleName], models...)
}

// ActivateAll registers queued models in moduleOrder (manifest dependency order).
// When moduleOrder is empty, modules are activated alphabetically.
func ActivateAll(moduleOrder []string) error {
	pendingMu.Lock()
	defer pendingMu.Unlock()
	if activated {
		return nil
	}
	if err := activateAllLocked(moduleOrder); err != nil {
		return err
	}
	activated = true
	return nil
}

// ResetActivationForTest clears activation state and the pending queue. Test-only.
func ResetActivationForTest() {
	pendingMu.Lock()
	defer pendingMu.Unlock()
	pendingByModule = map[string][]pendingModel{}
	activated = false
}

func activateAllLocked(moduleOrder []string) error {
	order := moduleOrderForActivation(moduleOrder)
	for _, moduleName := range order {
		for _, pending := range pendingByModule[moduleName] {
			if pending.extend {
				extendRegisteredModel(pending.name, pending.fields, moduleName)
				continue
			}
			orm.RegisterModelWithModule(&reflectedModel{name: pending.name, fields: pending.fields}, moduleName)
		}
	}
	return nil
}

func moduleOrderForActivation(moduleOrder []string) []string {
	if len(moduleOrder) == 0 {
		names := make([]string, 0, len(pendingByModule))
		for name := range pendingByModule {
			names = append(names, name)
		}
		sort.Strings(names)
		return names
	}
	var order []string
	seen := map[string]struct{}{}
	for _, name := range moduleOrder {
		if _, queued := pendingByModule[name]; !queued {
			continue
		}
		order = append(order, name)
		seen[name] = struct{}{}
	}
	var extras []string
	for name := range pendingByModule {
		if _, ok := seen[name]; ok {
			continue
		}
		extras = append(extras, name)
	}
	sort.Strings(extras)
	return append(order, extras...)
}

func mustActivateForTest() {
	if err := ActivateAll(nil); err != nil {
		panic(fmt.Sprintf("modelreg.ActivateAll: %v", err))
	}
}
