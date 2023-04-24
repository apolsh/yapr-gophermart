//go:generate mockgen -destination=../mocks/service.go -package=mocks github.com/apolsh/yapr-gophermart/internal/gophermart/service UserStorage,OrderStorage
package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	loyaltyHTTPClient "github.com/apolsh/yapr-gophermart/internal/gophermart/client"
	entity "github.com/apolsh/yapr-gophermart/internal/gophermart/dto"
	"github.com/apolsh/yapr-gophermart/internal/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

var serviceLogger = logger.LoggerOfComponent("gophmarket_service_logger")

type (
	UserStorage interface {
		NewUser(ctx context.Context, login, hashedPassword string) (string, error)
		Get(ctx context.Context, login string) (entity.User, error)
	}

	OrderStorage interface {
		SaveNewOrder(ctx context.Context, orderNum string, userID string) error
		UpdateOrder(ctx context.Context, orderNum string, status string, accrual decimal.Decimal) error
		GetOrdersByID(ctx context.Context, id string) ([]entity.Order, error)
		GetBalanceByUserID(ctx context.Context, id string) (entity.Balance, error)
		CreateWithdraw(ctx context.Context, id string, withdraw entity.Withdraw) error
		GetWithdrawalsByUserID(ctx context.Context, id string) ([]entity.Withdraw, error)
		GetAllUnfinishedAccrualOrderNums(ctx context.Context) ([]string, error)
	}
)

type GophermartServiceImpl struct {
	jwtSecretKey        string
	userStorage         UserStorage
	orderStorage        OrderStorage
	loyaltyService      loyaltyHTTPClient.LoyaltyService
	asyncWorker         *AsyncWorker
	AsyncWorkerMaxTries int
}

func (g *GophermartServiceImpl) Close() {
	g.asyncWorker.Close()
}

type jwtTokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

func NewGophermartServiceImpl(
	jwtSecretKey string,
	asyncWorkerMaxTries int,
	accrualSystemAddress string,
	userStorage UserStorage,
	orderStorage OrderStorage) (*GophermartServiceImpl, error) {

	if userStorage == nil || orderStorage == nil {
		return nil, errors.New("not all storages were initialized")
	}

	loyaltyService, err := loyaltyHTTPClient.NewLoyaltyServiceImpl(accrualSystemAddress)
	if err != nil {
		return nil, fmt.Errorf("failed during loyalty client for gophmart service init: %w", err)
	}

	return &GophermartServiceImpl{
		jwtSecretKey:        jwtSecretKey,
		userStorage:         userStorage,
		orderStorage:        orderStorage,
		loyaltyService:      loyaltyService,
		AsyncWorkerMaxTries: asyncWorkerMaxTries,
	}, nil
}

func (g *GophermartServiceImpl) AddUser(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", ErrorEmptyValue
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error during generating hashed password for user %s, cause: %w", login, err)
	}

	id, err := g.userStorage.NewUser(ctx, login, string(hashedPassword))
	if err != nil {
		return "", fmt.Errorf("error during saving new user: %s, cause: %w", login, err)
	}
	return g.generateToken(id)
}

func (g *GophermartServiceImpl) LoginUser(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", ErrorEmptyValue
	}

	user, err := g.userStorage.Get(ctx, login)
	if err != nil {
		return "", fmt.Errorf("error during recieving user: %s, cause: %w", login, err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		return "", ErrorInvalidPassword
	}

	return g.generateToken(user.ID)
}

func (g *GophermartServiceImpl) ParseJWTToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(g.jwtSecretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("error during parsing jwt token %w", err)
	}

	claims, ok := token.Claims.(*jwtTokenClaims)
	if !ok {
		return "", errors.New("invalid token claims type")
	}
	return claims.UserID, nil
}

func (g *GophermartServiceImpl) AddOrder(ctx context.Context, orderNum string, userID string) error {
	err := validateOrderFormat(orderNum)
	if err != nil {
		return fmt.Errorf("error during validating order format: %w", err)
	}

	err = g.orderStorage.SaveNewOrder(ctx, orderNum, userID)
	if err != nil {
		return fmt.Errorf("error during saving new order: %w", err)
	}

	if g.asyncWorker != nil {
		g.asyncWorker.ExecuteTask(func() {
			g.getAccrualAsync(0, orderNum)
		})
	}

	return nil
}

func (g *GophermartServiceImpl) GetBalanceByUserID(ctx context.Context, id string) (entity.Balance, error) {
	balance, err := g.orderStorage.GetBalanceByUserID(ctx, id)
	if err != nil {
		return entity.Balance{}, fmt.Errorf("error during recieving balance of user: %s, cause %w", id, err)
	}
	return balance, nil
}

func (g *GophermartServiceImpl) CreateWithdraw(ctx context.Context, id string, withdraw entity.Withdraw) error {
	err := validateOrderFormat(withdraw.Order)
	if err != nil {
		return fmt.Errorf("error during validating order format: %w", err)
	}
	err = g.orderStorage.CreateWithdraw(ctx, id, withdraw)
	if err != nil {
		return fmt.Errorf("error during creating withdrawal for user %s, cause %w", id, err)
	}
	return nil
}

func (g *GophermartServiceImpl) GetWithdrawalsByUserID(ctx context.Context, id string) ([]entity.Withdraw, error) {
	withdraw, err := g.orderStorage.GetWithdrawalsByUserID(ctx, id)
	if err != nil {
		return []entity.Withdraw{}, fmt.Errorf("error during recieving withdrawals of user: %s, cause %w", id, err)
	}
	return withdraw, nil
}

func (g *GophermartServiceImpl) GetOrdersByUser(ctx context.Context, id string) ([]entity.Order, error) {
	orders, err := g.orderStorage.GetOrdersByID(ctx, id)
	if err != nil {
		return []entity.Order{}, fmt.Errorf("error during recieving orders of user: %s, cause %w", id, err)
	}
	return orders, nil
}

func (g *GophermartServiceImpl) StartAccrualInfoSynchronizer(ctx context.Context, loyaltyServiceRateLimit int) error {
	asyncWorker, err := NewAsyncWorker(loyaltyServiceRateLimit)
	if err != nil {
		return fmt.Errorf("failed during async worker for gophmart service init: %w", err)
	}
	g.asyncWorker = asyncWorker

	orders, err := g.orderStorage.GetAllUnfinishedAccrualOrderNums(ctx)
	if err != nil {
		return fmt.Errorf("failed to recieve unfinished accrual order nums: %w", err)
	}

	for _, order := range orders {
		g.asyncWorker.ExecuteTask(func() {
			g.getAccrualAsync(0, order)
		})
	}
	return nil
}

func validateOrderFormat(orderNumAsString string) error {
	orderNum, err := strconv.Atoi(orderNumAsString)
	if err != nil {
		return ErrorInvalidOrderNumberFormat
	}

	var checksum int

	for i := 1; orderNum > 0; i++ {
		num := orderNum % 10
		if i%2 == 0 {
			num = num * 2
			if num > 9 {
				num = num%10 + num/10
			}
		}
		checksum += num
		orderNum = orderNum / 10
	}
	if checksum%10 != 0 {
		return ErrorInvalidOrderNumberFormat
	}
	return nil
}

func (g *GophermartServiceImpl) generateToken(id string) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwtTokenClaims{
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		id,
	})

	return token.SignedString([]byte(g.jwtSecretKey))
}

func (g *GophermartServiceImpl) getAccrualAsync(numOfTry int, orderNum string) {
	if numOfTry > g.AsyncWorkerMaxTries {
		log.Err(fmt.Errorf("maximum number of attempts to retrieve accrual info of order %s has been exceeded", orderNum))
		return
	}
	numOfTry++
	loyaltyInfo, err := g.loyaltyService.GetLoyaltyPoints(context.Background(), orderNum)
	if errors.Is(loyaltyHTTPClient.ErrTooManyRequests, err) {
		time.AfterFunc(1*time.Minute, func() {
			g.asyncWorker.ExecuteTask(func() {
				g.getAccrualAsync(numOfTry, orderNum)
			})
		})
	}
	if errors.Is(loyaltyHTTPClient.ErrOrderIsNotRegisteredYet, err) {
		time.AfterFunc(15*time.Second, func() {
			g.asyncWorker.ExecuteTask(func() {
				g.getAccrualAsync(numOfTry, orderNum)
			})
		})
	}

	if err != nil {
		serviceLogger.Error(fmt.Errorf("failed to recieve loyality points info: %w", err))
		return
	}

	if loyaltyInfo.Status == loyaltyHTTPClient.StatusProcessing || loyaltyInfo.Status == loyaltyHTTPClient.StatusRegistered {
		time.AfterFunc(15*time.Second, func() {
			g.asyncWorker.ExecuteTask(func() {
				g.getAccrualAsync(numOfTry, orderNum)
			})
		})
	}
	err = g.orderStorage.UpdateOrder(context.Background(), orderNum, loyaltyInfo.Status, loyaltyInfo.Accrual)
	if err != nil {
		serviceLogger.Error(fmt.Errorf("failed to update order: %s, cause: %w", orderNum, err))
		return
	}
}
