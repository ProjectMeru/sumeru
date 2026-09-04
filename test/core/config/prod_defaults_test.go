package config_test

import (
	"testing"

	"sumeru/core/server/config"
)

func TestApplyProductionSecurityDefaults_setsRateLimitWhenNotDev(t *testing.T) {
	c := &config.Config{DevMode: false, RateLimitRPM: 0}
	warns := config.ApplyProductionSecurityDefaults(c)
	if c.RateLimitRPM != 120 {
		t.Fatalf("RateLimitRPM=%d want 120", c.RateLimitRPM)
	}
	if len(warns) == 0 {
		t.Fatal("expected warning about defaulted rate limit")
	}
}

func TestApplyProductionSecurityDefaults_keepsDevUnlimited(t *testing.T) {
	c := &config.Config{DevMode: true, RateLimitRPM: 0}
	warns := config.ApplyProductionSecurityDefaults(c)
	if c.RateLimitRPM != 0 {
		t.Fatalf("dev RateLimitRPM=%d want 0", c.RateLimitRPM)
	}
	found := false
	for _, w := range warns {
		if w != "" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected dev_mode warnings")
	}
}
