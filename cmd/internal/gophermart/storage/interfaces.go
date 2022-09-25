package storage

import (
	"context"
	"errors"
)

type (
	UserStorage interface {
		NewUser(ctx context.Context, login, hashedPassword string) error
	}

	OrderStorage interface {
		Get(ctx context.Context) error
	}
)

var ErrorLoginIsAlreadyUsed = errors.New("login is already used")
