package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type UserRegisterSuite struct {
	suite.Suite
	db         *pgxpool.Pool
	client     *resty.Client
	authCookie string
}

type Order struct {
	Number     string          `json:"number"`
	Status     string          `json:"status"`
	Accrual    decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time       `json:"uploaded_at"`
}

const baseUserURL = "http://localhost:8080/api/user"
const loginURL = baseUserURL + "/login"
const ordersURL = baseUserURL + "/orders"

func TestUserRegisterSuite(t *testing.T) {
	suite.Run(t, new(UserRegisterSuite))
}

var db *pgxpool.Pool
var client *resty.Client
var authCookie string

func (s *UserRegisterSuite) SetupSuite() {
	s.client = resty.New()
	ctx := context.Background()
	dbURI := os.Getenv("DATABASE_URI")
	if dbURI == "" {
		dbURI = "postgresql://gophermartuser:gophermartpass@localhost:5432/gophermart"
	}

	db, err := pgxpool.Connect(ctx, dbURI)
	s.db = db
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (s *UserRegisterSuite) SetupTest() {
	login, password := "login123", "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	ctx := context.Background()
	_, _ = s.db.Exec(ctx, "TRUNCATE \"user\" CASCADE")
	_, _ = s.db.Exec(ctx, "TRUNCATE \"order\" CASCADE")
	_, _ = s.db.Exec(context.Background(), "INSERT INTO \"user\" (login, password) VALUES ($1, $2)", login, hashedPassword)
	response, _ := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(fmt.Sprintf(`{"login": "%s", "password":"%s"}`, login, password)).
		Post(loginURL)
	s.authCookie = response.Header().Get("Set-Cookie")
}

func (s *UserRegisterSuite) TestStoreNewOrderWithSuccess() {
	response, _ := s.client.R().
		SetHeader("Content-Type", "text/plain").
		SetCookie(&http.Cookie{Name: "Cookie", Value: authCookie}).
		SetBody("12345678903").
		Post(ordersURL)
	s.Equal(http.StatusAccepted, response.StatusCode())
}

func (s *UserRegisterSuite) TestStoreNewOrderWithBadOrderNum() {
	response, _ := s.client.R().
		SetHeader("Content-Type", "text/plain").
		SetCookie(&http.Cookie{Name: "Cookie", Value: authCookie}).
		SetBody("123456789").
		Post(ordersURL)
	s.Equal(http.StatusUnprocessableEntity, response.StatusCode())
}

func (s *UserRegisterSuite) TestGetOrdersList() {
	orderNum := "12345678903"
	_, _ = s.client.R().
		SetHeader("Content-Type", "text/plain").
		SetCookie(&http.Cookie{Name: "Cookie", Value: authCookie}).
		SetBody(orderNum).
		Post(ordersURL)

	var orders []Order

	response, _ := s.client.R().
		SetCookie(&http.Cookie{Name: "Cookie", Value: authCookie}).
		SetResult(&orders).
		Get(ordersURL)
	s.Equal(http.StatusOK, response.StatusCode())
	s.Assert().True(strings.Contains(response.Header().Get("Content-Type"), "application/json"))
	s.Equal(1, len(orders))
	s.Equal(orderNum, orders[0].Number)
	s.Assert().Contains([]string{"NEW", "PROCESSING"}, orders[0].Status)
}

func (s *UserRegisterSuite) TestGetOrdersEmptyList() {
	_, _ = s.db.Exec(context.Background(), "TRUNCATE \"order\" CASCADE")
	response, _ := s.client.R().
		SetCookie(&http.Cookie{Name: "Cookie", Value: authCookie}).
		Get(ordersURL)
	s.Equal(http.StatusNoContent, response.StatusCode())
}
