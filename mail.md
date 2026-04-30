Subject: Workshop prep — todoe repo

Hi,

Before the session, please complete the following setup.

--- Prerequisites ---

- Go 1.24 or later    https://go.dev/dl/
- Docker + Docker Compose
- Bun 1.x             https://bun.sh/
- curl

--- Get the repo ---

  git clone https://github.com/roofimon/todo-event todoe
  cd todoe
  git checkout min

--- What the `min` branch contains ---

This is your starting point:

- cmd/api/main.go         Hello World entry point (fmt.Println)
- go.mod                  Dependencies pre-added (Fiber, MongoDB driver, samber/mo)
- CLAUDE.md               Coding conventions for this codebase
- principles.md           Architecture rationale (Ports & Adapters)

Read CLAUDE.md and principles.md before the session — the workshop builds on them.

--- Workshop goal ---

You will implement a /health endpoint that checks MongoDB connectivity,
following the Ports & Adapters structure described in the docs.

The reference implementation lives on the `minapi` branch. You can peek at it
any time with:

  git diff min..minapi

--- What `minapi` adds ---

  internal/health/
    domain/health.go              Health model
    port/port.go                  UseCase + Repository interfaces
    application/service.go        Orchestration service
    adapter/mongo_repository.go   MongoDB side-effect adapter (mo.IOEither)
    adapter/http/handler.go       Fiber HTTP handler (package httpadapter)
  compose.yml                     MongoDB via Docker Compose
  cmd/api/main.go                 Full wire-up, Fiber on :3000

--- Verify your setup (on the `minapi` branch) ---

  git checkout minapi
  docker compose up -d
  go run ./cmd/api
  curl http://localhost:3000/health
  # → {"status":"ok","mongo":"ok"}

See you at the workshop.
