package application

import (
	"context"
	"errors"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/domain/task/domain"
	"todoe/domain/task/port"
)

var (
	ErrInvalidTitle  = errors.New("title must not be empty")
	ErrInvalidStatus = errors.New("invalid status")
)

type Service struct {
	repo port.Repository
}

var _ port.UseCase = (*Service)(nil)

func NewService(repo port.Repository) *Service {
	return &Service{repo: repo}
}

func validateTitle(title string) error {
	if title == "" {
		return ErrInvalidTitle
	}
	return nil
}

func validateStatus(s domain.Status) error {
	switch s {
	case domain.StatusPending, domain.StatusInProgress, domain.StatusDone:
		return nil
	}
	return ErrInvalidStatus
}

func (s *Service) CreateTask(ctx context.Context, title string) mo.Result[domain.Task] {
	if err := validateTitle(title); err != nil {
		return mo.Err[domain.Task](err)
	}
	task := domain.NewTask(title)
	if result := s.repo.Save(ctx, task); result.IsError() {
		return mo.Err[domain.Task](result.Error())
	}
	return mo.Ok(task)
}

func (s *Service) ListTasks(ctx context.Context) mo.Result[[]domain.Task] {
	return s.repo.FindAll(ctx)
}

func (s *Service) ChangeStatus(ctx context.Context, id bson.ObjectID, status domain.Status) mo.Result[domain.Task] {
	if err := validateStatus(status); err != nil {
		return mo.Err[domain.Task](err)
	}
	current := s.repo.FindByID(ctx, id)
	if current.IsError() {
		return mo.Err[domain.Task](current.Error())
	}
	next := current.MustGet().ChangeStatus(status)
	if result := s.repo.Save(ctx, next); result.IsError() {
		return mo.Err[domain.Task](result.Error())
	}
	return mo.Ok(next)
}
