package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type LoyaltyServiceImpl struct {
	client  *resty.Client
	baseURL string
}

const (
	StatusRegistered = "REGISTERED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
)

func NewLoyaltyServiceImpl(baseURL string) (LoyaltyService, error) {
	client := resty.New()
	return &LoyaltyServiceImpl{client: client, baseURL: baseURL}, nil
}

func (l LoyaltyServiceImpl) GetLoyaltyPoints(ctx context.Context, orderNum string) (LoyaltyPointsInfo, error) {
	loyaltyPointsRes := LoyaltyPointsInfo{}
	response, err := l.client.R().
		SetContext(ctx).
		SetResult(&loyaltyPointsRes).
		Get(fmt.Sprintf("%s/api/orders/%s", l.baseURL, orderNum))
	if err != nil {
		return loyaltyPointsRes, err
	}
	if response.StatusCode() == http.StatusNoContent {
		return loyaltyPointsRes, ErrOrderIsNotRegisteredYet
	}
	if response.StatusCode() == http.StatusTooManyRequests {
		return loyaltyPointsRes, ErrTooManyRequests
	}
	if response.StatusCode() == http.StatusInternalServerError {
		return loyaltyPointsRes, ErrUnknownLoyaltyService
	}

	return loyaltyPointsRes, nil
}
