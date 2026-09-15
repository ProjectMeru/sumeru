package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sumeru/core/server/config"
)

const shareTokenTTL = 7 * 24 * time.Hour

func shareTokenSecret() []byte {
	secret := strings.TrimSpace(config.AppConfig.CSRFSecret)
	if secret == "" {
		secret = "dev-share-token"
	}
	return []byte(secret)
}

// MintShareToken builds a signed portal share token for model/id (issuerUID must match validator context).
func MintShareToken(issuerUID int, model string, resID int64) (string, error) {
	exp := time.Now().UTC().Add(shareTokenTTL).Unix()
	return mintShareTokenAt(issuerUID, model, resID, exp)
}

func mintShareTokenAt(issuerUID int, model string, resID int64, exp int64) (string, error) {
	if issuerUID <= 0 || resID <= 0 || strings.TrimSpace(model) == "" || exp <= 0 {
		return "", fmt.Errorf("invalid share token input")
	}
	payload := fmt.Sprintf("%d|%s|%d|%d", issuerUID, model, resID, exp)
	mac := hmac.New(sha256.New, shareTokenSecret())
	_, _ = mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	body := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return body + "." + sig, nil
}

type parsedShareToken struct {
	IssuerUID int
	Model     string
	ResID     int64
	Expires   int64
}

func parseShareToken(token string) (parsedShareToken, error) {
	var out parsedShareToken
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 2 {
		return out, fmt.Errorf("invalid token")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return out, fmt.Errorf("invalid token payload")
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return out, fmt.Errorf("invalid token signature")
	}
	payload := string(payloadBytes)
	mac := hmac.New(sha256.New, shareTokenSecret())
	_, _ = mac.Write([]byte(payload))
	if !hmac.Equal(sigBytes, mac.Sum(nil)) {
		return out, fmt.Errorf("invalid token signature")
	}
	segments := strings.Split(payload, "|")
	if len(segments) != 4 {
		return out, fmt.Errorf("invalid token payload")
	}
	uid, err := strconv.Atoi(segments[0])
	if err != nil || uid <= 0 {
		return out, fmt.Errorf("invalid issuer")
	}
	resID, err := strconv.ParseInt(segments[2], 10, 64)
	if err != nil || resID <= 0 {
		return out, fmt.Errorf("invalid record id")
	}
	exp, err := strconv.ParseInt(segments[3], 10, 64)
	if err != nil || exp <= 0 {
		return out, fmt.Errorf("invalid expiry")
	}
	if time.Now().UTC().Unix() > exp {
		return out, fmt.Errorf("token expired")
	}
	out.IssuerUID = uid
	out.Model = segments[1]
	out.ResID = resID
	out.Expires = exp
	return out, nil
}
