package auth

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

type jwksCacheEntry struct {
	keys    map[string]crypto.PublicKey
	expires time.Time
}

var (
	jwksCacheMu sync.Mutex
	jwksCache   = map[string]jwksCacheEntry{}
	jwksTTL     = 15 * time.Minute
)

// VerifyIDToken validates an OIDC id_token signature and standard claims when jwksURL is set.
func VerifyIDToken(ctx context.Context, idToken, issuer, clientID, jwksURL string) (map[string]interface{}, error) {
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return nil, errors.New("id_token required")
	}
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed id_token")
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("id_token header: %w", err)
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("id_token header json: %w", err)
	}
	alg := strings.ToUpper(strings.TrimSpace(header.Alg))
	if alg != "RS256" && alg != "ES256" {
		return nil, fmt.Errorf("unsupported jwt alg %q", header.Alg)
	}
	jwksURL = strings.TrimSpace(jwksURL)
	if jwksURL == "" {
		return nil, errors.New("jwks_url required for verification")
	}
	pub, err := publicKeyForJWKS(ctx, jwksURL, header.Kid, alg)
	if err != nil {
		return nil, err
	}
	signingInput := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("id_token signature: %w", err)
	}
	hash := crypto.SHA256.New()
	_, _ = hash.Write([]byte(signingInput))
	digest := hash.Sum(nil)
	switch key := pub.(type) {
	case *rsa.PublicKey:
		if alg != "RS256" {
			return nil, fmt.Errorf("alg/key mismatch")
		}
		if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest, sig); err != nil {
			return nil, fmt.Errorf("rsa verify: %w", err)
		}
	case *ecdsa.PublicKey:
		if alg != "ES256" {
			return nil, fmt.Errorf("alg/key mismatch")
		}
		if !ecdsa.VerifyASN1(key, digest, sig) {
			return nil, errors.New("ecdsa verify failed")
		}
	default:
		return nil, errors.New("unsupported public key type")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("id_token payload: %w", err)
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("id_token claims: %w", err)
	}
	if err := validateIDTokenClaims(claims, issuer, clientID); err != nil {
		return nil, err
	}
	return claims, nil
}

func validateIDTokenClaims(claims map[string]interface{}, issuer, clientID string) error {
	now := time.Now().Unix()
	if exp, ok := claimInt64(claims["exp"]); ok && exp > 0 && now > exp+60 {
		return errors.New("id_token expired")
	}
	issuer = strings.TrimSpace(issuer)
	if issuer != "" {
		if iss, _ := claims["iss"].(string); strings.TrimSpace(iss) != issuer {
			return fmt.Errorf("issuer mismatch")
		}
	}
	clientID = strings.TrimSpace(clientID)
	if clientID != "" {
		if !audienceContains(claims["aud"], clientID) {
			return fmt.Errorf("audience mismatch")
		}
	}
	if sub, _ := claims["sub"].(string); strings.TrimSpace(sub) == "" {
		return errors.New("missing sub")
	}
	return nil
}

func audienceContains(aud interface{}, clientID string) bool {
	switch v := aud.(type) {
	case string:
		return strings.TrimSpace(v) == clientID
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) == clientID {
				return true
			}
		}
	}
	return false
}

func claimInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	default:
		return 0, false
	}
}

func publicKeyForJWKS(ctx context.Context, jwksURL, kid, alg string) (crypto.PublicKey, error) {
	jwksCacheMu.Lock()
	if ent, ok := jwksCache[jwksURL]; ok && time.Now().Before(ent.expires) {
		if kid != "" {
			if k, ok := ent.keys[kid]; ok {
				jwksCacheMu.Unlock()
				return k, nil
			}
		} else if len(ent.keys) == 1 {
			for _, k := range ent.keys {
				jwksCacheMu.Unlock()
				return k, nil
			}
		}
	}
	jwksCacheMu.Unlock()

	keys, err := fetchJWKS(ctx, jwksURL)
	if err != nil {
		return nil, err
	}
	jwksCacheMu.Lock()
	jwksCache[jwksURL] = jwksCacheEntry{keys: keys, expires: time.Now().Add(jwksTTL)}
	jwksCacheMu.Unlock()

	if kid != "" {
		if k, ok := keys[kid]; ok {
			return k, nil
		}
		return nil, fmt.Errorf("jwks kid %q not found", kid)
	}
	if len(keys) == 1 {
		for _, k := range keys {
			return k, nil
		}
	}
	return nil, errors.New("jwks kid required")
}

func fetchJWKS(ctx context.Context, jwksURL string) (map[string]crypto.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("jwks http %d", resp.StatusCode)
	}
	var doc struct {
		Keys []json.RawMessage `json:"keys"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	out := make(map[string]crypto.PublicKey, len(doc.Keys))
	for _, raw := range doc.Keys {
		var meta struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
			Crv string `json:"crv"`
			X   string `json:"x"`
			Y   string `json:"y"`
		}
		if err := json.Unmarshal(raw, &meta); err != nil {
			continue
		}
		if meta.Use != "" && meta.Use != "sig" {
			continue
		}
		kid := meta.Kid
		if kid == "" {
			kid = "_"
		}
		switch meta.Kty {
		case "RSA":
			pub, err := rsaPublicFromJWK(meta.N, meta.E)
			if err != nil {
				continue
			}
			out[kid] = pub
		case "EC":
			pub, err := ecPublicFromJWK(meta.Crv, meta.X, meta.Y)
			if err != nil {
				continue
			}
			out[kid] = pub
		}
	}
	if 0 == len(out) {
		return nil, errors.New("no usable jwks keys")
	}
	return out, nil
}

func rsaPublicFromJWK(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}
	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}
	if 0 == eInt {
		eInt = 65537
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: eInt}, nil
}

func ecPublicFromJWK(crv, xB64, yB64 string) (*ecdsa.PublicKey, error) {
	if crv != "P-256" {
		return nil, fmt.Errorf("unsupported crv %q", crv)
	}
	xBytes, err := base64.RawURLEncoding.DecodeString(xB64)
	if err != nil {
		return nil, err
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(yB64)
	if err != nil {
		return nil, err
	}
	// ponytail: JWK EC coordinates via big.Int until stdlib exposes ParseUncompressed for OIDC JWKS.
	pub := &ecdsa.PublicKey{Curve: elliptic.P256()} //nolint:staticcheck // SA1019 coordinate fill for JWKS
	pub.X = new(big.Int).SetBytes(xBytes)           //nolint:staticcheck // SA1019
	pub.Y = new(big.Int).SetBytes(yBytes)           //nolint:staticcheck // SA1019
	return pub, nil
}

// ParseIDTokenClaimsUnsafe decodes id_token payload without signature verification (dev or post-verify).
func ParseIDTokenClaimsUnsafe(idToken string) (email, sub string, claims map[string]interface{}) {
	s := strings.TrimSpace(idToken)
	if s == "" || !strings.Contains(s, ".") {
		return "", "", nil
	}
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return "", "", nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", nil
	}
	if json.Unmarshal(payload, &claims) != nil {
		return "", "", nil
	}
	email, _ = claims["email"].(string)
	sub, _ = claims["sub"].(string)
	return email, sub, claims
}
