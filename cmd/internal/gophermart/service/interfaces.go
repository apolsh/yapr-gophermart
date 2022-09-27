package service

import (
	"context"
	"errors"
)

type GophermartService interface {
	AddUser(ctx context.Context, login, password string) (string, error)
	LoginUser(ctx context.Context, login, password string) (string, error)
	ParseJWTToken(token string) (string, error)
}

var ErrorEmptyValue = errors.New("empty values is not allowed")
var ErrorInvalidPassword = errors.New("invalid password")
