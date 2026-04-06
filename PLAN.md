# StatusPage — Replicated Bootcamp

## Overview

A status page application (similar to Atlassian Statuspage) for the Replicated bootcamp.
Users manage infrastructure components, track incidents, and publish a live public status page.

## Tech Stack

- **Backend**: Go (JSON API)
- **Frontend**: React (Vite) — serves admin UI and public status page. Required for SDK integration in later tiers (license entitlement checks, update banners, support bundle button).
- **Database**: PostgreSQL (Bitnami Helm subchart) — primary data store for incidents, components, subscribers
- **Real-time**: Centrifugo v5 (Helm subchart) — pushes live status updates to browsers via WebSocket
- **Email**: Resend (external SaaS, free tier) — sends subscriber notification emails

## Architecture

```
Browser (admin, authenticated) ──► React App ──► Go API ──► PostgreSQL
                                                   │
                                                   ├──► Centrifugo (publish via POST /api, Authorization: apikey <key>)
                                                   │
                                                   └──► Resend API (send notification emails)

Browser (public, unauthenticated) ──► React App ──► Go API (read-only)
                                  ──► Centrifugo (WebSocket, client-side subscribe)
```

## Auth

- Admin pages require login (username/password against the Go API)
- Public status page is fully unauthenticated
- Admin links/nav are hidden from public pages

## Features

### Always available (base tier)
1. **Incident management** — create, update, resolve incidents with status timeline (investigating -> identified -> monitoring -> resolved)
2. **Component management** — define infrastructure components (API, website, database, etc.), set individual status (operational / degraded / major outage / maintenance)
3. **Public status page** — live page showing all component health and incident history, auto-updates via Centrifugo WebSocket
4. **Incident history** — public log of past incidents with full timeline

### Gated by license entitlement (subscriber notifications)
- "Subscribe to updates" button on public status page (hidden when disabled)
- Admin notification settings panel (locked when disabled)
- Subscriber list management
- Automatic email notifications on incident create/update via Resend
- In Tier 2: app queries Replicated SDK at runtime to check entitlement. For Tier 0: feature is always on (no SDK yet).

## Subcharts (Tier 0.3)

| Subchart | Purpose | Embedded by default | BYO opt-in |
|----------|---------|-------------------|------------|
| Bitnami PostgreSQL | Primary data store | Yes (StatefulSet) | Yes — provide external host/port/user/password/dbname |
| Centrifugo | Real-time WebSocket relay | Yes (Deployment) | Yes — provide external Centrifugo URL + API key |

Both subcharts are conditional via Helm values. When BYO is selected, the subchart pods are not deployed and the app connects to the external instance.

## Centrifugo v5 Notes

- Config key is `api_key` (not `http_api.key`)
- In docker-compose: use env vars (`CENTRIFUGO_API_KEY`), do NOT bind-mount config files (Docker Desktop for Mac issues)
- Go client: `Authorization: apikey <key>` header, POST to `/api` with `{"method":"publish","params":{...}}`
- Client-side subscriptions: set `CENTRIFUGO_ALLOW_SUBSCRIBE_FOR_CLIENT=true`

## Health Endpoint (Tier 0.4)

`GET /healthz` returns structured JSON:

```json
{
  "status": "ok",
  "checks": {
    "database": "ok",
    "centrifugo": "ok"
  }
}
```

Used by liveness/readiness probes and later by support bundle HTTP collector (Tier 3.3).

## HTTPS Options (Tier 0.5)

Three modes configured via Helm values:
1. **Auto** — cert-manager provisions a certificate via Let's Encrypt (requires cert-manager installed + ClusterIssuer)
2. **Manual** — user provides a TLS secret name containing their own cert/key
3. **Self-signed** — Helm generates a self-signed certificate (optional/bonus)

The customer provides their own domain (e.g. `status.acmecorp.com`). The vendor's custom domain (e.g. `registry.sgalsaleh.com`) is separate and used only for image proxying in Tier 2.2.

## Database Wait (Tier 0.6)

Init container in the app pod checks PostgreSQL is reachable before the main container starts. Prevents crash-loop when PostgreSQL takes longer to start than the app.

## Demoable Features (Tier 0.7)

1. **Incident lifecycle** — create an incident, post updates, resolve it, see the full timeline
2. **Live public status page** — open the page, change a component status in admin, see it update in real-time without page refresh (via Centrifugo)

## Configurable Values (for KOTS config screen in Tier 4/5)

These Helm values map to config screen items:
- **Embedded vs external PostgreSQL** — toggle + connection fields (host/port/user/password/dbname)
- **Embedded vs external Centrifugo** — toggle + connection fields (URL/API key)
- **Subscriber notifications** — enable/disable (gated by license entitlement in Tier 2)
- **Resend settings** — API key, from address (only relevant when notifications enabled)

## Forward Compatibility (Tiers 1–6)

Decisions made now to avoid rework later:

- **Containerization (Tier 1)**: Dockerfile is multi-stage, produces a minimal image suitable for CI builds. Go binary + React static assets in a single image.
- **Replicated SDK (Tier 2)**: React frontend will query the SDK at runtime for license entitlements (subscriber notifications) and update checks (banner). Go backend provides a proxy/passthrough endpoint if needed.
- **Custom metrics (Tier 2.4)**: App tracks incident count, component count, subscriber count — ready to send via SDK
- **Config screen (Tier 4/5)**: Helm values are structured to map cleanly to KOTS config items — embedded/BYO DB toggle, Resend settings, feature toggles
- **Support bundle (Tier 3)**: /healthz endpoint is reachable in-cluster for HTTP collector. App logs known failure patterns (DB connection failures) for textAnalyze. React UI will have a "Generate Support Bundle" button (Tier 3.7).
- **Preflight checks (Tier 3.1)**: External DB connectivity (when BYO), Resend API endpoint reachability, cluster resources, K8s version, distribution check
- **Air gap (Tier 4.3)**: All images will be proxied through Replicated's registry. Centrifugo is in-cluster so no external dependency for real-time features
- **Enterprise Portal (Tier 5)**: App name "StatusPage", branding-ready
