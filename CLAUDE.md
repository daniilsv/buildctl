# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Build Assistant is a CI/CD build tracking system with three components:
- **Backend** (`back/`) - Go REST API server
- **Frontend** (`front/`) - React web UI

The system tracks builds, stores logs, manages artifacts (files and container images), provides AI-powered commit analysis, and sends Telegram notifications.

## Common Development Commands

### Backend (Go)
```bash
cd back

# Build
go build -o bin/server ./cmd/server

# Run locally (requires .env or environment variables)
go run ./cmd/server/main.go
```

### Frontend (React + Vite)
```bash
cd front

# Install dependencies
pnpm install

# Development server (hot reload)
pnpm dev

# Build for production
pnpm build

# Preview production build
pnpm preview
```

### Docker Compose
```bash
# Production deployment
cd .infra
docker-compose -f compose.yml up -d

# Local development (includes PostgreSQL)
docker-compose -f compose.local.yml up -d
```

## Architecture

### Backend Structure

**Three-layer architecture:**
1. **HTTP Layer** (`internal/api/`) - Chi router, middleware, handlers
2. **Service Layer** (`internal/services/`) - Business logic
3. **Data Layer** (`db/`) - sqlc-generated queries, migrations

**Key directories:**
- `cmd/server/main.go` - Entry point with dependency injection
- `config/` - Environment variable configuration
- `db/migrations/` - SQL migration files (auto-run on startup)
- `db/queries/` - SQL query definitions for sqlc
- `db/gen/` - Generated Go code from sqlc
- `internal/api/handlers/` - HTTP handlers by feature (events, artifacts, projects, branches, builds, tokens)
- `internal/services/` - Business logic services
- `internal/workers/` - Async task processing pool
- `internal/auth/` - OIDC and token authentication
- `internal/notifications/` - Telegram and webhook notifications
- `internal/git/` - Gitea client for fetching commit info
- `internal/ai/` - OpenAI/OpenRouter client for commit summaries
- `pkg/s3/` - AWS S3 integration

**Database schema (PostgreSQL):**
- `projects` - Repository metadata, tokens, settings
- `branches` - Branch tracking per project
- `builds` - Build records with status and timestamps
- `build_logs` - Detailed log messages per build
- `artifacts` - Unified schema for file artifacts and container images (artifact_type field)
- `access_tokens` - API tokens (SHA256 hashed)
- `commits_summary` - Cached AI-generated summaries

**Query generation:** Uses [sqlc](https://sqlc.dev/) to generate type-safe Go code from SQL queries in `db/queries/*.sql`. Regenerate with `sqlc generate` if you modify queries.

**Migrations:** Embedded in binary using golang-migrate. Auto-run on server startup. Add new migrations as numbered SQL files in `db/migrations/`.

### Frontend Structure

**Technology stack:**
- React 18 + TypeScript
- Vite for build tooling
- React Router v6 for routing
- TanStack React Query v5 for server state management
- Axios for HTTP requests
- Tailwind CSS for styling

**Key directories:**
- `src/pages/` - Route-level page components (Dashboard, ProjectDetail, BranchDetail, BuildDetail, Settings, Login, Callback)
- `src/components/` - Reusable UI components and forms
- `src/api/` - API client layer with typed endpoints
- `src/utils/` - Authentication helpers (token storage)

**Authentication flow:**
1. OIDC login flow initiated at `/auth/login`
2. Callback at `/auth/callback` stores token in localStorage
3. All API requests include Bearer token via Axios interceptor
4. 401/403 responses trigger automatic logout and redirect to login

**API client pattern:**
- Base Axios instance in `api/client.ts` with interceptors
- Feature-specific API modules (projects, builds, branches, tokens)
- React Query hooks for caching and automatic invalidation

## Key Integration Points

### CI/CD to Backend Flow
```
GitHub Actions (buildctl event --status=started)
    ↓
POST /api/v1/events (token auth)
    ↓
EventService creates/updates Build record
    ↓
Worker pool enqueues async task (for success/failed events)
    ↓
Worker fetches commit info from Gitea
    ↓
Worker calls AI API for commit summary
    ↓
Worker sends Telegram/webhook notifications
    ↓
Worker updates branch.last_successful_commit
```

### Artifact Upload Flow
```
buildctl artifact upload
    ↓
POST /api/v1/artifacts/presign (gets S3 presigned URL)
    ↓
Direct upload to S3 via presigned URL
    ↓
POST /api/v1/artifacts/confirm (stores metadata in DB)
```

### Web UI Flow
```
User browses frontend
    ↓
OIDC authentication (token stored in localStorage)
    ↓
API requests with Bearer token
    ↓
Backend validates OIDC token (with 1-hour caching)
    ↓
JSON response rendered in React components
```

## Important Patterns & Conventions

### Authentication
- **Token Auth** (CI/CD): Bearer token → SHA256 hash → DB lookup in `access_tokens`
- **OIDC Auth** (Web UI): Bearer token → OIDC introspection → 1-hour cache → user info extraction
- Tokens are hashed with SHA256 before storage; raw tokens shown only once at creation

### Error Handling
- Standard Go error wrapping with `fmt.Errorf(...%w)`
- HTTP status codes: 400 (bad request), 401 (unauthorized), 404 (not found), 500 (server error)
- Structured logging with `log/slog` for debugging
- Worker failures logged but don't crash worker goroutines

### Database Patterns
- sqlc generates type-safe queries from SQL files
- Context passed through all DB operations
- Soft deletes for artifacts using `deleted_at` timestamp
- Cascade deletes on foreign keys (builds → logs, builds → artifacts)
- Proper indexing on foreign keys and frequently queried columns

### Async Processing
- Worker pool with configurable size (WORKER_POOL_SIZE env var)
- Task queue prevents blocking HTTP handlers
- Two task types: `process_success`, `process_failed`
- Graceful shutdown with context cancellation
- In-memory queue (lost tasks on restart are acceptable)

### Code Organization
- Dependency injection in `main.go`
- Interface-based design (handlers accept interfaces, not concrete types)
- Clear separation: handlers → services → queries
- Feature-based directory structure
- Exported types/functions for public API, unexported for internal helpers

### Frontend Patterns
- React Query for automatic cache invalidation after mutations
- Custom hooks for API calls
- Modal dialogs for confirmations and forms
- Status badges with color coding
- Tailwind utility classes for styling

### Artifact Handling
- Unified schema supports both file artifacts and container images
- `artifact_type` field: 'file' or 'container_image'
- S3 presigned URLs for secure direct uploads (no backend proxy)
- Soft deletes for audit trail (deleted_at timestamp)
- Public URLs generated with S3_PUBLIC_PREFIX for CDN support

## Environment Variables

**Backend** (see `back/config/config.go`):
- `DATABASE_URL` - PostgreSQL connection string (required)
- `PORT` - HTTP server port (default: 8080)
- `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL` - OIDC configuration
- `OPENAI_API_URL`, `OPENAI_API_KEY`, `OPENAI_MODEL` - AI integration (default: OpenRouter with Qwen model)
- `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY` - S3 configuration
- `S3_PATH_STYLE` - Use path-style S3 URLs (default: true, for S3-compatible services)
- `S3_PUBLIC_PREFIX` - Public CDN URL prefix for artifact downloads
- `TELEGRAM_BOT_TOKEN` - Telegram bot token for notifications
- `WORKER_POOL_SIZE` - Number of async workers (default: 3)

**Frontend:**
- API calls use `/api/v1` base path (proxied by nginx in production)

## API Structure

**Authentication:**
- `/auth/login` - Initiate OIDC login (GET)
- `/auth/callback` - OIDC callback handler (GET)
- `/api/v1/auth/userinfo` - Get current user info (GET, OIDC auth)

**Build tracking:**
- `/api/v1/events` - Submit build events (POST, token auth)
- `/api/v1/builds` - List builds (GET, OIDC auth)
- `/api/v1/builds/{id}` - Get build details (GET, OIDC auth)
- `/api/v1/builds/{id}/artifacts` - List build artifacts (GET, OIDC auth)

**Artifact management:**
- `/api/v1/artifacts/presign` - Get presigned S3 URL (POST, token auth)
- `/api/v1/artifacts/confirm` - Confirm artifact upload (POST, token auth)
- `/api/v1/artifacts/container` - Register container image (POST, token auth)
- `/api/v1/artifacts/{id}` - Delete artifact (DELETE, OIDC auth)
- `/api/v1/builds/{id}/artifacts` - Delete all build artifacts (DELETE, OIDC auth)

**Project/Branch management:**
- `/api/v1/projects` - CRUD for projects (GET/POST/PUT/DELETE, OIDC auth)
- `/api/v1/projects/{name}/branches` - CRUD for branches (GET/POST/PUT/DELETE, OIDC auth)

**Token management:**
- `/api/v1/tokens` - CRUD for API tokens (GET/POST/DELETE, OIDC auth)

## Data Flow Examples

### Tracking a Build
1. CI/CD starts: `buildctl event --status=started`
2. Creates Build record via POST /api/v1/events
3. CI/CD uploads artifacts: `buildctl artifact upload --file=dist.tar.gz`
4. Gets presigned URL, uploads to S3, confirms via POST /api/v1/artifacts/confirm
5. CI/CD registers container: `buildctl container register --image=myapp:v1.0.0 --digest=sha256:...`
6. CI/CD completes: `buildctl event --status=success`
7. Backend enqueues worker task
8. Worker fetches commit from Gitea, generates AI summary, sends Telegram notification
9. Frontend shows updated build in dashboard with artifacts

### Creating a Project
1. User logs in via OIDC flow (redirects to provider, callback stores token)
2. User creates project in web UI
3. POST /api/v1/projects creates project record
4. User creates API token in Settings page
5. Raw token shown once, user copies to GitHub Secrets
6. CI/CD uses token in buildctl commands

## Notes for Development

- The backend uses sqlc for query generation. If you modify `db/queries/*.sql`, run `sqlc generate` to regenerate Go code.
- Database migrations run automatically on server startup. Add new migrations with sequential numbering.
- Frontend uses pnpm for package management (not npm/yarn).
- Worker pool processes async tasks (AI summaries, notifications). Tasks are lost on restart (no persistent queue).
- OIDC token caching (1 hour) reduces load on identity provider.
- Artifact files are stored in S3; only metadata is in PostgreSQL.
- Container images can optionally include tar archives, but typically just metadata (name, tag, digest).
- All commit messages and AI summaries are in Russian.
- The system uses soft deletes for artifacts (deleted_at timestamp) to maintain audit trail.
