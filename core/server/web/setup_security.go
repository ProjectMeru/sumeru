package web

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"sumeru/core/orm"
	"sumeru/core/server/config"
)

const setupRateLimitWindow = time.Minute

var setupRateLimiter = struct {
	sync.Mutex
	attemptsByIP map[string][]time.Time
}{
	attemptsByIP: make(map[string][]time.Time),
}

// allowSetupRequest enforces localhost-only (when configured), optional setup token, and rate limits.
func allowSetupRequest(w http.ResponseWriter, r *http.Request, tokenFromBody string) bool {
	if !requireSetupEnvironment(w, r) {
		return false
	}
	if !validateSetupToken(w, r, tokenFromBody) {
		return false
	}
	return allowSetupRateLimit(w, clientIP(r))
}

func requireSetupEnvironment(w http.ResponseWriter, r *http.Request) bool {
	if orm.IsInitialized() {
		http.Error(w, "Setup already completed", http.StatusForbidden)
		return false
	}
	if config.AppConfig.SetupLocalhostOnly && !isLoopbackIP(clientIP(r)) {
		http.Error(w, "Setup is restricted to localhost", http.StatusForbidden)
		return false
	}
	return true
}

func validateSetupToken(w http.ResponseWriter, r *http.Request, tokenFromBody string) bool {
	expectedToken := strings.TrimSpace(config.AppConfig.SetupToken)
	if expectedToken == "" {
		return true
	}
	providedToken := setupTokenFromRequest(r, tokenFromBody)
	if providedToken == expectedToken {
		return true
	}
	http.Error(w, "Invalid setup token", http.StatusForbidden)
	return false
}

func setupTokenFromRequest(r *http.Request, tokenFromBody string) string {
	if headerToken := strings.TrimSpace(r.Header.Get(setupTokenHeader)); headerToken != "" {
		return headerToken
	}
	return strings.TrimSpace(tokenFromBody)
}

func allowSetupRateLimit(w http.ResponseWriter, requestIP string) bool {
	now := time.Now()
	setupRateLimiter.Lock()
	defer setupRateLimiter.Unlock()

	recentAttempts := pruneSetupAttempts(setupRateLimiter.attemptsByIP[requestIP], now)
	if len(recentAttempts) >= setupRateLimitMax {
		setupRateLimiter.attemptsByIP[requestIP] = recentAttempts
		http.Error(w, "Too many setup attempts", http.StatusTooManyRequests)
		return false
	}

	setupRateLimiter.attemptsByIP[requestIP] = append(recentAttempts, now)
	return true
}

func pruneSetupAttempts(attempts []time.Time, now time.Time) []time.Time {
	recentAttempts := make([]time.Time, 0, len(attempts))
	for _, attemptTime := range attempts {
		if now.Sub(attemptTime) <= setupRateLimitWindow {
			recentAttempts = append(recentAttempts, attemptTime)
		}
	}
	return recentAttempts
}

func clientIP(r *http.Request) string {
	remote := remoteAddrIP(r)
	if remote != "" && trustedProxyContains(remote) {
		if forwardedFor := strings.TrimSpace(r.Header.Get(forwardedForHeader)); forwardedFor != "" {
			if commaIndex := strings.Index(forwardedFor, ","); commaIndex >= 0 {
				return strings.TrimSpace(forwardedFor[:commaIndex])
			}
			return forwardedFor
		}
	}
	return remote
}

func remoteAddrIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}

func trustedProxyContains(ipAddress string) bool {
	parsedIP := net.ParseIP(strings.TrimSpace(ipAddress))
	if parsedIP == nil {
		return false
	}
	for _, network := range trustedProxyNetworks() {
		if network.Contains(parsedIP) {
			return true
		}
	}
	return false
}

func trustedProxyNetworks() []*net.IPNet {
	raw := strings.TrimSpace(config.AppConfig.TrustedProxies)
	if raw == "" {
		return nil
	}
	var networks []*net.IPNet
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if !strings.Contains(part, "/") {
			if ip := net.ParseIP(part); ip != nil {
				bits := 32
				if ip.To4() == nil {
					bits = 128
				}
				part = fmt.Sprintf("%s/%d", part, bits)
			}
		}
		_, network, err := net.ParseCIDR(part)
		if err != nil {
			continue
		}
		networks = append(networks, network)
	}
	return networks
}

func isLoopbackIP(ipAddress string) bool {
	parsedIP := net.ParseIP(strings.TrimSpace(ipAddress))
	return parsedIP != nil && parsedIP.IsLoopback()
}
