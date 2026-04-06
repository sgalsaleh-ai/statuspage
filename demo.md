# Demo Script

## 1. Setup

First visit: go to /admin/login, it redirects to /admin/setup. Set your admin password.

## 2. Create Components

| Name | Group | Description |
|------|-------|-------------|
| Web Application | Application | Core application serving the admin dashboard and public status page |
| Database | Infrastructure | PostgreSQL database for persistent storage of incidents, components, and subscribers |
| Real-time Messaging | Infrastructure | Centrifugo WebSocket server for live status updates |
| Email Notifications | Integrations | Resend email service for subscriber incident notifications |

## 3. Create Incidents

### Incident 1: Walk through full lifecycle

- Title: "Elevated API Response Times"
- Impact: Minor
- Initial message: "We are investigating reports of slow API responses."

Post updates (one at a time):
1. Status: Identified — "Root cause identified as a slow database query on the incidents endpoint."
2. Status: Monitoring — "Fix deployed, monitoring response times."
3. Status: Resolved — "Response times back to normal. Root cause was a missing database index."

### Incident 2: Leave active to show on public page

- Title: "Email Notification Delays"
- Impact: Minor
- Initial message: "Some subscribers are experiencing delayed incident notification emails."

## 4. Demo Real-time Updates

Open the public page (/) in a second browser tab. In admin, change a component status (e.g. set Database to "degraded"). Watch the public page update live without refresh.

## 5. Demo Subscriber Flow

On the public page, enter an email in the subscribe form. Go to admin > Subscribers to see it listed.

## 6. Demo Data Persistence

Delete the app pod (or `docker compose restart app`). Verify all components, incidents, and subscribers are still there.

## 7. Health Endpoint

Visit /healthz to show structured JSON response with database and centrifugo checks.
