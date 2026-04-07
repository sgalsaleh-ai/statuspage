# Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser                              │
│                                                             │
│  ┌──────────────┐    ┌──────────────────────────────────┐   │
│  │  Admin UI     │    │  Public Status Page              │   │
│  │  (React)      │    │  (React)                         │   │
│  │               │    │                                  │   │
│  │  - Components │    │  - Overall status banner         │   │
│  │  - Incidents  │    │  - Component list + health       │   │
│  │  - Subscribers│    │  - Active incidents              │   │
│  │  - Dashboard  │    │  - Subscribe to updates          │   │
│  └──────┬───────┘    └─────────┬──────────┬─────────────┘   │
│         │ REST API             │ REST API │ WebSocket       │
└─────────┼──────────────────────┼──────────┼─────────────────┘
          │                      │          │
          ▼                      ▼          ▼
┌─────────────────────────┐   ┌────────────────────────┐
│      Go API Server      │   │      Centrifugo v5     │
│      (port 8080)        │   │      (port 8000/9000)  │
│                         │   │                        │
│  /api/status            │   │  WebSocket server      │
│  /api/incidents         │   │  (port 8000 external)  │
│  /api/admin/*  (auth)   │──▶│                        │
│  /api/subscribers       │   │  HTTP API              │
│  /api/centrifugo/config │   │  (port 9000 internal)  │
│  /healthz               │   │                        │
│                         │   │  Channel: "status"     │
│  Publishes events on    │   │                        │
│  component/incident     │   └────────────────────────┘
│  changes via HTTP API   │        Helm subchart #2
│                         │        
│  Sends emails via       │
│  Resend API             │
│                         │
└────────────┬────────────┘
             │
             ▼
┌────────────────────────┐   ┌────────────────────────┐
│    PostgreSQL 16       │   │      Resend API        │
│                        │   │      (external SaaS)   │
│  Tables:               │   │                        │
│  - users               │   │  Subscriber email      │
│  - components          │   │  notifications         │
│  - incidents           │   │  (license-gated)       │
│  - incident_updates    │   │                        │
│  - subscribers         │   └────────────────────────┘
│                        │
└────────────────────────┘
     Helm subchart #1
     (embedded / BYO)
```

## Data Flow

```
Admin changes component status
         │
         ▼
   Go API Server
    │         │
    │         ▼
    │    PostgreSQL
    │    (persists)
    │
    ▼
  Centrifugo
  (publish to "status" channel)
         │
         ▼
  Browser WebSocket
  (public page updates live)
```

## Helm Chart Structure

```
chart/statuspage/
├── Chart.yaml              # Dependencies: postgresql, centrifugo
├── values.yaml             # All configurable values
├── values.schema.json      # JSON schema for validation
├── charts/
│   ├── postgresql-18.5.15.tgz   # Bitnami PostgreSQL (subchart #1, stateful)
│   └── centrifugo-11.8.10.tgz   # Centrifugo v5 (subchart #2)
└── templates/
    ├── _helpers.tpl         # DB URL helpers, embedded/BYO logic for PostgreSQL
    ├── deployment.yaml      # App deployment + init container (DB wait)
    ├── service.yaml         # ClusterIP (configurable type)
    ├── secrets.yaml         # DB password, centrifugo API key, JWT secret
    ├── ingress.yaml         # Optional ingress
    └── certificate.yaml     # TLS: auto (cert-manager), manual, self-signed
```

## Ports

| Service     | Internal (cluster) | External (browser) | Notes                    |
|-------------|-------------------|--------------------|--------------------------|
| App         | 8080              | 9090 (port-fwd)    | REST API + static assets |
| Centrifugo  | 8000 (ws), 9000 (api) | 8000 (port-fwd) | WS for browsers, API for app |
| PostgreSQL  | 5432              | —                  | Internal only            |
