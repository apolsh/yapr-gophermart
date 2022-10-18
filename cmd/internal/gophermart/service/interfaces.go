package service

import (
	"context"
	"errors"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity"
)

type GophermartService interface {
	AddUser(ctx context.Context, login, password string) (string, error)
	LoginUser(ctx context.Context, login, password string) (string, error)
	ParseJWTToken(token string) (string, error)
	AddOrder(ctx context.Context, orderNum int, userId string) error
	GetOrdersByUser(ctx context.Context, id string) ([]entity.Order, error)
}

var (
	ErrorEmptyValue               = errors.New("empty values is not allowed")
	ErrorInvalidPassword          = errors.New("invalid password")
	ErrorInvalidOrderNumberFormat = errors.New("invalid order number format")
)
