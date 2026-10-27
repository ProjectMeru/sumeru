package web

import (
	"fmt"
	"net/http"
	"strings"

	"sumeru/core/server/config"
)

// publicBaseURL returns the configured public origin or derives it from a trusted request.
func publicBaseURL(r *http.Request) (string, error) {
	configured := strings.TrimRight(strings.TrimSpace(config.AppConfig.WebBaseURL), "/")
	if configured != "" {
		return configured, nil
	}
	peer := remoteAddrIP(r)
	if !isLoopbackIP(peer) && !trustedProxyContains(peer) {
		return "", fmt.Errorf("web_base_url required when setup is not from localhost or a trusted proxy")
	}
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https") {
		scheme = "https"
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return "", fmt.Errorf("missing host")
	}
	return scheme + "://" + host, nil
}

func oauthRedirectURI(r *http.Request) (string, error) {
	base, err := publicBaseURL(r)
	if err != nil {
		return "", err
	}
	return base + oauthCallbackRoute, nil
}
