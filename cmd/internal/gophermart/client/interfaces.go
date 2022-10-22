package client

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
)

type LoyaltyService interface {
	GetLoyaltyPoints(ctx context.Context, orderNum int) (LoyaltyPointsInfo, error)
}

type LoyaltyPointsInfo struct {
	Order   string          `json:"order"`
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

var (
	TooManyRequestsError       = errors.New("loyalty service responded: too many requests")
	UnknownLoyaltyServiceError = errors.New("unknown loyalty service error")
)
