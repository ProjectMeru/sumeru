# Authentication and SSO

Sumeru ships **OIDC/OAuth2 login** (authorization code + PKCE), **TOTP** second factor, and **trusted devices**. **SAML is not implemented** — use IdPs that speak OIDC.

## OIDC / OAuth providers

Configure under **Settings → Security → Authentication providers** (`sys.auth.provider`).

| Field / behavior | Notes |
| ---------------- | ----- |
| Routes | `GET /web/auth/oauth/start?provider=<id>`, `GET /web/auth/oauth/callback` |
| PKCE | S256 challenge on start; verifier in HttpOnly state cookie |
| Token exchange | Uses provider `token_url` or `issuer_url` + `/token` |
| `id_token` | When `jwks_url` is set, signature verified (RS256/ES256) before account link |
| JIT / link | Policies on provider control link-by-email vs deny; identities stored on `core.user.identity` |
| Login UI | Enabled providers appear as SSO buttons on `/web/login` |

Local password login remains available unless system parameter `auth.local_enabled` is `false` **and** at least one provider is enabled.

Secrets (`client_secret`, etc.) are sensitive — see `core/security/fields.go` registry.

## TOTP

- Users enroll under **Settings → Account security** (`totp_enabled`, `totp_secret` on `core.user`).
- After password or OAuth, login may redirect to **`/web/login/totp`** until a valid code is entered (pending-MFA cookie, HMAC-bound).
- **Trusted device** cookie (`sumeru_trusted_device`) skips TOTP for that browser when valid.

## Other auth surfaces

| Mechanism | Purpose |
| --------- | ------- |
| Session cookie | DB-backed `sys.session`; see [ops-runbook.md](ops-runbook.md) |
| API keys | `sk_…` for JSON-RPC automation |
| Login-link notify | Admin action sends login URL email — **not** a password-reset token |
| Portal | Separate portal session routes under `/portal/*` |

## Operations

- Rotate provider secrets in UI; disable provider before decommissioning IdP clients.
- Shared **`csrf_secret`** required when running multiple app instances ([ha-ops.md](ha-ops.md)).
- For air-gapped or proxy-only IdPs, terminate trust at the reverse proxy only if you accept that Sumeru will not see OIDC tokens (custom bridge addon).
