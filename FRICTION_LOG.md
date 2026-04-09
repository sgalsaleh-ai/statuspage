# Replicated Product Friction Log

Issues encountered while integrating Replicated features during the bootcamp.

---

## Replicated CLI

### No dedicated .replicated file docs
- The `.replicated` file format is only documented in `replicated release create --help` output
- No reference page at docs.replicated.com
- Had to discover the format by trial and error and reading CLI help

---

## Replicated SDK

### Support bundle upload endpoint path wrong in docs
- Docs say `POST /api/v1/app/supportbundle`
- Actual SDK endpoint is `POST /api/v1/supportbundle` (no `/app/`)
- Had to read the SDK source code to find the correct path

### SDK subchart alias breaks license injection
- Using `alias: statuspage-sdk` in Chart.yaml causes license values to be injected under `replicated.*` but the SDK looks for them under `statuspage-sdk.*`
- The aliased subchart never receives its license data and crash-loops
- Fix: use `nameOverride` instead of `alias` for branding
- This is a common pattern that should be documented

---

## Replicated Proxy Registry

### No auth = 400 Bad Request (not 401)
- Pulling from proxy without authentication returns `400 Bad Request` instead of `401 Unauthorized`
- The error message says "failed to fetch anonymous token" which is confusing
- A 401 with "authentication required" would be clearer

### External registry must be added before proxy works
- Docker Hub images fail with "Unknown hostname" even though they're public
- Need to add Docker Hub as an external registry in Vendor Portal first
- Not obvious that public registries need explicit vendor portal configuration

---

## Helm Chart / HelmChart CR

### Helm release secret too large
- When a leftover `.tgz` was in the chart directory, the chart packaged itself recursively
- The resulting Helm release secret exceeded 1MB Kubernetes limit
- No clear error pointing to the recursive packaging — just "Too long"

---

## Preflight Specs: EC v3 vs Native Helm

### Conflicting requirements between install methods
- **Native Helm** requires the preflight spec inside the Helm chart (as a Secret with `troubleshoot.sh/kind: preflight` label), discovered via `helm template | preflight -`
- **EC v3** requires the preflight spec outside the chart in the kots/ manifests directory, discovered via `FindResourceByGVK`
- **EC requires v1beta3** — EC v3 explicitly rejects v1beta2
- This forces vendors to maintain two copies of the same preflight spec in different locations with different formats
- There's no single place to define preflights that works for both install methods

### EC v3 doesn't render Helm templates in kots/ manifests before parsing
- If the kots/preflight.yaml contains Helm template syntax (e.g. `{{- if not .Values.postgresql.enabled }}`), EC v3 fails with: `"failed to parse: yaml: line 6: did not find expected node content"`
- EC v3 reads the file as raw YAML and tries to parse it before any Helm rendering happens
- This means the kots/ preflight spec must be static — no conditional collectors or dynamic values
- The in-chart Secret version can use Helm templates (since it goes through `helm template` rendering), but the kots/ version cannot
- **Result**: The EC install gets a less capable preflight spec (no conditional external DB check) than the native Helm install, even though they should be equivalent

---

## Vendor Portal UI

### Helm chart reference not generated for existing releases
- Had a release already promoted with a Helm chart, but the Enterprise Portal chart reference showed "not generated"
- The issue was that the GitHub content integration was enabled after the release was promoted
- Chart reference docs are only generated at promotion time — existing releases aren't retroactively processed
- Had to create a new release and re-promote with the same version label to trigger generation
- Would be helpful if connecting the GitHub repo triggered regeneration for the current active release

### Content preview: branch selection resets when selecting customer
- In the Enterprise Portal Content preview section, selecting a branch first and then selecting a customer causes the branch dropdown to reset
- Have to re-select the branch after picking a customer every time
- Ref: https://www.loom.com/share/2c1576e262004555a9df46ca233e4a39

### "Connect GitHub" shown when feature flag is disabled
- The Enterprise Portal Content page shows the "Connect to GitHub" section even when the GitHub integration feature flag is disabled for the team
- This is confusing — clicking it either fails or leads to a dead end
- The section should be hidden when the feature flag is off

### "Restore default template" modal CSS broken
- In the Enterprise Portal email template editor, clicking "Restore default template" opens a confirmation modal with broken/unstyled CSS
- The modal content (title, description, buttons) appears to render without proper styling
- Located in `EmailTemplateEditor.tsx` — the `Modal` component may not be inheriting the app's CSS context
- Affects the "User Invitation" template and likely other email templates

---

## Enterprise Portal

### Expired verification code shows no error in UI
- When entering an expired verification code during enterprise login, the UI shows nothing — no error message, no feedback
- The error is only visible in the browser network tab (API returns 401)
- User has no idea the code expired — it just silently fails
- Should show a clear message like "Verification code has expired. Please request a new one."

### theme.yaml branding key not documented
- Fields must be nested under `branding:` key in theme.yaml
- The docs show flat structure but the code expects nested
- Discovered by reading vandoor source code

### Self-serve signup URL is wrong in Vendor Portal
- Vendor Portal shows the signup URL as `https://enterprise.replicated.com/signup`
- This URL immediately redirects to the login page because on the multi-tenant deployment, `/signup` has no app context — `PORTAL_APP_SLUG` is not set for any specific app
- The correct URL is `https://enterprise.replicated.com/{app_slug}/signup` (e.g. `https://enterprise.replicated.com/statuspage/signup`)
- Spent significant time debugging why emails weren't being sent, inspecting SQS consumers, integration-api code, email templates — the actual issue was the signup form was never reached because of this redirect
- The signup page checks `checkTrialSignupEnabled(appSlug)` which fails or returns false without the app slug, then redirects to `/login`
- The login page interprets "signup" as the app slug in the URL path, sending requests with `app_slug: "signup"` to the magic link endpoint
- **Impact**: Self-serve signup is completely broken when using the URL shown in Vendor Portal

### SVG logos don't render
- SVG files with `<text>` elements fail when converted to data URI
- No error — just shows broken image
- PNG works fine

### Template variable names don't match docs
- Docs say `{{ license.licenseId }}` but actual field is `{{ license.id }}`
- Had to read vandoor source code to find correct variable names

### No {{else}} in template syntax
- Only `{{#if}}` and `{{#ifEquals}}` are supported
- No `{{else}}` — using `{{#if}}` with a falsy condition produces blank page
- Not documented as a limitation

### main branch is hidden
- Enterprise Portal never shows content from `main` branch
- Must create version branches (e.g. `0.16.0`) for content to appear
- Documented but easy to miss — spent time debugging why preview showed no custom content

### Internal markdown links don't work
- Links like `[Requirements](pages/requirements.md)` don't navigate
- Absolute paths `[Requirements](/pages/requirements.md)` also don't work
- Navigation only works via sidebar clicks

---

## Vendor Portal / RBAC

### Missing RBAC permissions produce vague errors
- `replicated vm create` with missing VM permissions says `"You must read and accept the Compatibility Matrix Terms of Service"` — which sounds like a terms acceptance issue, not RBAC. The actual problem was missing `KOTS/VM/**` permission
- `replicated network update` with missing network permissions gives the same terms of service error
- Each time a new operation failed, we had to consult the RBAC resource names docs, try a new resource pattern, and re-test — a trial-and-error loop

### RBAC resource names hard to discover
- Customer operations split across two paths: `kots/app/*/license/**` for create/read/update but `kots/license/*/archive` for archiving
- No way to know what resource a command needs without looking up the docs or getting a permission error

---

## Custom Domain

### Cloudflare CNAME cross-user ban
- CNAME to `proxy.replicated.com` blocked when Cloudflare proxy (orange cloud) is enabled
- Error: "CNAME Cross-User Banned"
- Must use DNS-only (grey cloud) mode
- Not documented in Replicated custom domain setup

