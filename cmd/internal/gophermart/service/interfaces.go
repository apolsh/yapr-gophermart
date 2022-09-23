package service

import "context"

type GophermartService interface {
	Get(ctx context.Context) error
}
