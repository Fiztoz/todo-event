# CQRS Step 1 — Add GetTask to UseCase

## What changes and why

Read queries need a way to fetch a single task by ID. Without this, the only option is `ListTasks` which returns everything. Adding `GetTask` to the port completes the read surface and allows the `GET /tasks/:id` endpoint in the next step.

## Starting point: `domain/task/port/port.go`

```go
type UseCase interface {
    CreateTask(ctx context.Context, title string) mo.Result[domain.Task]
    ListTasks(ctx context.Context) mo.Result[[]domain.Task]
    ChangeStatus(ctx context.Context, id bson.ObjectID, status domain.Status) mo.Result[domain.Task]
}
```

## Target: add `GetTask`

```go
type UseCase interface {
    CreateTask(ctx context.Context, title string) mo.Result[domain.Task]
    ListTasks(ctx context.Context) mo.Result[[]domain.Task]
    GetTask(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task]
    ChangeStatus(ctx context.Context, id bson.ObjectID, status domain.Status) mo.Result[domain.Task]
}
```

## Key rules

- `port/` contains interfaces only — adding `GetTask` here does not change any adapter or service yet
- The compile-time check `var _ port.UseCase = (*Service)(nil)` in `service.go` will now fail until `GetTask` is implemented there

## Next step

Implement `GetTask` in `domain/task/application/service.go`.
