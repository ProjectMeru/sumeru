package web

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"sumeru/core/orm"
	"sumeru/core/server/auth"
	"sumeru/core/server/config"
)

const (
	oauthStartRoute    = "/web/auth/oauth/start"
	oauthCallbackRoute = "/web/auth/oauth/callback"
	oauthStateCookie   = "sumeru_oauth_state"
)

func registerOAuthRoutes() {
	registerPublic(http.MethodGet, oauthStartRoute, OAuthStartHandler)
	registerPublic(http.MethodGet, oauthCallbackRoute, OAuthCallbackHandler)
}

func OAuthStartHandler(w http.ResponseWriter, r *http.Request) {
	providerID := parsePositiveInt(r.URL.Query().Get("provider"))
	if 0 >= providerID {
		http.Error(w, "provider required", http.StatusBadRequest)
		return
	}
	ctx := orm.AuditedBypass(r.Context(), "oauth.start")
	prov, err := orm.SearchOne(ctx, "sys.auth.provider", map[string]interface{}{"id": providerID, "enabled": true})
	if err != nil || prov == nil {
		http.Error(w, "provider not found", http.StatusNotFound)
		return
	}
	state := randomURLSafe(32)
	verifier := randomURLSafe(64)
	challenge := pkceChallenge(verifier)
	setOAuthStateCookie(w, state, verifier, providerID)
	authURL := strings.TrimSpace(oauthRowString(prov, "authorize_url"))
	if authURL == "" {
		authURL = strings.TrimRight(strings.TrimSpace(oauthRowString(prov, "issuer_url")), "/") + "/authorize"
	}
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", oauthRowString(prov, "client_id"))
	redirectURI, err := oauthRedirectURI(r)
	if err != nil {
		redirectLoginOAuthError(w, r)
		return
	}
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", defaultScope(oauthRowString(prov, "scopes")))
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	sep := "?"
	if strings.Contains(authURL, "?") {
		sep = "&"
	}
	http.Redirect(w, r, authURL+sep+q.Encode(), http.StatusFound)
}

func OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	providerID, verifier, ok := validateOAuthStateCookie(r, state)
	if !ok || code == "" {
		redirectLoginOAuthError(w, r)
		return
	}
	clearOAuthStateCookie(w)
	ctx := r.Context()
	prov, err := orm.SearchOne(orm.AuditedBypass(ctx, "oauth.callback"), "sys.auth.provider", map[string]interface{}{"id": providerID})
	if err != nil || prov == nil {
		redirectLoginOAuthError(w, r)
		return
	}
	tokenURL := strings.TrimSpace(oauthRowString(prov, "token_url"))
	if tokenURL == "" {
		tokenURL = strings.TrimRight(strings.TrimSpace(oauthRowString(prov, "issuer_url")), "/") + "/token"
	}
	redirectURI, err := oauthRedirectURI(r)
	if err != nil {
		redirectLoginOAuthError(w, r)
		return
	}
	tok, err := exchangeOAuthCode(r, tokenURL, oauthRowString(prov, "client_id"), oauthRowString(prov, "client_secret"), code, verifier, redirectURI)
	if err != nil {
		redirectLoginOAuthError(w, r)
		return
	}
	email, sub, emailVerified, err := resolveOAuthIdentity(r.Context(), prov, tok)
	if err != nil {
		redirectLoginOAuthError(w, r)
		return
	}
	if err := validateOAuthIdentityClaims(prov, tok, sub); err != nil {
		redirectLoginOAuthError(w, r)
		return
	}
	userID, err := linkOAuthUser(ctx, providerID, sub, email, emailVerified, oauthRowString(prov, "link_policy"))
	if err != nil || 0 >= userID {
		redirectLoginOAuthError(w, r)
		return
	}
	completePasswordOrOAuthLogin(w, r, loginFinishOpts{
		UserID:   userID,
		Next:     resolveLoginNext(r),
		LogRoute: oauthCallbackRoute,
	})
}

func exchangeOAuthCode(r *http.Request, tokenURL, clientID, clientSecret, code, verifier, redirectURI string) (map[string]interface{}, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", clientID)
	form.Set("code_verifier", verifier)
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("token http %d", resp.StatusCode)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func linkOAuthUser(ctx context.Context, providerID int, subject, email string, emailVerified bool, policy string) (int, error) {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return 0, fmt.Errorf("missing subject")
	}
	if rec, err := orm.SearchOne(ctx, "core.user.identity", map[string]interface{}{
		"provider_id": providerID,
		"subject":     subject,
	}); err == nil && rec != nil {
		return int(oauthRowInt(rec, "user_id")), nil
	}
	email = strings.TrimSpace(strings.ToLower(email))
	if email != "" && emailVerified {
		if user, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"login": email}); err == nil && user != nil {
			uid := int(oauthRowInt(user, "id"))
			_ = upsertIdentity(ctx, providerID, subject, uid)
			return uid, nil
		}
	}
	switch strings.TrimSpace(policy) {
	case "jit_internal", "jit_portal":
		// ponytail: minimal JIT — create inactive-safe internal user with login=email or subject.
		login := email
		if login == "" {
			login = "oidc_" + subject
		}
		inst := orm.Registry["core.user"]
		uid, err := orm.Create(ctx, inst, map[string]interface{}{
			"login":     login,
			"name":      login,
			"email":     email,
			"active":    true,
			"user_type": map[bool]string{true: "portal", false: "internal"}[policy == "jit_portal"],
		})
		if err != nil {
			return 0, err
		}
		_ = upsertIdentity(ctx, providerID, subject, uid)
		return uid, nil
	default:
		return 0, fmt.Errorf("unknown user")
	}
}

func upsertIdentity(ctx context.Context, providerID int, subject string, uid int) error {
	inst := orm.Registry["core.user.identity"]
	_, err := orm.Create(ctx, inst, map[string]interface{}{
		"provider_id": providerID,
		"subject":     subject,
		"user_id":     uid,
	})
	return err
}

func validateOAuthIdentityClaims(prov map[string]interface{}, tok map[string]interface{}, sub string) error {
	sub = strings.TrimSpace(sub)
	if sub == "" {
		return fmt.Errorf("missing subject")
	}
	if !config.AppConfig.DevMode {
		if strings.TrimSpace(oauthRowString(prov, "jwks_url")) != "" {
			idTok, _ := tok["id_token"].(string)
			if strings.TrimSpace(idTok) == "" {
				return fmt.Errorf("id_token required")
			}
		}
	}
	return nil
}

func resolveOAuthIdentity(ctx context.Context, prov map[string]interface{}, tok map[string]interface{}) (email, sub string, emailVerified bool, err error) {
	idTok, _ := tok["id_token"].(string)
	jwksURL := strings.TrimSpace(oauthRowString(prov, "jwks_url"))
	issuer := strings.TrimSpace(oauthRowString(prov, "issuer_url"))
	clientID := strings.TrimSpace(oauthRowString(prov, "client_id"))
	idTok = strings.TrimSpace(idTok)
	if idTok != "" && jwksURL != "" {
		claims, verr := auth.VerifyIDToken(ctx, idTok, issuer, clientID, jwksURL)
		if verr != nil {
			return "", "", false, verr
		}
		email, _ = claims["email"].(string)
		sub, _ = claims["sub"].(string)
		return email, sub, oauthEmailVerified(claims), nil
	}
	if !config.AppConfig.DevMode && (idTok != "" || jwksURL != "") {
		return "", "", false, fmt.Errorf("oidc verification required")
	}
	var claims map[string]interface{}
	if idTok != "" {
		email, sub, claims = auth.ParseIDTokenClaimsUnsafe(idTok)
		emailVerified = oauthEmailVerified(claims)
	}
	if sub == "" {
		sub, _ = tok["sub"].(string)
	}
	if email == "" {
		email, _ = tok["email"].(string)
	}
	if sub == "" && strings.TrimSpace(oauthRowString(prov, "provider_type")) == "oauth2" {
		if access, _ := tok["access_token"].(string); strings.TrimSpace(access) != "" {
			email, sub = fetchOAuthUserInfo(ctx, access, issuer)
		}
	}
	if sub == "" {
		return "", "", false, fmt.Errorf("missing subject")
	}
	return email, sub, emailVerified, nil
}

func oauthEmailVerified(claims map[string]interface{}) bool {
	if claims == nil {
		return false
	}
	switch v := claims["email_verified"].(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(strings.TrimSpace(v), "true")
	default:
		return false
	}
}

func fetchOAuthUserInfo(ctx context.Context, accessToken, issuer string) (email, sub string) {
	base := strings.TrimRight(strings.TrimSpace(issuer), "/")
	if base == "" {
		return "", ""
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/userinfo", nil)
	if err != nil {
		return "", ""
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 400 {
		return "", ""
	}
	var claims map[string]interface{}
	if json.Unmarshal(body, &claims) != nil {
		return "", ""
	}
	email, _ = claims["email"].(string)
	sub, _ = claims["sub"].(string)
	return email, sub
}

func redirectLoginOAuthError(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, loginRoute+"?"+flashMessageParam+"="+url.QueryEscape(oauthDeniedMsg), http.StatusFound)
}

type loginAuthProvider struct {
	ID    int
	Label string
}

func listEnabledAuthProviders(ctx context.Context) []loginAuthProvider {
	ctx = orm.AuditedBypass(ctx, "login.providers")
	rows, err := orm.Search(ctx, "sys.auth.provider", [][]interface{}{{"enabled", "=", true}})
	if err != nil {
		return nil
	}
	if 0 == len(rows) {
		return nil
	}

	out := make([]loginAuthProvider, 0, len(rows))
	for _, row := range rows {
		id := int(oauthRowInt(row, "id"))
		if 0 >= id {
			continue
		}
		label := strings.TrimSpace(oauthRowString(row, "button_label"))
		if label == "" {
			label = strings.TrimSpace(oauthRowString(row, "name"))
		}
		if label == "" {
			label = "Sign in"
		}
		out = append(out, loginAuthProvider{ID: id, Label: label})
	}
	return out
}

func authLocalEnabled(ctx context.Context) bool {
	raw := strings.TrimSpace(strings.ToLower(orm.GetConfig(orm.AuditedBypass(ctx, "login.config"), authLocalConfigKey, "true")))
	switch raw {
	case "0", "false", "no", "off":
		providers := listEnabledAuthProviders(ctx)
		return 0 == len(providers)
	default:
		return true
	}
}

func loginPageCompanyName(ctx context.Context) string {
	ctx = orm.AuditedBypass(ctx, "login.brand")
	rows, err := orm.SearchLimit(ctx, "core.company", nil, 1)
	if err != nil {
		return ""
	}
	if 0 == len(rows) {
		return ""
	}

	return strings.TrimSpace(oauthRowString(rows[0], "name"))
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func randomURLSafe(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func defaultScope(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "openid email profile"
	}
	return s
}

func oauthRowString(rec map[string]interface{}, key string) string {
	if v, ok := rec[key].(string); ok {
		return v
	}
	return ""
}

func oauthRowInt(rec map[string]interface{}, key string) int64 {
	switch v := rec[key].(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	default:
		return 0
	}
}

// OAuth state cookie helpers (state|verifier|providerID base64 json).
func setOAuthStateCookie(w http.ResponseWriter, state, verifier string, providerID int) {
	payload, _ := json.Marshal(map[string]interface{}{
		"state": state, "verifier": verifier, "provider_id": providerID, "exp": time.Now().Add(10 * time.Minute).Unix(),
	})
	setNamedCookie(w, oauthStateCookie, base64.RawURLEncoding.EncodeToString(payload), "/", 600, sessionSameSite())
}

func validateOAuthStateCookie(r *http.Request, state string) (providerID int, verifier string, ok bool) {
	c, err := r.Cookie(oauthStateCookie)
	if err != nil || c.Value == "" {
		return 0, "", false
	}
	raw, err := base64.RawURLEncoding.DecodeString(c.Value)
	if err != nil {
		return 0, "", false
	}
	var m map[string]interface{}
	if json.Unmarshal(raw, &m) != nil {
		return 0, "", false
	}
	if fmt.Sprint(m["state"]) != state {
		return 0, "", false
	}
	providerID = int(oauthRowInt(m, "provider_id"))
	verifier, _ = m["verifier"].(string)
	return providerID, verifier, providerID > 0 && verifier != ""
}

func clearOAuthStateCookie(w http.ResponseWriter) {
	clearNamedCookie(w, oauthStateCookie, "/", sessionSameSite())
}
