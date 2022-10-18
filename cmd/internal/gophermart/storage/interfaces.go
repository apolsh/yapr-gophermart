package storage

import (
	"context"
	"errors"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity"
	"github.com/rs/zerolog/log"
)

type (
	UserStorage interface {
		NewUser(ctx context.Context, login, hashedPassword string) (string, error)
		Get(ctx context.Context, login string) (entity.User, error)
	}

	OrderStorage interface {
		SaveOrder(ctx context.Context, orderNum int, userID string) error
		GetOrdersByID(ctx context.Context, id string) ([]entity.Order, error)
	}
)

var (
	ErrorLoginIsAlreadyUsed       = errors.New("login is already used")
	ItemNotFound                  = errors.New("requested element not found")
	OrderAlreadyStored            = errors.New("order is already uploaded by user")
	OrderAlreadyStoredByOtherUser = errors.New("order is already uploaded by another user")
	UnknownDatabaseError          = errors.New("unknown database error")
)

func HandleUnknownDatabaseError(err error) error {
	log.Error().Err(err).Msg(err.Error())
	return UnknownDatabaseError
}
