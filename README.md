# todoe

Distributed task management system demonstrating:

- **Ports & Adapters** (hexagonal) architecture
- **Append-only event store** with CQRS read models
- **Choreography-based sagas** via NATS
- **Database-per-service** isolation
- **Authentication** with session tokens and bcrypt credentials

---

## Architecture

```
┌─────────────────────────────────┐   ┌──────────────────────────────────┐
│         cmd/api  :3000          │   │      cmd/onboarding  :3002        │
│                                 │   │                                   │
│  Tasks (auth-protected CRUD)    │   │  User registration                │
│  Auth  (login / logout)         │   │  Email verification               │
│  DB: todoe                      │   │  Profile completion               │
│                                 │   │  DB: todoe_onboarding             │
└────────────┬────────────────────┘   └──────────────┬────────────────────┘
             │  NATS: user.activated                  │  NATS: user.events
             │◄───────────────────────────────────────┘
             │
         ┌───▼──────────────────────────────────────────────────┐
         │                       NATS                           │
         │  user.events · task.events · credit.results          │
         └───┬──────────────┬──────────────┬────────────────────┘
             │              │              │
      ┌──────▼───┐   ┌──────▼───┐   ┌─────▼──────┐
      │ cmd/audit│   │cmd/credit│   │cmd/welcome │
      │ → Loki   │   │→ scores  │   │→ logs steps│
      └──────────┘   └──────────┘   └────────────┘
```

### Services

| Binary | Port | Database | Responsibility |
|---|---|---|---|
| `cmd/api` | 3000 | `todoe` | Tasks (protected), Auth |
| `cmd/onboarding` | 3002 | `todoe_onboarding` | User registration → activation |
| `cmd/credit` | — | — | Credit scoring via NATS |
| `cmd/audit` | — | — | Event log to Loki |
| `cmd/welcome` | — | — | Onboarding step logger |

### Onboarding → Auth handoff (Event Notification + gRPC Callback)

When a user completes onboarding, `cmd/onboarding` emits `user.activated` on NATS carrying **only the user ID** (event notification pattern — no data in the payload). `cmd/api` receives the notification and calls back to the onboarding gRPC server (`GetUser`) to fetch the email and name, then creates a bcrypt credential. The generated temp password is logged to stdout.

```
cmd/onboarding                NATS                    cmd/api
CompleteProfile() ──► user.activated {user_id} ──► handler
                                                      │
                                                      ▼ gRPC: GetUser(user_id)
                                              cmd/onboarding :50051
                                                      │
                                                      ▼
                                              ActivateUser(id, email, name)
```

This keeps events thin and avoids data coupling — consumers decide what they need and fetch it.

---

## Prerequisites

- Go `1.25+`
- Docker + Docker Compose
- Bun `1.x` (frontend)

## Setup

```bash
go mod download && go mod verify
cd web/vue && bun install && cd ../..
cd web/onboarding && bun install && cd ../..
```

## Environment Variables

All have local defaults — no `.env` required for development.

| Variable | Default | Used by |
|---|---|---|
| `MONGO_URI` | `mongodb://root:root@localhost:27017` | api, onboarding |
| `NATS_URL` | `nats://127.0.0.1:4222` | all services |
| `DB_NAME` | `todoe` (api) / `todoe_onboarding` (onboarding) | api, onboarding |
| `PORT` | `3000` (api) / `3002` (onboarding) | api, onboarding |
| `GRPC_PORT` | `50051` | onboarding (gRPC server) |
| `ONBOARDING_GRPC_ADDR` | `localhost:50051` | api (gRPC client) |
| `LOKI_URL` | `http://localhost:3100` | audit |

## Start Infrastructure

```bash
docker compose up -d
```

Ports: MongoDB `27017` · NATS `4222` · Loki `3100` · Grafana `3001`

## Run Services

Open a terminal per service:

```bash
go run ./cmd/api          # :3000 — tasks + auth
go run ./cmd/onboarding   # :3002 — user onboarding
go run ./cmd/credit       # credit scoring worker
go run ./cmd/welcome      # onboarding step logger
go run ./cmd/audit        # event audit → Loki
```

## Frontends

```bash
cd web/vue && bun run dev        # tasks UI  → http://localhost:5173
cd web/onboarding && bun run dev # onboarding UI → http://localhost:5174
```

The tasks UI proxies `/api` to `:3000`. The onboarding UI proxies `/api` to `:3002`.

## End-to-End Flow

### 1. Onboard a user (port 3002)

```bash
# Register
curl -s -X POST http://localhost:3002/users/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice","email":"alice@example.com"}' | jq .

# Verify email (copy token from cmd/welcome log)
curl -s -X POST http://localhost:3002/users/<id>/verify-email \
  -H 'Content-Type: application/json' \
  -d '{"token":"<token>"}' | jq .

# Complete profile (after credit approved — watch cmd/welcome logs)
curl -s -X POST http://localhost:3002/users/<id>/complete-profile \
  -H 'Content-Type: application/json' \
  -d '{"bio":"Software engineer"}' | jq .
```

After `complete-profile`, `cmd/api` logs the generated temp password:
```
INFO auth: user activated — credential created email=alice@example.com temp_password=<password>
```

### 2. Log in and use tasks (port 3000)

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:3000/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"<temp_password>"}' | jq -r .token)

# Create a task
curl -s -X POST http://localhost:3000/tasks/ \
  -H 'Content-Type: application/json' \
  -H "Authorization: $TOKEN" \
  -d '{"title":"my first task"}' | jq .

# List tasks
curl -s http://localhost:3000/tasks/ -H "Authorization: $TOKEN" | jq .
```

### Register standalone credentials (optional)

```bash
curl -s -X POST http://localhost:3000/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"dev@local.com","password":"secret"}' | jq .
```

## Build & Test

```bash
go build ./...
go test ./...
```

## Observability

- Grafana: [http://localhost:3001](http://localhost:3001)
- Audit events pushed by `cmd/audit` into Loki (provisioned automatically)

## Stop

```bash
docker compose down          # stop containers
docker compose down -v       # stop + remove volumes
```
