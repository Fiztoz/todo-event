# Step 03 — Change the Repository port: `Save` → `Append`

## What changes and why

The storage contract is the hinge between domain logic and infrastructure. The old
`Save(task)` method wrote a full snapshot document. The new `Append` method writes a
single named event carrying only the fields that changed.

This is the most consequential interface change in the migration: once it lands every
implementor (the real Mongo repository and any future in-memory test double) must
satisfy the new shape.

`GetTask` is added to `UseCase` at the same time because the HTTP detail endpoint
(Step 06) needs it. Adding it here keeps the port coherent: the interface reflects all
business operations before any implementation is touched.

## File: `domain/task/port/port.go`

### `UseCase` — add `GetTask`

```go
// Before
type UseCase interface {
	CreateTask(ctx context.Context, title string) mo.Result[domain.Task]
	ListTasks(ctx context.Context) mo.Result[[]domain.Task]
	ChangeStatus(ctx context.Context, id bson.ObjectID, status domain.Status) mo.Result[domain.Task]
}

// After
type UseCase interface {
	CreateTask(ctx context.Context, title string) mo.Result[domain.Task]
	ListTasks(ctx context.Context) mo.Result[[]domain.Task]
	GetTask(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task]
	ChangeStatus(ctx context.Context, id bson.ObjectID, status domain.Status) mo.Result[domain.Task]
}
```

### `Repository` — replace `Save` with `Append`

```go
// Before
type Repository interface {
	Save(ctx context.Context, task domain.Task) mo.Result[struct{}]
	FindAll(ctx context.Context) mo.Result[[]domain.Task]
	FindByID(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task]
}

// After
type Repository interface {
	Append(ctx context.Context, aggregateID bson.ObjectID, eventType string, payload any) mo.Result[struct{}]
	FindAll(ctx context.Context) mo.Result[[]domain.Task]
	FindByID(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task]
}
```

`FindAll` and `FindByID` signatures are unchanged.

## Breaking change notice

After this step the build is broken until Steps 04 and 05 are also applied:

- `MongoRepository` still implements `Save`, not `Append` → fixed in Step 04
- `Service` still calls `repo.Save` and does not implement `GetTask` → fixed in Step 05

If you need the build to stay green at every commit, apply Steps 03, 04, and 05 as a
single atomic commit.

## Checklist

- [ ] `GetTask` added to `UseCase`
- [ ] `Save` replaced by `Append` in `Repository`
- [ ] `FindAll` and `FindByID` signatures unchanged
