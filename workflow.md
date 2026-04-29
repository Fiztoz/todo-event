# Message Workflows

## Architecture Overview

A modular monolith following **Ports & Adapters** (hexagonal) architecture with event sourcing and CQRS, split across multiple independently deployable binaries.

---

## Domain: Task (event sourcing + CQRS)

- **Event sourcing** — task changes stored as immutable events in `task_events` (MongoDB)
- **CQRS** — `ProjectionHandler` maintains a `tasks_view` read model so queries never replay events
- **Use cases** — `CreateTask`, `ListTasks`, `GetTask`, `ChangeStatus`
- **HTTP** — `POST /tasks`, `GET /tasks`, `GET /tasks/:id`, `PATCH /tasks/:id/status`

---

## Domain: User (onboarding workflow)

4-step onboarding demonstrating choreography message workflows.

```
Step 1 (user)      POST /users/register              → user.registered
Step 2 (user)      POST /users/:id/verify-email       → user.email_verified
Step 3 (automatic) cmd/credit reacts                 → user.credit_scored
Step 4 (user)      POST /users/:id/complete-profile   → user.profile_completed
                   (blocked if credit denied)
```

### Full Choreography

```
cmd/api           cmd/welcome              cmd/credit
  │                    │                       │
  │──registered ──────▶│ "token sent"           │
  │                    │                       │
  │──email_verified ──▶│ "email confirmed"      │
  │                    │    ──email_verified ──▶│ calls fake credit API
  │                    │                       │ publishes to credit.results
  │                    │◀── credit_scored ─────│
  │                    │ "score: 720 approved"  │
  │                    │  OR "score: 210 denied"│
  │──profile_completed▶│ "welcome aboard!"      │
  │  (if approved)     │                       │
```

---

## Binaries

| Binary | Role |
|---|---|
| `cmd/api` | HTTP API — task + user domains |
| `cmd/audit` | Subscribes to `task.events` via NATS, pushes to Loki |
| `cmd/credit` | Subscribes to `user.events`, scores credit, publishes to `credit.results` |
| `cmd/welcome` | Subscribes to `user.events`, logs each onboarding step |

## NATS Subjects

| Subject | Publisher | Subscribers |
|---|---|---|
| `task.events` | `cmd/api` | `cmd/audit`, `cmd/welcome` |
| `user.events` | `cmd/api` | `cmd/credit`, `cmd/welcome` |
| `credit.results` | `cmd/credit` | `cmd/api` (loops back to call `RecordCreditScore`) |

---

## Infrastructure

All services defined in `compose.yml`:

| Service | Port | Purpose |
|---|---|---|
| MongoDB | 27017 | Event store + read models |
| NATS | 4222 | Integration message broker |
| Loki | 3100 | Audit log storage |
| Grafana | 3001 | Log visualisation (Loki pre-provisioned) |

---

## Key Patterns

### Domain Events (internal)
In-process `EventBus` — used by `ProjectionHandler` to maintain read models. Never crosses service boundaries.

### Integration Messages (external)
NATS — used for cross-service communication. Each service is independently deployable and knows nothing about the internals of others.

### Event-Carried State Transfer
NATS messages carry the full entity state (e.g., full `User` struct), so consumers never need to call back to the source service.

### Choreography
Services react to events independently. No central coordinator. Adding a new reaction = new service + new `Subscribe` call, publishing service unchanged.

### Automatic Workflow Step
The credit scoring step (step 3) fires without user interaction — `cmd/credit` reacts to `user.email_verified` automatically, demonstrating that not all workflow steps require human input.

### Event Sourcing
Append-only event log per domain. State is never updated in place — each change appends a new event. Read models are maintained by projection handlers.

### CQRS
Write side appends to the event store. Read side queries the projected view. Queries are O(1) lookups, not event replays.
