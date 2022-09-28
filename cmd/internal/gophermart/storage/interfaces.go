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
		SaveOrder(ctx context.Context, order entity.Order) error
	}
)

var (
	ErrorLoginIsAlreadyUsed       = errors.New("login is already used")
	ItemNotFound                  = errors.New("requested element not found")
	OrderAlreadyStored            = errors.New("order is already uploaded by user")
	OrderAlreadyStoredByOtherUser = errors.New("order is already uploaded by another user")
)
