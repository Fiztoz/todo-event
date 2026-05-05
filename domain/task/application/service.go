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

var (
	ErrInvalidTitle  = errors.New("title must not be empty")
	ErrInvalidStatus = errors.New("status must be one of: pending, in_progress, done")
)

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
		Status:    domain.StatusPending,
		CreatedAt: time.Now(),
	}
	s.publisher.Publish(domain.EventCreated, domain.CreatedPayload{Task: task})
	return mo.Ok(task)
}

func (s *Service) ListTasks(ctx context.Context) mo.Result[[]domain.Task] {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetTask(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task] {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ChangeStatus(ctx context.Context, id bson.ObjectID, status domain.Status) mo.Result[domain.Task] {
	if !status.IsValid() {
		return mo.Err[domain.Task](ErrInvalidStatus)
	}
	current := s.repo.FindByID(ctx, id)
	if current.IsError() {
		return mo.Err[domain.Task](current.Error())
	}
	next := current.MustGet()
	next.Status = status
	s.publisher.Publish(domain.EventStatusChanged, domain.StatusChangedPayload{TaskID: id, Status: status})
	return mo.Ok(next)
}
