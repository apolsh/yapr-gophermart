package storage

import (
	"context"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity"
)

type (
	UserStorage interface {
		Get(ctx context.Context, login, hashedPassword string) (user entity.User, err error)
	}

	OrderStorage interface {
		Get(ctx context.Context) error
	}
)
