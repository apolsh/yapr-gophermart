package service

import (
	"errors"
)

var (
	ErrorEmptyValue               = errors.New("empty values is not allowed")
	ErrorInvalidPassword          = errors.New("invalid password")
	ErrorInvalidOrderNumberFormat = errors.New("invalid order number format")
)
