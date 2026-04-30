# todoe

A Go modular monolith using **Ports & Adapters** (hexagonal) architecture, append-only event sourcing, and publish-only domain events.

See [`principles.md`](principles.md) for architecture rationale and [`CLAUDE.md`](CLAUDE.md) for coding conventions.

---

## Prerequisites

- Go `1.24+`
- Docker + Docker Compose
- Bun `1.x` (for frontend apps)

## Project Structure

```
cmd/api/          — binary entry point
internal/
  health/         — health check domain
    domain/       — models and constants
    port/         — UseCase and Repository interfaces
    adapter/      — MongoDB adapter + HTTP handler
    application/  — Service
web/
  vue/            — task UI (Vue 3 + TypeScript + Vite)
  onboarding/     — onboarding UI (Vue 3 + TypeScript + Vite)
```

## Environment Variables

All have local defaults — no `.env` required to run locally.

| Variable    | Default                              | Description         |
|-------------|--------------------------------------|---------------------|
| `MONGO_URI` | `mongodb://root:root@localhost:27017` | MongoDB connection   |

## Start Infrastructure

```bash
docker compose up -d
```

Starts MongoDB on port `27017`.

## Run

```bash
go run ./cmd/api
```

API listens on `http://localhost:3000`.

## Endpoints

| Method | Path      | Description         |
|--------|-----------|---------------------|
| `GET`  | `/health` | MongoDB connectivity check |

## Build & Test

```bash
go build ./cmd/api
go test ./...
```

## Frontend

```bash
cd web/vue && bun install && bun dev        # → :5173
cd web/onboarding && bun install && bun dev # → :5173
```
