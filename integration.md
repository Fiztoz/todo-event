# Integration Events

There are two kinds of events in this project:

## Domain Event

Happens *inside* a domain, handled in the same process.

```
task.created  →  SaveHandler (same domain, same process)
```

The task domain fires `task.created`. The `SaveHandler` reacts and saves to MongoDB. Nothing outside the task domain knows or cares.

## Integration Event

Crosses a *domain boundary*. One domain publishes, another domain subscribes.

```
task.created  →  NotificationDomain listens  →  sends an email
               → AnalyticsDomain listens     →  records a metric
```

The task domain still just calls `publisher.Publish("task.created", ...)` and stops. It has no idea who else is listening. The notification domain subscribes independently and reacts in its own way.

## Key Difference

| | Domain Event | Integration Event |
|---|---|---|
| Who handles it | Same domain | Different domain |
| Coupling | Tight (same module) | Loose (just the event contract) |
| Example | `SaveHandler` saves the task | `NotificationHandler` sends email |

## In This Codebase

`pkg/event.Bus` already supports this pattern. Wiring a cross-domain handler is just:

```go
// main.go — notification domain subscribes to task domain's event
bus.Subscribe("task.created", notificationadapter.NewTaskCreatedHandler(notifService))
```

The task domain's service does not change at all. The integration event *is* the same event — what makes it an "integration event" is that the subscriber lives in a different domain.

In a distributed system, integration events would travel over a message broker (Kafka, RabbitMQ). In our modular monolith, the same `pkg/event.Bus` serves both purposes since everything runs in one process.
