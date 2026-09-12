# ChatterBox web

React + TypeScript + Vite frontend for the ChatterBox chat API.

## Develop

Requires the Go API running locally (see the repo root `README.md` / `docker-compose.yml`).

```sh
npm install
npm run dev
```

The dev server proxies `/auth`, `/chats`, `/users`, `/ws` to `http://localhost:8080` (see
`vite.config.ts`), so the app and API appear same-origin and no CORS setup is needed in dev.

## Scripts

- `npm run dev` — start the dev server
- `npm run build` — typecheck (`tsc -b`) then production build
- `npm run lint` — oxlint
- `npm run format` / `format:check` — Prettier

## Structure

```
src/
  lib/            fetch client, tokens, WebSocket client
  types.ts        API response shapes (kept in sync with internal/*/types.go)
  components/ui/  Button, Input, Avatar, Spinner, Dialog, Toast — no UI library
  components/     ProtectedRoute, ConnectionStatus
  features/
    auth/         login/register
    chats/        chat list + shell layout
    messages/     conversation view
    realtime/     WebSocket wiring
```

Build milestones and backend API notes are tracked outside this repo (project memory / issue
tracker) — ask before assuming a feature is out of scope.
