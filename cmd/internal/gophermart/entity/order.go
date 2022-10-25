package entity

import (
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

type Order struct {
	Number     string          `json:"number"`
	Status     string          `json:"status"`
	Accrual    decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time       `json:"uploaded_at"`
	UserID     string          `json:"-"`
}

func NewOrder(number int, userID string) Order {
	return Order{
		Number:     strconv.Itoa(number),
		UserID:     userID,
		UploadedAt: time.Now(),
		Status:     StatusNew,
	}
}
