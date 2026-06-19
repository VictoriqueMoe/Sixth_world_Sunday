# Sixth World Sunday

A private, invite-based, Shadowrun ("Sixth World") themed, Discord-style community chat node. Self-hosted, web-only, with session-cookie auth.

## Features

- **Channels** — group chat rooms with:
  - Text chat with markdown and syntax-highlighted code blocks
  - Image / video / file uploads
  - Message replies, `@mentions`, emoji reactions, pinned messages
  - Per-message edit and delete
  - Per-channel message search
  - Per-channel unread badges (suppressed when the channel is muted)
  - Public and private channels, channel creation and browsing / discovery
  - "Ghost" (silent) join
  - Per-channel nickname and avatar overrides
- **Voice and screen-sharing** in channels via a self-hosted [LiveKit](https://livekit.io/) server. Optional, off by default, gated by an admin "Enable Voice Chat" setting.
- **Users and auth** — invite-based registration, login, email verification, password reset, rich profiles (bio, avatar, banner, social links), online presence, and user blocking.
- **Roles** — Sysop / Admin / Moderator, plus admin-defined vanity role pills.
- **Moderation** — kick / ban / timeout, banned-word rules (global and per-channel), user and message reports, and an audit log.
- **Notifications** — in-app notification bell, browser desktop notifications (Notification API; fires for replies, mentions, reactions, etc. while a tab is open or backgrounded), and optional email notifications.
- **Admin console** — stats, user management, invites, reports, banned words, vanity roles, content rules, and site settings.
- **Full-text search** over users and chat messages, backed by Postgres `pg_trgm` and generated `tsvector` columns.

## Not included

To keep contributors from re-adding them, the following are deliberately **not** part of this app: direct messages, art galleries, GIF / Giphy integration, live streaming, a mobile app, forum posts / comments, games, and closed-tab web push (desktop notifications only fire while a tab is open or backgrounded).

## Tech stack

**Backend**

- Go 1.26 with [Fiber v3](https://gofiber.io/)
- PostgreSQL via [`jackc/pgx/v5`](https://github.com/jackc/pgx), with `pg_trgm` and generated `tsvector` search columns
- [`pressly/goose`](https://github.com/pressly/goose) for migrations — a single embedded SQL migration, applied automatically on startup
- [LiveKit](https://livekit.io/) (self-hosted) for voice and screen-share
- testcontainers-go for repository-layer tests against a real Postgres

**Frontend**

- React 19 + TypeScript
- Vite
- react-router v7
- TanStack Query

The frontend builds into `../static/` and is embedded into the Go binary via `go:embed`, so the single binary serves both the SPA and the JSON API.

## Prerequisites

- Go 1.26 or newer (see `go.mod`)
- Node.js LTS
- Docker, or a local PostgreSQL instance
- A LiveKit server (optional — only needed for voice / screen-share)

## Configuration

Two env files sit next to each other.

**`postgres.env`** — Postgres bootstrap credentials, read by both the postgres container and the app. Copy from `postgres.env.example`:

| Variable            | Default                | Description            |
|---------------------|------------------------|------------------------|
| `POSTGRES_USER`     | `sixthworld`           | DB role for the app    |
| `POSTGRES_PASSWORD` | `changeme`             | DB password            |
| `POSTGRES_DB`       | `sixth_world_sunday`   | Database name          |

**`.env`** — everything else. Copy from `.env.example`:

| Variable            | Default                 | Description                                                                                  |
|---------------------|-------------------------|----------------------------------------------------------------------------------------------|
| `PORT`              | `4323`                  | HTTP port the Go server listens on                                                           |
| `BASE_URL`          | `http://localhost:4323` | Public base URL, used for CORS and for absolute links in emails, OG embeds, and notifications |
| `UPLOAD_DIR`        | `uploads`               | Directory for uploaded files (relative to the working dir)                                   |
| `LOG_LEVEL`         | `info`                  | Initial log level (overridable from the admin panel at runtime)                              |
| `POSTGRES_HOST`     | `postgres`              | Postgres host (`postgres` is the compose service name)                                       |
| `POSTGRES_PORT`     | `5432`                  | Postgres port. In-container port under compose; use `5017` for app-on-host / Postgres-in-Docker |
| `POSTGRES_SSL_MODE` | `disable`               | Postgres SSL mode (`disable`, `require`, `verify-ca`, `verify-full`)                         |
| `DATABASE_URL`      | (empty)                 | Full connection string. If set, overrides the discrete `POSTGRES_*` vars                     |

Most runtime behaviour lives in the database in the `site_settings` table and is editable from the admin panel with hot reload — including the LiveKit URL / API key / secret, SMTP / email-provider settings, Cloudflare email, Turnstile keys, upload and body size limits, registration mode, maintenance mode, default theme, channel limits, and the Sentry / OTLP / Pyroscope endpoints. Any of these can be **seeded** from the env file by setting an uppercased version of the setting key (e.g. `SMTP_HOST`, `LIVEKIT_URL`, `TURNSTILE_SITE_KEY`, `SENTRY_DSN`); the env value becomes the initial default. The `.env` file is otherwise only for what must exist before the DB is reachable.

## Local development

The app needs Postgres reachable before it will boot. Migrations are embedded and applied automatically on startup, so there is no manual migration step.

**1. Start Postgres**

```bash
# Just the postgres service from compose (maps host 127.0.0.1:5017 -> container 5432)
docker compose up -d postgres
```

**2. Configure env**

```bash
cp postgres.env.example postgres.env
cp .env.example .env
```

For an app-on-host setup connecting to the Dockerized Postgres, set `POSTGRES_HOST=127.0.0.1` and `POSTGRES_PORT=5017` in `.env`.

**3. Build the frontend**

```bash
cd frontend
npm install
npm run build   # tsc + vite build into ../static/
```

**4. Run the Go server**

```bash
go run .
```

The server listens on `:4323` (override with `PORT`). The first user to register is assigned the Sysop (super admin) role, so start there to unlock the admin console.

For frontend iteration, `npm run dev` (in `frontend/`) runs the Vite dev server with HMR and proxies the API to the Go server.

## Docker deployment

**Build and run the full stack locally:**

```bash
docker compose up -d --build
```

This builds the multi-stage image (Node build stage -> Go build stage -> Alpine runtime with FFmpeg and libwebp-tools) and runs the app alongside Postgres, Valkey, and LiveKit. The app's container port `4323` is published on host port `2313`, so visit `http://localhost:2313`.

**Use the prebuilt image instead of building:**

```bash
docker compose -f docker-compose.prod.yml up -d
```

This pulls `ghcr.io/victoriquemoe/sixth_world_sunday:latest`.

Persistent data:

- **Postgres** — the named volume `postgres_data`.
- **Uploaded media** — the app service bind-mounts `./data:/app/data`; set `UPLOAD_DIR=data/uploads` so uploads land on the host mount.

Run behind a reverse proxy (Caddy, Nginx, ...) for TLS. The server sets its own cache headers (`/static/assets/*` immutable, HTML `no-store`, API `no-cache`); the proxy just forwards requests and upgrades WebSocket connections. If you enable voice, copy `livekit.yaml.example` to `livekit.yaml`, set a real key/secret in its `keys:` block, enter the matching values under Admin -> Settings -> Voice Chat, and route a public `wss://` host to the LiveKit container.

## Project layout

```
main.go              entrypoint; reads PORT, starts the Fiber app
server.go            wires the dependency graph (no DI container)
internal/            backend, layered controller -> service -> repository
  config/            env + site-setting definitions
  db/migrations/     single embedded goose migration (00001_init.sql)
frontend/            React 19 + Vite SPA; builds into ../static/
static/              embedded build output (go:embed)
```

The schema is one consolidated initial migration. Further schema changes are added as fresh goose migrations on top:

```bash
goose -dir internal/db/migrations create <name> sql
```

## Testing

```bash
go test ./...
```

Repository-layer tests boot a real `postgres:latest` container via testcontainers-go, so **Docker must be running** on the host.

Generated Go interface mocks (flagged in `.mockery.yml`) are regenerated with:

```bash
bash scripts/regen_mocks.sh
```

Frontend checks:

```bash
cd frontend
npm run lint      # eslint, --max-warnings=0
npm run test      # vitest
npm run build     # tsc + vite build
```

## License

Released under the [MIT License](LICENSE).
