package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type Withdraw struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}
