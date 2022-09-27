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
	"golang.org/x/crypto/bcrypt"
)

const baseUserURL = "http://localhost:8080/api/user"
const registerURL = baseUserURL + "/register"
const loginURL = baseUserURL + "/login"

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

func (s *UserRegisterSuite) SetupSuite() {
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
}

func (s *UserRegisterSuite) SetupTest() {
	ctx := context.Background()
	_, _ = s.db.Exec(ctx, "TRUNCATE \"user\" CASCADE")
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
	token := response.Header().Get("Authorization")
	s.Assert().NotEmpty(token)
	setCookie := response.Header().Get("Set-Cookie")
	s.Assert().NotEmpty(setCookie)
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
	_, _ = s.db.Exec(context.Background(), "INSERT INTO \"user\" (login, password) VALUES ($1, $2)", dummyLogin, "123")

	response, _ := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(fmt.Sprintf(`{"login": "%s", "password":"password"}`, dummyLogin)).
		Post(registerURL)
	s.Equal(http.StatusConflict, response.StatusCode())
}

func (s *UserRegisterSuite) TestUserLoginSuccessfully() {
	dummyLogin := "dummyLogin"
	dummyPassword := "123"
	dummyHashedPassword, _ := bcrypt.GenerateFromPassword([]byte(dummyPassword), bcrypt.DefaultCost)
	_, _ = s.db.Exec(context.Background(), "INSERT INTO \"user\" (login, password) VALUES ($1, $2)", dummyLogin, string(dummyHashedPassword))

	response, _ := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(fmt.Sprintf(`{"login": "%s", "password":"%s"}`, dummyLogin, dummyPassword)).
		Post(loginURL)

	s.Equal(http.StatusOK, response.StatusCode())
	token := response.Header().Get("Authorization")
	s.Assert().NotEmpty(token)
	setCookie := response.Header().Get("Set-Cookie")
	s.Assert().NotEmpty(setCookie)
}

func (s *UserRegisterSuite) TestUserLoginBadRequest() {
	response, _ := s.client.R().
		SetHeader("Content-Type", "text/plain").
		SetBody(`some text`).
		Post(registerURL)
	s.Equal(http.StatusBadRequest, response.StatusCode())
}

func (s *UserRegisterSuite) TestUserLoginWrongPassword() {
	dummyLogin := "dummyLogin"
	dummyPassword := "123"
	_, _ = s.db.Exec(context.Background(), "INSERT INTO \"user\" (login, password) VALUES ($1, $2)", dummyLogin, dummyPassword)

	response, _ := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(fmt.Sprintf(`{"login": "%s", "password":"%s"}`, dummyLogin, "321")).
		Post(loginURL)

	s.Equal(http.StatusUnauthorized, response.StatusCode())
}
