# todoe

A Go modular monolith scaffold using **Ports & Adapters** (hexagonal) architecture, append-only event sourcing, and publish-only domain events.

See [`principles.md`](principles.md) for the full architecture rationale and [`CLAUDE.md`](CLAUDE.md) for coding conventions.

---

## Prerequisites

- Go `1.24+`
- Docker + Docker Compose
- Bun `1.x` (for frontend apps)

## Project Structure

```
cmd/api/          — binary entry point
internal/         — one subdirectory per domain
  <domain>/
    domain/       — models, constants, value objects
    port/         — UseCase (PortIn) and Repository (PortOut) interfaces
    adapter/      — infrastructure adapters + HTTP handlers
    application/  — Service (orchestration + pure logic)
web/
  vue/            — task UI (Vue 3 + TypeScript + Vite)
  onboarding/     — onboarding UI (Vue 3 + TypeScript + Vite)
```

## Run

```bash
go run ./cmd/api
```

## Build

```bash
go build ./cmd/api
go test ./...
```

## Frontend

```bash
cd web/vue && bun install && bun dev        # task UI → :5173
cd web/onboarding && bun install && bun dev # onboarding UI → :5173
```
