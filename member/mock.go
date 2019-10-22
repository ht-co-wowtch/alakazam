package member

import (
	"github.com/stretchr/testify/mock"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"time"
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

func (m *MockMember) GetMessageSession(uid string) (*models.Member, error) {
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

type MockCache struct {
	mock.Mock
}

func (m *MockCache) login(member *models.Member, key, server string) error {
	arg := m.Called(member, key, server)
	return arg.Error(0)
}

func (m *MockCache) set(member *models.Member) (bool, error) {
	arg := m.Called(member)
	return arg.Bool(0), arg.Error(1)
}

func (m *MockCache) get(uid string) (*models.Member, error) {
	arg := m.Called(uid)
	return arg.Get(0).(*models.Member), arg.Error(1)
}

func (m *MockCache) getKey(uid string) ([]string, error) {
	arg := m.Called(uid)
	return arg.Get(0).([]string), arg.Error(1)
}

func (m *MockCache) logout(uid, key string) (bool, error) {
	arg := m.Called(uid, key)
	return arg.Bool(0), arg.Error(1)
}

func (m *MockCache) delete(uid string) (bool, error) {
	arg := m.Called(uid)
	return arg.Bool(0), arg.Error(1)
}

func (m *MockCache) refreshExpire(uid string) error {
	arg := m.Called(uid)
	return arg.Error(0)
}

func (m *MockCache) setName(name map[string]string) error {
	arg := m.Called(name)
	return arg.Error(0)
}

func (m *MockCache) getName(uid []string) (map[string]string, error) {
	arg := m.Called(uid)
	return arg.Get(0).(map[string]string), arg.Error(1)
}

func (m *MockCache) setBanned(uid string, expired time.Duration) (bool, error) {
	arg := m.Called(uid)
	return arg.Bool(0), arg.Error(1)
}

func (m *MockCache) isBanned(uid string) (bool, error) {
	arg := m.Called(uid)
	return arg.Bool(0), arg.Error(1)
}

func (m *MockCache) delBanned(uid string) (bool, error) {
	arg := m.Called(uid)
	return arg.Bool(0), arg.Error(1)
}
