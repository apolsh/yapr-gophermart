package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

const registerURL = "http://localhost:8080/api/user/register"

type UserRegisterSuite struct {
	suite.Suite
	db     *pgxpool.Pool
	client *resty.Client
}

func TestUserRegisterSuite(t *testing.T) {
	suite.Run(t, new(UserRegisterSuite))
}

var db *pgxpool.Pool
var client *resty.Client

func (s *UserRegisterSuite) SetupTest() {
	s.client = resty.New()
	ctx := context.Background()
	dbURI := os.Getenv("DATABASE_URI")
	if dbURI == "" {
		dbURI = "postgresql://gophermartuser:gophermartpass@localhost:5432/gophermart"
	}

	db, err := pgxpool.New(ctx, dbURI)
	s.db = db
	if err != nil {
		fmt.Println(err.Error())
	}
	_, _ = db.Exec(ctx, "TRUNCATE \"user\"")
}

func (s *UserRegisterSuite) TearDownSuite() {
	s.db.Close()
}

func (s *UserRegisterSuite) TestUserRegisterSuccess() {
	response, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"login":"login", "password":"password"}`).
		Post(registerURL)
	if err != nil {
		fmt.Println(err.Error())
	}
	s.Equal(http.StatusOK, response.StatusCode())
}

func (s *UserRegisterSuite) TestUserRegisterBadRequest() {
	response, _ := s.client.R().
		SetHeader("Content-Type", "text/plain").
		SetBody(`some text`).
		Post(registerURL)
	s.Equal(http.StatusBadRequest, response.StatusCode())
}

func (s *UserRegisterSuite) TestUserRegisterLoginIsAlreadyUsed() {
	dummyLogin := "dummyLogin"
	_, err := s.db.Exec(context.Background(), "INSERT INTO \"user\" (login, password) VALUES ($1, $2)", dummyLogin, "123")
	if err != nil {
		fmt.Println(err.Error())
	}

	response, _ := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(fmt.Sprintf(`{"login": "%s", "password":"password"}`, dummyLogin)).
		Post(registerURL)
	s.Equal(http.StatusConflict, response.StatusCode())
}
