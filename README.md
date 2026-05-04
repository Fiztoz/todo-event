# todoe

Modular monolith with Ports & Adapters architecture, append-only persistence, and event-driven side effects. Domains run as independent binaries communicating over RabbitMQ.

## Services

| Binary | Port | Role |
|---|---|---|
| `cmd/api` | `3000` | Auth (`/auth/*`), tasks (`/tasks/*`), health (`/health`). Consumes `user.activated` to create credentials. |
| `cmd/onboarding` | `3002` | User registration & onboarding flow (`/users/*`). Publishes `user.events`; consumes `credit.results`. |
| `cmd/captcha` | `3010` | Captcha challenge issue + verify (`/captcha/*`). |
| `cmd/welcome` | — | Logs the four onboarding milestones from `user.events`. |
| `cmd/credit` | — | Scores users on `user.email_verified` and publishes to `credit.results`. |
| `cmd/audit` | — | Forwards `task.events` to Loki. |

## Architecture

### Service topology

HTTP from the browsers, fanout pub/sub over RabbitMQ between services, MongoDB per domain, audit stream to Loki/Grafana.

```
   ┌──────────────────┐                          ┌──────────────────┐
   │   Task UI        │                          │  Onboarding UI   │
   │   web/vue        │                          │  web/onboarding  │
   └────────┬─────────┘                          └─┬──────┬─────┬───┘
            │                                      │      │     │
            │ /auth/*                /api/users/*  │      │     │ /captcha/*
            │ /tasks/*              /auth/login    │      │     │
            ▼                                      ▼      │     ▼
   ┌──────────────────┐            ┌──────────────────┐  │  ┌──────────────────┐
   │  cmd/api  :3000  │ ◄──────────┤ cmd/onboarding   │  │  │  cmd/captcha     │
   │  auth · tasks    │            │  :3002  users    │  │  │  :3010           │
   │  health          │            │                  │  │  │                  │
   └─┬───────────┬────┘            └─┬────────────┬───┘  │  └────────┬─────────┘
     │           │                   │            │      │           │
     │ pub       │ sub               │ pub        │ sub  │           │
     │ task.     │ user.events       │ user.      │ credit│          │
     │ events    │ (authen.user.     │ events     │.results          │
     │           │  events queue)    │            │                  │
     │           │                   │            │                  │
     ▼           ▲                   ▼            ▲                  ▼
   ╔═══════════════════════════════════════════════════╗      ┌──────────────┐
   ║                    RabbitMQ                        ║      │   MongoDB    │
   ║                                                    ║◄─────┤  (shared,    │
   ║  task.events     ─► audit.task.events    ─► audit  ║      │   per-domain │
   ║  user.events     ─► welcome.user.events  ─► welcome║      │   collections)│
   ║                  ─► credit.user.events   ─► credit ║      └──────────────┘
   ║                  ─► authen.user.events   ─► api    ║
   ║  credit.results  ─► onboarding.credit.results      ║
   ║                                          ─► onboard║
   ╚═══════════════════════════════════════════════════╝
            ▲                            ▲
            │ pub user.events            │ pub credit.results
            │  (from onboarding)         │  (from credit)
            │                            │
   ┌──────────────────┐         ┌──────────────────┐         ┌──────────────┐
   │  cmd/welcome     │         │  cmd/credit      │         │  cmd/audit   │
   │  logs 4 steps    │         │  scores users    │         │  → Loki      │
   └──────────────────┘         └──────────────────┘         └──────┬───────┘
                                                                    ▼
                                                             ┌──────────────┐
                                                             │  Grafana     │
                                                             │  :3001       │
                                                             └──────────────┘
```

### Onboarding choreography

The four-step user flow as it crosses processes. Every cross-service link goes through RabbitMQ, durable and replayable.

```
  Browser     Onboarding     RabbitMQ        Welcome     Credit       API
     │            │             │               │           │           │
     │ POST /users/register     │               │           │           │
     ├───────────►│             │               │           │           │
     │            │  user.registered            │           │           │
     │            ├────────────►│  ─────────────►│ step 1/4 │           │
     │            │             │               │           │           │
     │ POST /users/:id/verify-email              │           │           │
     ├───────────►│             │               │           │           │
     │            │  user.email_verified        │           │           │
     │            ├────────────►│  ─────────────►│ step 2/4 │           │
     │            │             │  ─────────────────────────►│           │
     │            │             │               │  scores   │           │
     │            │             │  ◄─────credit.results──────┤           │
     │            │◄────────────┤               │           │           │
     │            │  RecordCreditScore          │           │           │
     │            │  user.credit_scored         │           │           │
     │            ├────────────►│  ─────────────►│ step 3/4 │           │
     │            │             │               │           │           │
     │ POST /users/:id/complete-profile          │           │           │
     ├───────────►│             │               │           │           │
     │            │  user.profile_completed     │           │           │
     │            ├────────────►│  ─────────────►│ step 4/4 │           │
     │            │  user.activated             │           │           │
     │            ├────────────►│  ──────────────────────────────────────►│
     │            │             │               │           │ ActivateUser
     │            │             │               │           │           │
     │ POST /auth/login                         │           │           │
     ├──────────────────────────────────────────────────────────────────►│
     │◄───────────────────────── session token ─────────────────────────┤
```

## Prerequisites

Install these first:

- Go `1.25+`
- Docker + Docker Compose
- Bun `1.x` (for frontend apps)
- `curl` (for quick API checks)

## Repo Setup

```bash
# from repo root
go mod download
go mod verify
```

Frontend dependencies:

```bash
cd web/vue && bun install
cd ../onboarding && bun install
cd ../..
```

## Environment Variables

The backend binaries use environment variables, but all of them have local defaults.

Create a `.env` (optional but recommended) in repo root:

```env
# API
MONGO_URI=mongodb://root:root@localhost:27017
AMQP_URL=amqp://guest:guest@localhost:5672/

# Audit service
LOKI_URL=http://localhost:3100
```

If omitted:
- `MONGO_URI` defaults to `mongodb://root:root@localhost:27017`
- `AMQP_URL` defaults to `amqp://guest:guest@localhost:5672/`
- `LOKI_URL` defaults to `http://localhost:3100`

## Start Infrastructure

Run MongoDB, RabbitMQ, Loki, Grafana:

```bash
docker compose -f compose.yml up -d
```

Exposed ports:
- MongoDB: `27017`
- RabbitMQ AMQP: `5672`
- RabbitMQ management UI: `15672` (guest/guest)
- Loki: `3100`
- Grafana: `3001` (container `3000`)

## Run Services (6 terminals)

```bash
go run ./cmd/api          # :3000  auth, tasks, health
go run ./cmd/onboarding   # :3002  user registration & onboarding
go run ./cmd/captcha      # :3010  captcha challenges
go run ./cmd/welcome      #        logs onboarding milestones
go run ./cmd/credit       #        scores users on email-verified
go run ./cmd/audit        #        forwards task events to Loki
```

Each binary connects to MongoDB and RabbitMQ on startup; HTTP services additionally listen on the port shown above.

## Run Frontends (optional)

### Task UI

```bash
cd web/vue
bun run dev
```

### Onboarding UI

```bash
cd web/onboarding
bun run dev
```

## Quick Verification

Health check:

```bash
curl http://localhost:3000/health
```

Create a task:

```bash
curl -X POST http://localhost:3000/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"first task","description":"setup complete"}'
```

List tasks:

```bash
curl http://localhost:3000/tasks
```

## Build & Test

```bash
go build ./...
go test ./...
```

## Messaging

Cross-domain events flow through RabbitMQ fanout exchanges with durable queues and persistent delivery. Each publishing service declares the queues its downstream consumers expect at startup, so messages buffer on disk while consumers are offline:

| Exchange | Queue | Consumer |
|---|---|---|
| `task.events` | `audit.task.events` | `cmd/audit` |
| `user.events` | `welcome.user.events` | `cmd/welcome` |
| `user.events` | `credit.user.events` | `cmd/credit` |
| `user.events` | `authen.user.events` | `cmd/api` (creates credential on `user.activated`) |
| `credit.results` | `onboarding.credit.results` | `cmd/onboarding` |

Inspect queues, bindings, and ready/unacked counts at [http://localhost:15672](http://localhost:15672) (guest/guest).

## Observability

- Grafana: [http://localhost:3001](http://localhost:3001)
- Loki datasource is provisioned from `provisioning/`.
- Audit events are pushed by `cmd/audit` into Loki.

## Stop Everything

```bash
docker compose -f compose.yml down
```

To also remove persisted volumes:

```bash
docker compose -f compose.yml down -v
```
