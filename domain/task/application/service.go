package application

import (
	"context"
	"errors"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/domain/task/domain"
	"todoe/domain/task/port"
)

var ErrInvalidTitle = errors.New("title must not be empty")

type Service struct {
	repo      port.Repository
	publisher port.Publisher
}

var _ port.UseCase = (*Service)(nil)

func NewService(repo port.Repository, publisher port.Publisher) *Service {
	return &Service{repo: repo, publisher: publisher}
}

func validateTitle(title string) error {
	if title == "" {
		return ErrInvalidTitle
	}
	return nil
}

func (s *Service) CreateTask(ctx context.Context, title string) mo.Result[domain.Task] {
	if err := validateTitle(title); err != nil {
		return mo.Err[domain.Task](err)
	}
	task := domain.Task{
		ID:        bson.NewObjectID(),
		Title:     title,
		CreatedAt: time.Now(),
	}
	s.publisher.Publish(domain.EventCreated, domain.CreatedPayload{Task: task})
	return mo.Ok(task)
}

func (s *Service) ListTasks(ctx context.Context) mo.Result[[]domain.Task] {
	return s.repo.FindAll(ctx)
}
