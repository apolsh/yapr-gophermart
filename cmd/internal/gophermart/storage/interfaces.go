package storage

import (
	"context"
	"errors"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity"
)

type (
	UserStorage interface {
		NewUser(ctx context.Context, login, hashedPassword string) (string, error)
		Get(ctx context.Context, login string) (entity.User, error)
	}

	OrderStorage interface {
		Get(ctx context.Context) error
	}
)

var (
	ErrorLoginIsAlreadyUsed = errors.New("login is already used")
	ItemNotFound            = errors.New("requested element not found")
)
