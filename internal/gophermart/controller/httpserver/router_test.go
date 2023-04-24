package httpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apolsh/yapr-gophermart/internal/gophermart/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type userCreds struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

var (
	login    = "login"
	password = "password"
	token    = "token"
)

type RouterSuite struct {
	suite.Suite
	service *mocks.MockGophermartService
	handler http.Handler
	ctrl    *gomock.Controller
}

func TestRouterSuite(t *testing.T) {
	suite.Run(t, new(RouterSuite))
}

func (s *RouterSuite) SetupTest() {
	r := chi.NewRouter()
	ctrl := gomock.NewController(s.T())
	s.ctrl = ctrl
	s.service = mocks.NewMockGophermartService(ctrl)

	RegisterRoutes(r, s.service)

	s.handler = r
}

func (s *RouterSuite) TestRegisterUserSuccess() {
	s.service.EXPECT().AddUser(gomock.Any(), login, password).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", credsBody(login, password))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()

	s.handler.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusOK, resp.Code)
}

func (s *RouterSuite) TestRegisterUserWrongMimeType() {
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", credsBody(login, password))
	req.Header.Set("Content-Type", "text/plain")

	resp := httptest.NewRecorder()

	s.handler.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusBadRequest, resp.Code)
}

func (s *RouterSuite) TestLoginUserSuccess() {
	s.service.EXPECT().LoginUser(gomock.Any(), login, password).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", credsBody(login, password))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()

	s.handler.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusOK, resp.Code)
}

func (s *RouterSuite) TestLoginUserWrongMimeType() {
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", credsBody(login, password))
	req.Header.Set("Content-Type", "text/plain")

	resp := httptest.NewRecorder()

	s.handler.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusBadRequest, resp.Code)
}

func credsBody(login, password string) io.Reader {
	creds, _ := json.Marshal(userCreds{login, password})
	return bytes.NewBuffer(creds)
}
