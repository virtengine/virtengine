# Frontend Security Audit

Date: 2026-01-29

Scope:
- Portal SDK (lib/portal)
- Portal deployment headers (pkg/provider_daemon/playbooks/deploy_portal.yaml)

Checklist:
- CSP present and hardened (no inline/eval, explicit sources)
- XSS prevention: React escaping + explicit sanitization for user inputs
- CSRF protection: token acquisition + header injection for session APIs
- Input sanitization: plain text, numeric, and JSON inputs
- OAuth2/OIDC flow: PKCE + state/nonce helpers
- Session management: httpOnly cookies, rotation, refresh handling
- Security headers: HSTS, Referrer-Policy, Permissions-Policy, etc.

Notes:
- OAuth/OIDC helper utilities are provided for state/nonce/PKCE generation and validation.
- Session APIs enforce CSRF headers and disable caching for sensitive calls.
- Inputs are sanitized at submission/verification boundaries to prevent injection.
