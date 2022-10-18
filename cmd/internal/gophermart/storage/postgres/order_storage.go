package postgres

import (
	"context"
	"errors"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity"
	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/storage"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderStoragePG struct {
	pool *pgxpool.Pool
}

const (
	constraintUniqOrderNumber = "order_pk"
)

func NewOrderStoragePG(pool *pgxpool.Pool) storage.OrderStorage {

	return &OrderStoragePG{pool: pool}
}

func (o OrderStoragePG) SaveOrder(ctx context.Context, orderNum int, userID string) error {
	order := entity.NewOrder(orderNum, userID)
	//language=postgresql
	s := "INSERT INTO \"order\" (number, status, accrual, uploaded_at, user_id) VALUES ($1, $2, $3, $4, $5)"
	_, err := o.pool.Exec(ctx, s, order.Number, order.Status, pgxdecimal.Decimal(order.Accrual), order.UploadedAt, order.UserId)
	var pgErr *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == constraintUniqOrderNumber {
				var userId string
				//language=postgresql
				q := "SELECT user_id FROM \"order\" WHERE number = $1"
				err := o.pool.QueryRow(ctx, q, order.Number).Scan(&userId)
				if err != nil {
					return storage.HandleUnknownDatabaseError(err)
				}

				if userId == order.UserId {
					return storage.OrderAlreadyStored
				}
				return storage.OrderAlreadyStoredByOtherUser
			}
		}
		return storage.HandleUnknownDatabaseError(err)
	}
	return nil
}

func (o OrderStoragePG) GetOrdersByID(ctx context.Context, id string) ([]entity.Order, error) {
	q := "SELECT number, status, accrual, uploaded_at, user_id FROM \"order\" WHERE user_id = $1"

	rows, err := o.pool.Query(ctx, q, id)
	if err != nil {
		return nil, err
	}

	orders := make([]entity.Order, 0)
	var order entity.Order
	for rows.Next() {
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt, &order.UserId)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
