package service

import (
	"context"
	"errors"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/storage"
	"golang.org/x/crypto/bcrypt"
)

type GophermartServiceImpl struct {
	userStorage  storage.UserStorage
	orderStorage storage.OrderStorage
}

func NewGophermartServiceImpl(userStorage storage.UserStorage, orderStorage storage.OrderStorage) (GophermartService, error) {

	if userStorage == nil || orderStorage == nil {
		return nil, errors.New("not all storages were initialized")
	}

	return &GophermartServiceImpl{userStorage: userStorage, orderStorage: orderStorage}, nil
}

func (g GophermartServiceImpl) AddUser(ctx context.Context, login, password string) error {
	if login == "" || password == "" {
		return ErrorEmptyValue
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return g.userStorage.NewUser(ctx, login, string(hashedPassword))
}
