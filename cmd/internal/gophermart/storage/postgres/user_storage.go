package postgres

import (
	"context"
	"errors"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/storage"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStoragePG struct {
	pool *pgxpool.Pool
}

const (
	constraintUniqLogin = "user_login_uindex"
)

func NewUserStoragePG(pool *pgxpool.Pool) *UserStoragePG {
	return &UserStoragePG{pool: pool}
}

func (s UserStoragePG) NewUser(ctx context.Context, login, hashedPassword string) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO \"user\" (login, password) VALUES ($1, $2)", login, hashedPassword)
	var pgErr *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == constraintUniqLogin {
				return storage.ErrorLoginIsAlreadyUsed
			}
		}
	}

	return nil
}
