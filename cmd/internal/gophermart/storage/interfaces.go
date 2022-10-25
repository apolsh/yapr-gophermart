package storage

import (
	"context"
	"errors"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity"
	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity/dto"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

type (
	UserStorage interface {
		NewUser(ctx context.Context, login, hashedPassword string) (string, error)
		Get(ctx context.Context, login string) (entity.User, error)
	}

	OrderStorage interface {
		SaveNewOrder(ctx context.Context, orderNum string, userID string) error
		UpdateOrder(ctx context.Context, orderNum string, status string, accrual decimal.Decimal) error
		GetOrdersByID(ctx context.Context, id string) ([]entity.Order, error)
		GetBalanceByUserID(ctx context.Context, id string) (dto.Balance, error)
		CreateWithdraw(ctx context.Context, id string, withdraw dto.Withdraw) error
		GetWithdrawalsByUserID(ctx context.Context, id string) ([]dto.Withdraw, error)
	}
)

var (
	ErrorLoginIsAlreadyUsed          = errors.New("login is already used")
	ErrItemNotFound                  = errors.New("requested element not found")
	ErrOrderAlreadyStored            = errors.New("order is already uploaded by user")
	ErrOrderAlreadyStoredByOtherUser = errors.New("order is already uploaded by another user")
	ErrUnknownDatabase               = errors.New("unknown database error")
	ErrInsufficientFunds             = errors.New("insufficient funds to complete the operation")
)

func HandleUnknownDatabaseError(err error) error {
	log.Error().Err(err).Msg(err.Error())
	return ErrUnknownDatabase
}
