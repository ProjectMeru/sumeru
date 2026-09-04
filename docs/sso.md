# SSO / identity providers

Sumeru does **not** yet include built-in SAML or OIDC login.

## Current options

- Local `core.user` passwords (bcrypt) via UI `password_plain` / `orm.SetUserPassword`
- API keys (`sk_…`) for automation
- Admin **login-link notify** email (`ActionResetPassword`) — sends a login URL only; it does **not** issue a password-reset token

## Recommended enterprise path

1. Terminate SSO at a reverse proxy or identity-aware proxy (e.g. OAuth2 proxy) and map the authenticated identity into Sumeru in a future bridge module, **or**
2. Implement OIDC authorization-code login as a `base` extension when product prioritizes it.

Track status against the enterprise readiness DoD before removing the pre-alpha banner.
