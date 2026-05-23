# Repository Guidelines

## Project Shape

This repo contains an iKuai router monitoring dashboard with a Go backend and a Vue/Vite frontend.

- `backend/cmd/server/main.go` is the backend entrypoint. It wires config, logging, the monitor service, CORS, `/health`, and `/api/v1/*` routes.
- `backend/internal/config` reads runtime config from environment variables.
- `backend/internal/controller` contains Gin handlers and response envelopes.
- `backend/internal/service` contains iKuai SDK calls, mock data, DTOs, and aggregation logic.
- `frontend/src/api/monitor.ts` is the shared typed API client for the main monitor pages.
- `frontend/src/composables` owns polling, formatting, filtering, sorting, and local UI state.
- `frontend/src/views` and `frontend/src/components` contain the dashboard UI.
- `docs/design.md` is the product and architecture reference, but source code is the authority when they diverge.

## Commands

Backend commands run from the repo root:

```sh
go run ./backend/cmd/server
go test ./...
gofmt -w backend
```

Frontend commands run from `frontend/`:

```sh
pnpm install
pnpm dev
pnpm run type-check
pnpm run build
pnpm run lint
pnpm run format
```

Use `pnpm run build` as the strongest frontend verification because it runs `vue-tsc --build` before Vite. Note that `pnpm run lint` and `pnpm run format` are mutating commands in this project.

## Runtime Configuration

The backend listens on `PORT`, defaulting to `8080`.

Relevant environment variables:

- `MOCK_MODE=true` forces mock data.
- `MOCK_MODE=false` forces a real iKuai connection attempt.
- `IKUAI_URL` sets the router URL.
- `IKUAI_USERNAME` and `IKUAI_PASSWORD` set router login credentials.

The service currently chooses mock mode automatically only when no password is configured. Prefer `MOCK_MODE=true` for local UI work unless you explicitly need to test against a real router.

Do not add secrets, tokens, real router credentials, or personal network details to new files. Prefer environment variables and redact sensitive values in docs or logs.

## Backend Conventions

- Keep new HTTP routes under `/api/v1`.
- Return JSON using the existing envelope pattern: `{"code": 200, "data": ...}` or `{"code": 500, "message": ...}`.
- Keep external response DTOs in `service` with explicit `json` tags.
- Preserve the current unit convention: speed fields are bytes per second, traffic totals are bytes. Let the frontend format display units.
- Real iKuai SDK access is guarded by `MonitorService.mu`; keep that serialization unless you have verified the SDK/session behavior is safe concurrently.
- If a real iKuai call is unavailable or risky while developing UI, add or update high-fidelity mock data rather than blocking the frontend.
- Run `gofmt` on changed Go files.

## Frontend Conventions

- Use Vue 3 `<script setup lang="ts">` and the Composition API.
- Prefer typed API functions in `frontend/src/api/monitor.ts` instead of ad hoc `axios` calls in views. The Vite dev server proxies `/api` to `http://localhost:8080`.
- When adding polling, always clear timers in `onUnmounted`.
- Keep data-fetching and stateful behavior in composables when it will be shared or grows beyond view-local glue.
- Use existing CSS variables and glass-panel/card classes from `frontend/src/assets/main.css` before adding new visual primitives.
- Use `lucide-vue-next` for icons and ECharts for charts when those fit the existing UI.
- Avoid reintroducing Vite starter/demo assets or styles into active screens.

## API Surface To Preserve

Current backend endpoints:

- `GET /health`
- `GET /api/v1/monitor/interface`
- `GET /api/v1/monitor/lan?search=...`
- `GET /api/v1/monitor/network-map`
- `GET /api/v1/monitor/security-hub`
- `GET /api/v1/monitor/multi-wan`

When changing response shapes, update the Go DTOs, frontend TypeScript interfaces, mock data, and all consuming views/composables together.

## Validation Expectations

For backend changes, run:

```sh
go test ./...
```

For frontend changes, run:

```sh
cd frontend
pnpm run build
```

For full-stack changes, smoke test the backend and one proxied frontend path:

```sh
MOCK_MODE=true go run ./backend/cmd/server
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/monitor/interface
```

Then start the frontend with `pnpm dev` and verify it reaches the backend through the `/api` proxy.

## Editing Guidance

- Keep changes scoped; this is a small app with simple layering.
- Check existing implementation before following `docs/design.md` blindly.
- Do not replace the stack with a UI framework such as Element Plus or Ant Design.
- Avoid broad refactors unless the requested change actually needs them.
- Preserve mock mode and real-router mode when touching service logic.
