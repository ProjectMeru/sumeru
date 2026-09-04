package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"sumeru/core/applog"
	"sumeru/core/event"
	"sumeru/core/metrics"
)

func dispatchWebhook(ctx context.Context, rawURL string, ev event.Event) error {
	if err := validateWebhookURL(rawURL); err != nil {
		metrics.Inc("sumeru_webhook_blocked_total")
		applog.Warn(ctx, applog.Event{
			Message:   "webhook URL rejected",
			Component: "automation",
			Operation: "webhook",
			Status:    "blocked",
			Context:   map[string]interface{}{"url": rawURL},
			Err:       err,
		})
		return err
	}
	body, err := json.Marshal(map[string]interface{}{
		"event":   ev.Name,
		"actor":   ev.Actor,
		"payload": ev.Payload,
	})
	if err != nil {
		return err
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	dialIPs, err := resolveWebhookDialIPs(u.Hostname())
	if err != nil {
		metrics.Inc("sumeru_webhook_blocked_total")
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			DialContext: pinnedWebhookDialer(dialIPs),
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("webhook redirect limit exceeded")
			}
			if err := validateWebhookURL(req.URL.String()); err != nil {
				return err
			}
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		applog.Warn(ctx, applog.Event{
			Message:   "webhook server action failed",
			Component: "automation",
			Operation: "webhook",
			Status:    "failed",
			Context:   map[string]interface{}{"url": rawURL, "status": resp.StatusCode},
		})
	}
	return nil
}

func pinnedWebhookDialer(allowed []net.IP) func(ctx context.Context, network, addr string) (net.Conn, error) {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, ip := range allowed {
		allowedSet[ip.String()] = struct{}{}
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		ip := net.ParseIP(host)
		if ip == nil {
			return nil, fmt.Errorf("webhook dial host must be IP")
		}
		if _, ok := allowedSet[ip.String()]; !ok {
			return nil, fmt.Errorf("webhook dial IP not in validated set")
		}
		if blockedWebhookIP(ip) {
			return nil, fmt.Errorf("webhook dial IP not allowed")
		}
		var d net.Dialer
		return d.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
	}
}

func resolveWebhookDialIPs(host string) ([]net.IP, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, fmt.Errorf("webhook host required")
	}
	if ip := net.ParseIP(host); ip != nil {
		if blockedWebhookIP(ip) {
			return nil, fmt.Errorf("webhook IP not allowed")
		}
		return []net.IP{ip}, nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("webhook host lookup: %w", err)
	}
	var out []net.IP
	for _, ip := range ips {
		if blockedWebhookIP(ip) {
			return nil, fmt.Errorf("webhook host resolves to blocked address")
		}
		out = append(out, ip)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("webhook host resolved to no addresses")
	}
	return out, nil
}

func validateWebhookURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("empty webhook url")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid webhook url: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "https" && scheme != "http" {
		return fmt.Errorf("webhook scheme %q not allowed", u.Scheme)
	}
	host := strings.TrimSpace(u.Hostname())
	if host == "" {
		return fmt.Errorf("webhook host required")
	}
	lowerHost := strings.ToLower(host)
	if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".localhost") || lowerHost == "metadata.google.internal" {
		return fmt.Errorf("webhook host not allowed")
	}
	_, err = resolveWebhookDialIPs(host)
	return err
}

func blockedWebhookIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast() {
		return true
	}
	// CGNAT / shared address space (RFC 6598) — not covered by IsPrivate().
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return true
		}
	}
	return false
}
