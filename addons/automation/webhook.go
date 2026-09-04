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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("webhook redirect limit exceeded")
			}
			return validateWebhookURL(req.URL.String())
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
	if ip := net.ParseIP(host); ip != nil {
		if blockedWebhookIP(ip) {
			return fmt.Errorf("webhook IP not allowed")
		}
		return nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("webhook host lookup: %w", err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("webhook host resolved to no addresses")
	}
	for _, ip := range ips {
		if blockedWebhookIP(ip) {
			return fmt.Errorf("webhook host resolves to blocked address")
		}
	}
	return nil
}

func blockedWebhookIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast()
}
