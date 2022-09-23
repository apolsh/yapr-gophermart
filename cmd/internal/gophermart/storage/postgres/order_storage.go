package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderStoragePG struct {
	pool *pgxpool.Pool
}

func NewOrderStoragePG(pool *pgxpool.Pool) *OrderStoragePG {
	return &OrderStoragePG{pool: pool}
}

func (o OrderStoragePG) Get(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
