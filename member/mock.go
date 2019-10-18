package member

import (
	"github.com/stretchr/testify/mock"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
)

type MockMember struct {
	mock.Mock
}

func (m *MockMember) Get(uid string) (*models.Member, error) {
	arg := m.Called(uid)
	return arg.Get(0).(*models.Member), arg.Error(1)
}

func (m *MockMember) GetSession(uid string) (*models.Member, error) {
	arg := m.Called(uid)
	return arg.Get(0).(*models.Member), arg.Error(1)
}

func (m *MockMember) Login(rid int, token, server string) (*models.Member, string, error) {
	arg := m.Called(rid, token, server)
	return arg.Get(0).(*models.Member), arg.String(1), arg.Error(2)
}

func (m *MockMember) Logout(uid, key string) (bool, error) {
	arg := m.Called(uid, key)
	return arg.Bool(0), arg.Error(1)
}

func (m *MockMember) Heartbeat(uid string) error {
	arg := m.Called(uid)
	return arg.Error(0)
}

