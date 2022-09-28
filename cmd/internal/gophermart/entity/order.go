package entity

import (
	"time"
)

type Order struct {
	Number     int       `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual,omitempty"` //TODO replace
	UploadedAt time.Time `json:"uploaded_at"`
	UserId     string    `json:"-"`
}

func NewOrder(number int, userId string) Order {
	return Order{
		Number: number,
		//TODO: implement me
	}
}
