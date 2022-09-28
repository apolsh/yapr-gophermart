package service

import (
	"context"
	"errors"
)

type GophermartService interface {
	AddUser(ctx context.Context, login, password string) (string, error)
	LoginUser(ctx context.Context, login, password string) (string, error)
	ParseJWTToken(token string) (string, error)
	AddOrder(ctx context.Context, orderNum int, userId string) error
}

var (
	ErrorEmptyValue               = errors.New("empty values is not allowed")
	ErrorInvalidPassword          = errors.New("invalid password")
	ErrorInvalidOrderNumberFormat = errors.New("invalid order number format")
)
