# Demo Recording Script

## Why This App? (say this)

So for bootcamp I built a status page app. It lets you manage infrastructure components, track incidents with a full timeline, and publish a live public status page that updates in real-time.

I picked this because it maps really cleanly to everything in the rubric. It needs a database, so PostgreSQL is our stateful subchart with embedded by default and a BYO toggle for external instances. It needs real-time updates, so Centrifugo is our second subchart handling WebSocket connections.

For the license-gated feature, we have subscriber notifications — when it's off, the status page is view-only. When it's on, people can subscribe and get email alerts when incidents happen. The app checks the SDK at runtime so you can flip it without redeploying.

It also gives us natural custom metrics like incident count, subscriber count, and uptime percentage. The health endpoint checks both the database and Centrifugo, which the support bundle uses to verify the app is healthy. For preflights, we validate connectivity to external PostgreSQL when BYO is configured, and check that the email service is reachable. And the config screen has real things to configure — embedded vs external database, email settings, feature toggles.

## Prerequisites

```bash
# Make sure you're in the repo root
cd /Users/salah/go/src/github.com/sgalsaleh-ai/statuspage

# Set Replicated env vars
export REPLICATED_API_TOKEN="d7c2c91d13b1d4744cea5b9dc0ade0818cc683d5d9a72e943db1653a6c8bcbce"
export REPLICATED_APP="statuspage"

# Make sure Node 22 is active
source ~/.nvm/nvm.sh && nvm use 22
```

## Part 1: Show the App Running Locally (Tier 0.1)

```bash
# Build and start with docker-compose
docker compose build app
docker compose up -d

# Verify health
curl -s http://localhost:8080/healthz | python3 -m json.tool
```

Open http://localhost:8080 in browser. Show:
- Public status page (empty initially)
- Go to /admin/login → redirects to /admin/setup
- Set admin password
- Show the admin dashboard

## Part 2: Show the Dockerfile (Tier 0.1)

```bash
cat Dockerfile
```

Talk through: multi-stage build — Node 22 builds React frontend, Go 1.25 builds backend, final image is minimal Alpine with both.

## Part 3: Helm Chart (Tier 0.2)

```bash
# Show chart structure
ls -la chart/statuspage/
cat chart/statuspage/Chart.yaml
cat chart/statuspage/values.schema.json | python3 -m json.tool | head -30

# Lint
helm lint ./chart/statuspage
```

## Part 4: Subcharts — Embedded and BYO (Tier 0.3)

```bash
# Show subcharts in Chart.yaml
grep -A 4 dependencies chart/statuspage/Chart.yaml
```

Show in values.yaml:
- `postgresql.enabled: true` (embedded by default)
- `externalPostgresql.*` fields for BYO
- `centrifugo.enabled: true` (embedded by default)
- Centrifugo is always embedded (no BYO needed)

## Part 5: Deploy to CMX Cluster

```bash
# Stop docker-compose first
docker compose down

# Update chart dependencies
helm dependency update ./chart/statuspage

# Create a Replicated release
replicated release create --promote Unstable --version "0.1.0+demo" -o json

# Create a CMX cluster
CLUSTER_ID=$(replicated cluster create \
  --distribution kind \
  --version 1.32.11 \
  --name "demo" \
  --ttl 5h \
  --wait 5m \
  -o json | jq -r '.id')
echo "Cluster ID: $CLUSTER_ID"

# Get kubeconfig
replicated cluster kubeconfig "$CLUSTER_ID" --stdout > /tmp/demo-kubeconfig
export KUBECONFIG=/tmp/demo-kubeconfig

# Login to Replicated registry
helm registry login registry.replicated.com \
  --username statuspage \
  --password "$REPLICATED_API_TOKEN"

# Install via Helm
helm install statuspage \
  oci://registry.replicated.com/statuspage/unstable/statuspage \
  --set image.repository=ghcr.io/sgalsaleh-ai/statuspage \
  --set image.tag=latest \
  --set app.centrifugoPublicURL=http://localhost:8000 \
  --wait --timeout 180s
```

## Part 6: Kubernetes Best Practices (Tier 0.4)

```bash
export KUBECONFIG=/tmp/demo-kubeconfig

# Show all pods running
kubectl get pods

# Show probes defined
kubectl get deployment statuspage -o yaml | grep -A 5 "livenessProbe\|readinessProbe"

# Show resource requests/limits
kubectl get deployment statuspage -o yaml | grep -A 4 "resources:" | head -10

# Show health endpoint (structured response)
kubectl exec deployment/statuspage -c statuspage -- wget -qO- http://localhost:8080/healthz
```

## Part 7: Show Init Container — DB Wait (Tier 0.6)

```bash
# Show init container in deployment
kubectl get deployment statuspage -o yaml | grep -A 10 "initContainers"
```

Talk through: busybox init container checks PostgreSQL is reachable before the main app starts. Prevents crash-loop when DB takes longer to start.

## Part 8: Port Forward and Demo Features (Tier 0.7)

```bash
# Port forward app and centrifugo
kubectl port-forward svc/statuspage 9090:8080 > /dev/null 2>&1 &
kubectl port-forward svc/statuspage-centrifugo 8000:8000 > /dev/null 2>&1 &
```

### Feature 1: Incident Lifecycle Management

Open http://localhost:9090

1. Go to admin → set password
2. Go to Components tab → add components:
   - "Web Application" (group: Application) — operational
   - "Database" (group: Infrastructure) — operational
   - "Real-time Messaging" (group: Infrastructure) — operational
   - "Email Notifications" (group: Integrations) — operational
3. Go to Incidents tab → create incident:
   - Title: "Elevated API Response Times"
   - Impact: Minor
   - Message: "We are investigating reports of slow API responses."
4. Click into the incident → post updates:
   - Status: Identified — "Root cause identified as a slow database query on the incidents endpoint."
   - Status: Monitoring — "Fix deployed, monitoring response times."
   - Status: Resolved — "Response times back to normal. Root cause was a missing database index."
5. Show the public page (/) — incident appears with full timeline

### Feature 2: Live Public Status Page

1. Open public page http://localhost:9090 in a **second browser tab**
2. In admin tab, go to Components → click the colored dots next to "Database" to change it to "degraded" (yellow)
3. Watch the public page update **in real-time without refreshing** — the component status changes live via Centrifugo WebSocket
4. Change it back to "operational" — again updates live

### Subscriber Flow

1. On the public page, enter an email in the subscribe form
2. Go to admin → Subscribers tab — show it listed

## Part 9: HTTPS Options (Tier 0.5)

Show in values.yaml the three TLS modes:

```bash
grep -A 10 "^tls:" chart/statuspage/values.yaml
```

- `auto` — cert-manager + Let's Encrypt
- `manual` — user provides TLS secret
- `selfSigned` — self-signed cert via cert-manager

Show the certificate.yaml template:

```bash
cat chart/statuspage/templates/certificate.yaml
```

## Part 10: CI/CD (Tier 1)

Show in browser:
- GitHub repo: https://github.com/sgalsaleh-ai/statuspage
- GHCR packages: https://github.com/sgalsaleh-ai/statuspage/pkgs/container/statuspage
- GitHub Actions: show a passing PR workflow and release workflow
- `.replicated` file that describes the app layout
- Vendor Portal: show the release in the Unstable channel

## Cleanup

```bash
# Stop port forwards
pkill -f "port-forward"

# Remove cluster (or let TTL expire)
replicated cluster rm "$CLUSTER_ID"
```
