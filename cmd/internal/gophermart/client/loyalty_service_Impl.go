package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type LoyaltyServiceImpl struct {
	client  *resty.Client
	baseURL string
}

func NewLoyaltyServiceImpl(baseURL string) (LoyaltyService, error) {
	client := resty.New()
	return &LoyaltyServiceImpl{client: client, baseURL: baseURL}, nil
}

func (l LoyaltyServiceImpl) GetLoyaltyPoints(ctx context.Context, orderNum int) (LoyaltyPointsInfo, error) {
	loyaltyPointsRes := LoyaltyPointsInfo{}
	response, err := l.client.R().
		SetContext(ctx).
		SetResult(&loyaltyPointsRes).
		Get(fmt.Sprintf("%s/api/orders/%s", l.baseURL, strconv.Itoa(orderNum)))
	if err != nil {
		return loyaltyPointsRes, err
	}
	if response.StatusCode() == http.StatusNoContent {
		return loyaltyPointsRes, OrderIsNotRegisteredYet
	}
	if response.StatusCode() == http.StatusTooManyRequests {
		return loyaltyPointsRes, TooManyRequestsError
	}
	if response.StatusCode() == http.StatusInternalServerError {
		return loyaltyPointsRes, UnknownLoyaltyServiceError
	}

	return loyaltyPointsRes, nil
}
