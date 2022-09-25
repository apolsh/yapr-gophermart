package service

import (
	"context"
	"errors"
)

type GophermartService interface {
	AddUser(ctx context.Context, login, password string) error
}

var ErrorEmptyValue = errors.New("empty values is not allowed")
