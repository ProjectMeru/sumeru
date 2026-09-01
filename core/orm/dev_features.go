package orm

import (
	"strings"
	"sync"

	"sumeru/core/server/config"
)

var (
	devFeatureMu sync.RWMutex
	devFeatures  map[string]bool
)

// InitDevFeatures parses dev_features INI (comma-separated: sql, access, xml).
func InitDevFeatures(raw string) {
	devFeatureMu.Lock()
	defer devFeatureMu.Unlock()
	devFeatures = map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if part != "" {
			devFeatures[part] = true
		}
	}
	if config.AppConfig.DevMode {
		devFeatures["access"] = true
	}
}

// DevFeatureEnabled reports whether a --dev feature flag is active.
func DevFeatureEnabled(name string) bool {
	devFeatureMu.RLock()
	defer devFeatureMu.RUnlock()
	if devFeatures == nil {
		return config.AppConfig.DevMode && strings.EqualFold(strings.TrimSpace(name), "access")
	}
	return devFeatures[strings.ToLower(strings.TrimSpace(name))]
}
