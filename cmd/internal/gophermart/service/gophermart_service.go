package service

import (
	"context"
	"errors"
	"time"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity"
	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/storage"
	"github.com/apolsh/yapr-gophermart/config"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type GophermartServiceImpl struct {
	jwtSecretKey string
	userStorage  storage.UserStorage
	orderStorage storage.OrderStorage
}

type jwtTokenClaims struct {
	jwt.RegisteredClaims
	UserId string `json:"user_id"`
}

func NewGophermartServiceImpl(cfg config.Config, userStorage storage.UserStorage, orderStorage storage.OrderStorage) (GophermartService, error) {

	if userStorage == nil || orderStorage == nil {
		return nil, errors.New("not all storages were initialized")
	}

	return &GophermartServiceImpl{jwtSecretKey: cfg.TokenSecretKey, userStorage: userStorage, orderStorage: orderStorage}, nil
}

func (g GophermartServiceImpl) AddUser(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", ErrorEmptyValue
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	id, err := g.userStorage.NewUser(ctx, login, string(hashedPassword))
	if err != nil {
		return "", err
	}
	return g.generateToken(id)
}

func (g GophermartServiceImpl) LoginUser(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", ErrorEmptyValue
	}

	user, err := g.userStorage.Get(ctx, login)
	if err != nil {
		return "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		return "", ErrorInvalidPassword
	}

	return g.generateToken(user.Id)
}

func (g GophermartServiceImpl) ParseJWTToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(g.jwtSecretKey), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwtTokenClaims)
	if !ok {
		return "", errors.New("invalid token claims type")
	}
	return claims.UserId, nil
}

func (g GophermartServiceImpl) AddOrder(ctx context.Context, orderNum int, userId string) error {
	err := validateOrderFormat(orderNum)
	if err != nil {
		return err
	}

	return g.orderStorage.SaveOrder(ctx, orderNum, userId)
}

func (g GophermartServiceImpl) GetOrdersByUser(ctx context.Context, id string) ([]entity.Order, error) {
	return g.orderStorage.GetOrdersByID(ctx, id)
}

func validateOrderFormat(orderNum int) error {
	//simplified Luna algorithm
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

func (g GophermartServiceImpl) generateToken(id string) (string, error) {
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
