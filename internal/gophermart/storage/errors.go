package storage

import (
	"errors"
)

var (
	ErrorLoginIsAlreadyUsed          = errors.New("login is already used")
	ErrItemNotFound                  = errors.New("requested element not found")
	ErrOrderAlreadyStored            = errors.New("order is already uploaded by user")
	ErrOrderAlreadyStoredByOtherUser = errors.New("order is already uploaded by another user")
	ErrInsufficientFunds             = errors.New("insufficient funds to complete the operation")
)
