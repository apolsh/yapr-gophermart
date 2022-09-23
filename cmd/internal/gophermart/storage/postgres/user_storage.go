package postgres

import (
	"context"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStoragePG struct {
	pool *pgxpool.Pool
}

func NewUserStoragePG(pool *pgxpool.Pool) *UserStoragePG {
	return &UserStoragePG{pool: pool}
}

func (P UserStoragePG) Get(ctx context.Context, login, hashedPassword string) (entity.User, error) {
	//TODO implement me
	panic("implement me")
	return entity.User{}, nil
}
