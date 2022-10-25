package client

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
)

type LoyaltyService interface {
	GetLoyaltyPoints(ctx context.Context, orderNum string) (LoyaltyPointsInfo, error)
}

type LoyaltyPointsInfo struct {
	Order   string          `json:"order"`
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

var (
	ErrTooManyRequests         = errors.New("loyalty service responded: too many requests")
	ErrUnknownLoyaltyService   = errors.New("unknown loyalty service error")
	ErrOrderIsNotRegisteredYet = errors.New("order is not registered yet")
)
