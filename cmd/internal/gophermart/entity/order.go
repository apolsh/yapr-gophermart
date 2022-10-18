package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	StatusNew        string = "NEW"
	StatusProcessing        = "PROCESSING"
	StatusInvalid           = "INVALID"
	StatusProcessed         = "PROCESSED"
)

type Order struct {
	Number     int             `json:"number"`
	Status     string          `json:"status"`
	Accrual    decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time       `json:"uploaded_at"`
	UserId     string          `json:"-"`
}

func NewOrder(number int, userId string) Order {
	return Order{
		Number:     number,
		UserId:     userId,
		UploadedAt: time.Now(),
		Accrual:    decimal.NewFromInt(0),
		Status:     StatusNew,
	}
}
