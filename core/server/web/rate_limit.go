package web

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"sumeru/core/server/api"
	"sumeru/core/server/config"
)

type rateBucket struct {
	count  int
	window time.Time
}

var (
	rateMu       sync.Mutex
	rateByIP     = map[string]*rateBucket{}
	rateByUID    = map[string]*rateBucket{}
	rateLimitOn  bool
	rateLimitRPM int
)

// InitRateLimit reads rate_limit_rpm from loaded config.
func InitRateLimit() {
	rateLimitRPM = config.AppConfig.RateLimitRPM
	rateLimitOn = rateLimitRPM > 0
}

func rateLimitedPath(path string) bool {
	if strings.HasPrefix(path, apiRPCRoutePrefix) {
		return true
	}
	switch path {
	case loginRoute, totpLoginRoute, oauthStartRoute, oauthCallbackRoute:
		return true
	default:
		return false
	}
}

func allowRateBucket(buckets map[string]*rateBucket, key string) bool {
	key = strings.TrimSpace(key)
	if key == "" {
		key = "unknown"
	}
	now := time.Now()
	b, ok := buckets[key]
	if !ok || now.Sub(b.window) >= time.Minute {
		buckets[key] = &rateBucket{count: 1, window: now}
		return true
	}
	if b.count >= rateLimitRPM {
		return false
	}
	b.count++
	return true
}

func allowRate(clientIP string, uid int) bool {
	if !rateLimitOn {
		return true
	}
	rateMu.Lock()
	defer rateMu.Unlock()
	if !allowRateBucket(rateByIP, clientIP) {
		return false
	}
	if uid > 0 {
		return allowRateBucket(rateByUID, strconv.Itoa(uid))
	}
	return true
}

func writeRateLimitResponse(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, apiRPCRoutePrefix) {
		api.WriteResponse(w, http.StatusTooManyRequests, api.Fail(api.CodeRateLimited, "Too many requests", nil))
		return
	}
	http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
}

func enforceRateLimit(w http.ResponseWriter, r *http.Request) bool {
	if !rateLimitedPath(r.URL.Path) {
		return true
	}
	uid := SessionUserID(r)
	if allowRate(clientIP(r), uid) {
		return true
	}
	writeRateLimitResponse(w, r)
	return false
}

// ResetRateLimitForTest clears in-memory rate limit buckets (tests only).
func ResetRateLimitForTest() {
	rateMu.Lock()
	defer rateMu.Unlock()
	rateByIP = map[string]*rateBucket{}
	rateByUID = map[string]*rateBucket{}
}

// SetRateLimitRPMForTest configures RPM and enables limiting (tests only).
func SetRateLimitRPMForTest(rpm int) {
	rateLimitRPM = rpm
	rateLimitOn = rpm > 0
}

// EnforceRateLimitForTest exposes enforceRateLimit for handler tests.
func EnforceRateLimitForTest(w http.ResponseWriter, r *http.Request) bool {
	return enforceRateLimit(w, r)
}

// pruneAttemptsAfter keeps timestamps strictly after cutoff (in-place slice reuse).
func pruneAttemptsAfter(attempts []time.Time, cutoff time.Time) []time.Time {
	out := attempts[:0]
	for _, attemptTime := range attempts {
		if attemptTime.After(cutoff) {
			out = append(out, attemptTime)
		}
	}
	return out
}

// pruneAttemptsWithin keeps timestamps within window ending at now.
func pruneAttemptsWithin(attempts []time.Time, now time.Time, window time.Duration) []time.Time {
	cutoff := now.Add(-window)
	return pruneAttemptsAfter(attempts, cutoff)
}
