package port

import (
	"context"

	"github.com/samber/mo"
	"todoe/domain/task/domain"
)

type UseCase interface {
	CreateTask(ctx context.Context, title string) mo.Result[domain.Task]
	ListTasks(ctx context.Context) mo.Result[[]domain.Task]
}

type Repository interface {
	Save(ctx context.Context, task domain.Task) mo.Result[struct{}]
	FindAll(ctx context.Context) mo.Result[[]domain.Task]
}

type Publisher interface {
	Publish(name string, payload any)
}
