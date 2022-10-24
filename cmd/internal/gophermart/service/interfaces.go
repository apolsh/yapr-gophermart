package service

import (
	"context"
	"errors"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity"
	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity/dto"
)

type GophermartService interface {
	AddUser(ctx context.Context, login, password string) (string, error)
	LoginUser(ctx context.Context, login, password string) (string, error)
	ParseJWTToken(token string) (string, error)
	AddOrder(ctx context.Context, orderNum int, userId string) error
	GetOrdersByUser(ctx context.Context, id string) ([]entity.Order, error)
	GetBalanceByUserID(ctx context.Context, id string) (dto.Balance, error)
	CreateWithdraw(ctx context.Context, id string, withdraw dto.Withdraw) error
	GetWithdrawalsByUserID(ctx context.Context, id string) ([]dto.Withdraw, error)
}

var (
	ErrorEmptyValue               = errors.New("empty values is not allowed")
	ErrorInvalidPassword          = errors.New("invalid password")
	ErrorInvalidOrderNumberFormat = errors.New("invalid order number format")
)
