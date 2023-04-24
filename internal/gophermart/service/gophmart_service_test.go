package service

import (
	"context"
	"testing"

	"github.com/apolsh/yapr-gophermart/internal/gophermart/dto"
	"github.com/apolsh/yapr-gophermart/internal/gophermart/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	token          = "token"
	login          = "login"
	password       = "password"
	hashedPassword = "$2a$10$zkIMBhdT7Lvw3RRWoJ1UFu6TOAamrWSn6ZA.U5mBS5Gjo7r1OV5Ku"
	userID         = "userID"
	user           = dto.User{ID: userID, Login: login, HashedPassword: hashedPassword}
	accrualSystem  = "http://dummyAccrualSystem.com"
)

type ServiceSuite struct {
	suite.Suite
	orderStorage *mocks.MockOrderStorage
	userStorage  *mocks.MockUserStorage
	ctrl         *gomock.Controller
	service      *GophermartServiceImpl
}

func TestRouterSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.ctrl = ctrl
	s.userStorage = mocks.NewMockUserStorage(ctrl)
	s.orderStorage = mocks.NewMockOrderStorage(ctrl)

	service, _ := NewGophermartServiceImpl(token, 0, accrualSystem, s.userStorage, s.orderStorage)
	s.service = service
}

func (s *ServiceSuite) TestAddUserWithSuccess() {
	s.userStorage.EXPECT().NewUser(gomock.Any(), login, gomock.Any()).Return(userID, nil)

	token, err := s.service.AddUser(context.Background(), login, password)
	assert.NoError(s.T(), err)
	assert.True(s.T(), len(token) > 0)
}

func (s *ServiceSuite) TestAddUserWihEmptyValues() {
	_, err := s.service.AddUser(context.Background(), "", password)
	assert.Error(s.T(), ErrorEmptyValue, err)
	_, err = s.service.AddUser(context.Background(), login, "")
	assert.Error(s.T(), ErrorEmptyValue, err)
}

func (s *ServiceSuite) TestLoginUserWithSuccess() {
	s.userStorage.EXPECT().Get(gomock.Any(), login).Return(user, nil)

	token, err := s.service.LoginUser(context.Background(), login, password)
	assert.NoError(s.T(), err)
	assert.True(s.T(), len(token) > 0)
}

func (s *ServiceSuite) TestLoginUserInvalidPassword() {
	s.userStorage.EXPECT().Get(gomock.Any(), login).Return(user, nil)

	_, err := s.service.LoginUser(context.Background(), login, "dummyPassword")
	assert.Error(s.T(), ErrorEmptyValue, err)
}

func (s *ServiceSuite) TestLoginUserWihEmptyValues() {
	_, err := s.service.LoginUser(context.Background(), "", password)
	assert.Error(s.T(), ErrorEmptyValue, err)
	_, err = s.service.LoginUser(context.Background(), login, "")
	assert.Error(s.T(), ErrorEmptyValue, err)
}
