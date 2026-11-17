<p align="center"><img src="branding/dockup-logo-light.svg" alt="Dockup" height="64"></p>

# Dockup

**Push-to-deploy for Docker Compose.** Minimal, self-hosted CI/CD for **Docker Compose**. One hub with a clean web UI, a tiny agent on every other server (Beszel-style: the agent dials *out*, no inbound ports), and one static binary for both.

Point it at a compose path on a server, or at a git repo (+ `.env` path), and it builds and deploys it with live logs.

## Features

| Jenkins-style capability | In Dockup |
|---|---|
| Jobs | **Projects**: a compose stack on one server, from a path **or** git repo/branch/compose-file |
| Console output | Live streaming logs (SSE), downloadable, secrets masked |
| Build history, retention | Numbered builds, per-project retention, status sparkline |
| Parameterised builds | Declared parameters with defaults/choices, exported as env for compose interpolation |
| Triggers | Manual, **webhook** (GitHub HMAC / GitLab token / `?token=`), **cron**, **SCM polling** (cron + "only if branch changed") |
| Agents / nodes | Remote agents over one outbound connection; auto-reconnect; online status, docker version |
| Credentials | Git tokens / SSH keys, AES-256-GCM at rest, never shown again |
| Pipeline steps | pre-deploy / post-deploy shell commands, health-check gate, `--no-cache`, `--pull`, profiles |
| Rollback | Redeploy any earlier successful git commit |
| Cancel | Kills the whole process group, locally or on a remote agent |
| Notifications | Slack/Discord/any JSON webhook: always / failures / status changes |
| Auth | Local accounts + **OAuth**: GitHub, Google, any OIDC provider (PKCE). Roles: admin / operator / viewer |
| API | Bearer API tokens (`dk_...`) for scripts and CI |
| Audit | Who did what, when |

One build per project runs at a time; others queue. Repeated webhook/schedule triggers coalesce into one pending deploy.

## What's new in 0.3.0

- **New name and identity: Dockup** (formerly DeployHub). The mark is a "D" with an up-arrow cut out: *D + up*.
  Logo files are in [`branding/`](branding/). The panel now ships self-hosted Inter and JetBrains Mono (no CDN).
- **Guided first run**: after creating the admin, a four-step guide checks the host (Docker, Compose, HTTPS),
  picks a server (or onboards an agent and waits for it to connect), creates the first project (including a
  one-click **sample stack** that publishes no ports) and runs the first deploy live, ending in a celebration
  and next steps. A **Get started checklist** on the dashboard tracks the rest (webhook, team, agents,
  credentials, SSO); reopen the guide from the account menu.
- Configuration is now `DOCKUP_*`; every `DEPLOYHUB_*` name still works. API tokens are `dk_…` and agent tokens
  `dks_…`; existing `dh_…`/`dhs_…` tokens keep working.

### Upgrading from DeployHub

- **.deb**: `sudo apt install ./dockup_0.3.0_amd64.deb` replaces the `deployhub` package, copies
  `/var/lib/deployhub` to `/var/lib/dockup` (the old directory is kept) and keeps reading
  `/etc/deployhub/*.env`. Agents: install the same package, then `systemctl disable --now dockup && systemctl enable --now dockup-agent`.
- **Docker**: rename the service/volume paths in your compose file to `dockup` (or keep the old paths: the data
  format is unchanged). Old `DEPLOYHUB_*` variables are accepted.
- Everyone signs in again once (the session cookie was renamed).

## What's new in 0.2.0

**Web UI**: rebuilt as a dense, keyboard-friendly control plane (dark and light themes):
dashboard with KPIs, 7-day pass rate, agent cards, build queue and live feed; projects grouped by
folder with health, history and split deploy menus; project overview with duration/stability chart,
**stage view** matrix and build history with one-click rollback; live console with folding pipeline
stages, line numbers, timestamps, filter, jump-to-error, follow and wrap; sectioned project editor with
encrypted env table, parameters, cron preview (next runs) and an unsaved-changes bar; server onboarding
with copy-paste install commands that waits for the agent to connect; credentials, users, SSO settings,
audit log with search and CSV export; account page with password change, active sessions and expiring
API tokens; **⌘K / Ctrl-K command palette**.

**Security**: one-time setup code for the first admin; no "first OAuth user becomes admin";
auto sign-up requires an email/domain allow-list; login throttling per IP (rightmost X-Forwarded-For)
and per account with constant-time failures; build-log path traversal fixed; viewers/operators no
longer see hooks, env file paths, notification webhooks, secret parameter defaults or credentials in
repo URLs; builds never inherit `DOCKUP_*` secrets; decrypted env files are deleted after each
build; git tokens are scoped to the repository host; secret parameters and sensitive env values are
masked; password change/reset signs out other sessions; session list and revocation; expiring API
tokens; stricter CSP, COOP/CORP, Permissions-Policy, HSTS on HTTPS, `no-store` API responses;
CSV-injection-safe audit export; IP addresses in the audit log.

**Reliability**: idle hub no longer rewrites `state.json` every 2 s; failed SCM polls no longer pile
up or evict build history; changing a project's repo URL re-points the checkout; standard cron
semantics (day-of-month OR day-of-week, `7` = Sunday, `@daily` etc.); agents report running builds,
get a 3-minute reconnect grace and lost builds are detected; failure reasons show the command's last
output line; consistent ports (Docker: 8100, .deb: 8080, change with DOCKUP_ADDR).

## Install

### Docker (hub)

```bash
docker compose up -d --build        # http://localhost:8100
docker logs dockup               # shows the one-time setup code for the first admin
```

Edit `docker-compose.yml` first: the data dir and stacks dir are bind-mounted at the **same path** inside and outside the container (required, because compose resolves relative bind mounts on the host).

### apt / .deb (Debian, Ubuntu; amd64 + arm64)

```bash
sudo apt install ./dockup_0.3.0_amd64.deb     # starts the hub on :8080 and prints the setup code
```

Host your own apt repo from the built packages (see `scripts/make-apt-repo.sh`), then clients do:

```bash
echo 'deb [trusted=yes] https://YOUR-HOST/apt ./' | sudo tee /etc/apt/sources.list.d/dockup.list
sudo apt update && sudo apt install dockup
```

Sign it for production: `GPG_KEY=<id> scripts/make-apt-repo.sh`, publish `dockup.gpg`, and use `[signed-by=/usr/share/keyrings/dockup.gpg]`.

Config: `/etc/dockup/dockup.env`. Logs: `journalctl -u dockup`.

### Add another server

In the UI: **Servers → Add server** shows a ready-to-paste command containing the token. Either:

```bash
# docker
DOCKUP_HUB=https://deploy.example.com DOCKUP_TOKEN=dks_xxx docker compose -f docker-compose.agent.yml up -d

# apt
sudo apt install ./dockup_0.3.0_amd64.deb
sudoedit /etc/dockup/agent.env           # DOCKUP_HUB and DOCKUP_TOKEN
sudo systemctl disable --now dockup      # optional: this machine is only an agent
sudo systemctl enable --now dockup-agent
```

### From source

```bash
go build -o dockup ./cmd/dockup      # Go 1.24+, zero third-party dependencies
./dockup serve
```

## OAuth / SSO

Set variables on the hub and restart. Register `<DOCKUP_URL>/auth/<provider>/callback` with the provider.

```
DOCKUP_URL=https://deploy.example.com
DOCKUP_GITHUB_CLIENT_ID / _SECRET
DOCKUP_GOOGLE_CLIENT_ID / _SECRET
DOCKUP_OIDC_ISSUER / _CLIENT_ID / _CLIENT_SECRET / _NAME     # Keycloak, Authentik, Okta, Entra, ...
```

First admin, one of:
- open the UI and use the **one-time setup code** (printed in the hub log and stored in `<data>/setup-code`), or
- `DOCKUP_ADMIN_USER` + `DOCKUP_ADMIN_PASSWORD`, or
- `DOCKUP_ADMIN_EMAIL=you@example.com` to invite an SSO admin (binds on first sign-in with that verified email).

Who may sign in afterwards:
1. An admin **invites an email** (Users page). That person's verified provider email binds on first login.
2. Or enable **auto sign-up** in Settings, limited to an allow-list of emails/domains (`@example.com`), with role viewer/operator.

Unverified provider emails are never trusted for binding.

## Using it

- **Path mode**: absolute directory (or file) on the target server, e.g. `/opt/stacks/blog`. `.env` path is optional (absolute or relative to the compose file).
- **Git mode**: repo URL + branch (+ optional compose file path). Private repos: add a credential. Checkout lives in the server's data dir.
- **Inline env**: key/values stored encrypted, passed as an extra `--env-file`.
- **Webhook**: the project's *Webhook & API* tab shows the URL and secret. GitHub/Gitea: content type JSON, secret = shown secret (HMAC verified). GitLab: secret token. Only pushes to the project's branch trigger.
- **API**: `curl -X POST -H "Authorization: Bearer dk_..." -H 'Content-Type: application/json' -d '{"action":"deploy"}' https://hub/api/projects/<id>/builds`

## Security notes

- The hub/agent drive the host Docker daemon: treat access to Dockup as root on those machines.
- Put TLS in front (Caddy/nginx/Traefik) and set `DOCKUP_URL` + `DOCKUP_TRUST_PROXY=1`.
- Back up the data dir **including `secret.key`** (or set `DOCKUP_SECRET_KEY`); without it stored credentials cannot be decrypted.
- Sessions are HttpOnly + SameSite=Lax cookies; mutating requests need an `X-Dockup` header (CSRF); login is rate limited per IP and per account; passwords use PBKDF2-SHA256 (210k iterations, 10+ characters).
- Builds run with a cleaned environment: `DOCKUP_*` variables (secret key, OAuth secrets, agent token) are never passed to compose, hooks or git.
- Only admins see hook commands, env file paths, notification URLs and webhook secrets.
- Pre/post commands and compose files run with full privileges by design, so only admins can edit projects.

## Layout

```
cmd/dockup     entrypoint + embedded web UI
internal/hub      HTTP API, auth, OAuth, queue, scheduler, agent link
internal/agent    remote worker
internal/runner   git + docker compose execution (shared by hub-local and agents)
internal/store    JSON state, encrypted secrets
internal/cron     5-field cron
packaging/ scripts/ Dockerfile docker-compose*.yml
```

State is one JSON file plus per-build log files: ideal for tens of servers and thousands of builds. A SQL backend is the natural next step beyond that.
