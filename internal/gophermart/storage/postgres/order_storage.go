package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	dto "github.com/apolsh/yapr-gophermart/internal/gophermart/dto"
	"github.com/apolsh/yapr-gophermart/internal/gophermart/storage"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shopspring/decimal"
)

type OrderStoragePG struct {
	pool *pgxpool.Pool
}

const (
	constraintUniqOrderNumber    = "order_pk"
	constraintNonNegativeBalance = "current_non_negative"
)

func NewOrderStoragePG(pool *pgxpool.Pool) *OrderStoragePG {
	return &OrderStoragePG{pool: pool}
}

func (o *OrderStoragePG) SaveNewOrder(ctx context.Context, orderNum string, userID string) error {
	order := dto.NewOrder(orderNum, userID)
	//language=postgresql
	s := "INSERT INTO \"order\" (number, status, accrual, uploaded_at, user_id) VALUES ($1, $2, $3, $4, $5)"
	_, err := o.pool.Exec(ctx, s, orderNum, order.Status, order.Accrual, order.UploadedAt, order.UserID)
	var pgErr *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == constraintUniqOrderNumber {
				var userID string
				//language=postgresql
				q := "SELECT user_id FROM \"order\" WHERE number = $1"
				err := o.pool.QueryRow(ctx, q, order.Number).Scan(&userID)
				if err != nil {
					return fmt.Errorf("error during saving new order %s, cause: %w", orderNum, err)
				}

				if userID == order.UserID {
					return storage.ErrOrderAlreadyStored
				}
				return storage.ErrOrderAlreadyStoredByOtherUser
			}
		}
		return fmt.Errorf("error during saving new order %s, cause: %w", orderNum, err)
	}
	return nil
}

func (o *OrderStoragePG) UpdateOrder(ctx context.Context, orderNum string, status string, accrual decimal.Decimal) error {
	//language=postgresql
	q := "UPDATE \"order\" SET status = $1, accrual = $2 WHERE number = $3"
	_, err := o.pool.Exec(ctx, q, status, accrual, orderNum)
	if err != nil {
		return fmt.Errorf("error during updating order %s, cause: %w", orderNum, err)
	}
	return nil
}

func (o *OrderStoragePG) GetOrdersByID(ctx context.Context, id string) ([]dto.Order, error) {
	//language=postgresql
	q := "SELECT number, status, accrual, uploaded_at, user_id FROM \"order\" WHERE user_id = $1"

	rows, err := o.pool.Query(ctx, q, id)
	if err != nil {
		return nil, err
	}

	orders := make([]dto.Order, 0)
	var order dto.Order
	for rows.Next() {
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt, &order.UserID)
		if err != nil {
			return nil, fmt.Errorf("error during recieving orders of user %s, cause: %w", id, err)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (o *OrderStoragePG) GetBalanceByUserID(ctx context.Context, id string) (dto.Balance, error) {
	//language=postgresql
	q := "SELECT current, withdrawn FROM \"balance\" WHERE user_id = $1"
	var balance dto.Balance

	err := o.pool.QueryRow(ctx, q, id).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return balance, fmt.Errorf("error during recieving balance of user %s, cause: %w", id, err)
	}
	return balance, nil
}

func (o *OrderStoragePG) CreateWithdraw(ctx context.Context, id string, withdraw dto.Withdraw) error {
	//language=postgresql
	q := "INSERT INTO withdrawal (\"order\", sum, processed_at, user_id) VALUES ($1, $2, $3, $4)"
	_, err := o.pool.Exec(ctx, q, withdraw.Order, withdraw.Sum, time.Now(), id)

	var pgErr *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == constraintNonNegativeBalance {
				return storage.ErrInsufficientFunds
			}
		}
		return fmt.Errorf("error during creating withdrawal of user %s, cause: %w", id, err)
	}
	return nil
}

func (o *OrderStoragePG) GetWithdrawalsByUserID(ctx context.Context, id string) ([]dto.Withdraw, error) {
	//language=postgresql
	q := "SELECT \"order\", sum, processed_at FROM withdrawal WHERE user_id = $1"
	rows, err := o.pool.Query(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("error during recieving withdrawals of user %s, cause: %w", id, err)
	}

	withdrawals := make([]dto.Withdraw, 0)
	var withdraw dto.Withdraw
	for rows.Next() {
		err := rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("error during recieving orders of user %s, cause: %w", id, err)
		}
		withdrawals = append(withdrawals, withdraw)
	}
	return withdrawals, nil
}

func (o *OrderStoragePG) GetAllUnfinishedAccrualOrderNums(ctx context.Context) ([]string, error) {
	q := "SELECT number FROM \"order\" WHERE status = $1"
	rows, err := o.pool.Query(ctx, q, dto.StatusNew)
	if err != nil {
		return nil, fmt.Errorf("error during recieving unfinished accrual order ids cause: %w", err)
	}

	orders := make([]string, 0)
	var orderNum string
	for rows.Next() {
		err := rows.Scan(&orderNum)
		if err != nil {
			return nil, fmt.Errorf("error during recieving unfinished accrual order ids cause: %w", err)
		}
		orders = append(orders, orderNum)
	}

	return orders, nil
}
