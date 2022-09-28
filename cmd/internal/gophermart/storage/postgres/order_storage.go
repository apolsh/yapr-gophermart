package postgres

import (
	"context"
	"errors"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity"
	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/storage"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderStoragePG struct {
	pool *pgxpool.Pool
}

const (
	constraintUniqOrderNumber = "order_pk"
)

func NewOrderStoragePG(pool *pgxpool.Pool) *OrderStoragePG {
	return &OrderStoragePG{pool: pool}
}

func (o OrderStoragePG) SaveOrder(ctx context.Context, order entity.Order) error {
	s := "INSERT INTO order (number, status, accrual, uploaded_at, user_id) VALUES ($1, $2, $3, $4, $5)"
	_, err := o.pool.Exec(ctx, s, order.Number, order.Status, order.Accrual, order.UploadedAt, order.UserId)
	var pgErr *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == constraintUniqOrderNumber {
				var userId string
				q := "SELECT user_id FROM order WHERE number = $1"
				err := o.pool.QueryRow(ctx, q, order.Number).Scan(&userId)
				if err != nil {
					//TODO: throw err add logging
				}

				if userId == order.UserId {
					return storage.OrderAlreadyStored
				}
				return storage.OrderAlreadyStoredByOtherUser
			}
			//TODO: throw err add logging
		}
	}
	return nil
}
