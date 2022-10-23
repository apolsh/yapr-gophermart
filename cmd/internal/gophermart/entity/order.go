package entity

import (
	"strconv"
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
	Number     string          `json:"number"`
	Status     string          `json:"status"`
	Accrual    decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time       `json:"uploaded_at"`
	UserId     string          `json:"-"`
}

func NewOrder(number int, userId string) Order {
	return Order{
		Number:     strconv.Itoa(number),
		UserId:     userId,
		UploadedAt: time.Now(),
		Status:     StatusNew,
	}
}
