package service

import (
	"context"
	"errors"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/storage"
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

func (g GophermartServiceImpl) Get(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
