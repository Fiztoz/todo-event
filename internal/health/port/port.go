package port

import (
	"context"

	"github.com/samber/mo"
	"todoe/internal/health/domain"
)

type UseCase interface {
	CheckHealth(ctx context.Context) mo.Result[domain.Health]
}

type Repository interface {
	Ping(ctx context.Context) mo.Result[struct{}]
}
