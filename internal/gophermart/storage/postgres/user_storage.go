package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/apolsh/yapr-gophermart/internal/gophermart/dto"
	"github.com/apolsh/yapr-gophermart/internal/gophermart/storage"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
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

func (s *UserStoragePG) NewUser(ctx context.Context, login, hashedPassword string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, "INSERT INTO \"user\" (login, password) VALUES ($1, $2)  RETURNING id", login, hashedPassword).Scan(&id)

	var pgErr *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == constraintUniqLogin {
				return "", storage.ErrorLoginIsAlreadyUsed
			}
		}
		return "", fmt.Errorf("storage error while creating user %s, cause: %w", login, err)
	}

	return id, nil
}

func (s *UserStoragePG) Get(ctx context.Context, login string) (dto.User, error) {
	q := "SELECT id, login, password from \"user\" WHERE login = $1"
	var user dto.User
	err := s.pool.QueryRow(ctx, q, login).Scan(&user.ID, &user.Login, &user.HashedPassword)
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return dto.User{}, storage.ErrItemNotFound
		}
		return dto.User{}, fmt.Errorf("storage error while getting user %s, cause: %w", login, err)
	}
	return user, nil
}
