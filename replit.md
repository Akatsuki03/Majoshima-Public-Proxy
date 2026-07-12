# New API - LLM Gateway

## Project Overview
New API is a next-generation LLM gateway and AI asset management system. It provides a unified interface for multiple AI providers (OpenAI, Claude, Gemini, etc.) with quota management, multi-user support, and a modern web UI.

Originally forked from One API, enhanced with modernized UI, multi-language support, and expanded provider compatibility.

## Tech Stack
- **Backend**: Go 1.25 with Gin framework, GORM ORM
- **Frontend**: React 18 + Vite 5, Semi UI (ByteDance), Tailwind CSS
- **Package Manager**: Bun (frontend), Go Modules (backend)
- **Database**: SQLite by default (MySQL/PostgreSQL supported via `SQL_DSN` env var)
- **Cache**: Redis (optional, via `REDIS_CONN_STRING` env var)

## Project Structure
```
/main.go          - Go entry point, embeds web/dist
/common           - Shared utilities, database init
/controller       - API route handlers
/model            - GORM database models
/router           - Gin route definitions
/relay            - Core AI provider relay/proxy logic
/middleware       - Auth, rate limiting, logging
/service          - Business logic (quota, token counting)
/web              - React frontend source
  /src/components - UI components
  /src/pages      - Page components
/new-api          - Compiled Go binary (git-ignored)
/web/dist         - Built frontend (embedded in Go binary)
/start.sh         - Development startup script
```

## Development Workflow

### Starting the Application
The workflow runs `bash start.sh` which:
1. Builds the Go binary (`new-api`) if it doesn't exist
2. Starts the Go backend on port 3000
3. Starts the Vite dev server on port 5000 (proxies /api, /mj, /pg to port 3000)

### Frontend Dev Server
- Runs on `0.0.0.0:5000` with `allowedHosts: true` for Replit proxy compatibility
- Vite proxies API calls to Go backend at `http://localhost:3000`

### Rebuilding the Go Binary
After changing Go source files:
```bash
go build -o new-api .
```

### Building the Frontend
```bash
cd web && bun run build
```
The built files go to `web/dist/`, which the Go binary embeds at compile time.

## Discord OAuth Configuration
- Discord OAuth login is enabled with server membership + role verification
- Settings stored in SQLite `options` table: `discord.enabled`, `discord.client_id`, `discord.client_secret`, `discord.server_id`, `discord.role_id`
- OAuth scope includes `identify+openid+guilds.members.read` for guild/role checks
- Guild membership verified via `GET /users/@me/guilds/{guild_id}/member`
- If `server_id` is set, users must be in that Discord server to log in
- If `role_id` is also set, users must have that specific role
- Production callback URL: `https://rouyashiki.zeabur.app/oauth/discord`
- Admin can configure via System Settings → Discord OAuth section
- Error messages for guild/role failures defined in `i18n/locales/*.yaml`

## Key Environment Variables
- `PORT` - Server port (default: 3000)
- `SQL_DSN` - Database connection string (default: SQLite)
- `REDIS_CONN_STRING` - Redis connection (optional)
- `GIN_MODE` - Set to `debug` for debug mode
- `SESSION_SECRET` - Session secret key
- `INITIAL_ROOT_TOKEN` - Initial admin token

## Deployment - Zeabur (Production)
- **Platform**: Zeabur VPS with Docker
- **Domain**: `https://rouyashiki.zeabur.app`
- **Database**: MySQL (Zeabur managed service)
- **Dockerfile**: Multi-stage build (Bun frontend → Go backend → Debian slim runtime)
- **GitHub Repo**: `Akatsuki03/Majoshima-Public-Proxy`

### Zeabur Environment Variables (New API service)
| Variable | Value | Description |
|----------|-------|-------------|
| `SQL_DSN` | `${MYSQL_USERNAME}:${MYSQL_PASSWORD}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${MYSQL_DATABASE}` | MySQL connection (auto from MySQL service) |
| `PORT` | `3000` | Server port (default in Dockerfile) |
| `GIN_MODE` | `release` | Production mode (set in Dockerfile) |
| `TZ` | `Asia/Tokyo` | Timezone (set in Dockerfile) |

### Zeabur MySQL Service Variables (auto-generated)
| Variable | Value |
|----------|-------|
| `MYSQL_DATABASE` | `zeabur` |
| `MYSQL_HOST` | `${CONTAINER_HOSTNAME}` |
| `MYSQL_USERNAME` | `root` |
| `MYSQL_PASSWORD` | `${MYSQL_ROOT_PASSWORD}` |
| `MYSQL_PORT` | `${DATABASE_PORT}` |

### Dev vs Production Isolation
- **Development (Replit)**: SQLite (`one-api.db`), Vite dev server on :5000, Go on :3000
- **Production (Zeabur)**: MySQL via `SQL_DSN`, single Go binary serves everything on :3000
- No `SQL_DSN` env var = SQLite mode (dev); `SQL_DSN` set = MySQL mode (prod)
- OAuth settings (Discord client_id/secret, server_id, role_id) stored in database `options` table — separate per environment

### Deploying Updates
1. Push code to GitHub (`Akatsuki03/Majoshima-Public-Proxy`)
2. Zeabur auto-deploys from GitHub on push (if connected)
3. Docker build runs: frontend → backend → runtime image
4. App starts with `SQL_DSN` from MySQL service

## Replit Deployment (Autoscale)
- **Build**: `cd web && bun install && DISABLE_ESLINT_PLUGIN=true bun run build && cd .. && go build -o new-api .`
- **Run**: `PORT=5000 ./new-api`

## First Run
On first launch, the app shows a System Initialization wizard to:
1. Verify database connection
2. Set admin account credentials
3. Select usage mode
4. Complete initialization
