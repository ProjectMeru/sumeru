package orm

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ValidatePasswordPolicy checks length (and basic complexity) against config / defaults.
func ValidatePasswordPolicy(plain string) error {
	minLen := 8
	if s := GetConfig(BackgroundBypass(), "auth.password_min_length", "8"); s != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && n > 0 {
			minLen = n
		}
	}
	if len(plain) < minLen {
		return fmt.Errorf("password must be at least %d characters", minLen)
	}
	requireComplex := strings.EqualFold(GetConfig(BackgroundBypass(), "auth.password_complexity", "0"), "1")
	if !requireComplex {
		return nil
	}
	var upper, lower, digit bool
	for _, r := range plain {
		switch {
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsDigit(r):
			digit = true
		}
	}
	if !upper || !lower || !digit {
		return fmt.Errorf("password must include upper, lower, and digit characters")
	}
	return nil
}

// BackgroundBypass returns a background context with security bypass for config reads.
func BackgroundBypass() context.Context {
	return context.WithValue(context.Background(), bypassKey, true)
}
