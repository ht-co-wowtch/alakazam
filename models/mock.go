package models

import (
	"github.com/stretchr/testify/mock"
	"time"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Find(uid string) (*Member, error) {
	arg := m.Called(uid)
	return arg.Get(0).(*Member), arg.Error(1)
}

func (m *MockDB) CreateUser(member *Member) (bool, error) {
	arg := m.Called(member)
	return arg.Bool(0), arg.Error(1)
}

func (m *MockDB) UpdateUser(member *Member) (bool, error) {
	arg := m.Called(member)
	return arg.Bool(0), arg.Error(1)
}

func (m *MockDB) GetMembers(ids []int) ([]Member, error) {
	arg := m.Called(ids)
	return arg.Get(0).([]Member), arg.Error(1)
}

func (m *MockDB) GetMembersByUid(uid []string) ([]Member, error) {
	arg := m.Called(uid)
	return arg.Get(0).([]Member), arg.Error(1)
}

func (m *MockDB) SetBlockade(uid string) (int64, error) {
	arg := m.Called(uid)
	return arg.Get(0).(int64), arg.Error(1)
}

func (m *MockDB) DeleteBanned(uid string) (int64, error) {
	arg := m.Called(uid)
	return arg.Get(0).(int64), arg.Error(1)
}

func (m *MockDB) SetBannedLog(memberId int, sec time.Duration, isSystem bool) (bool, error) {
	arg := m.Called(memberId, sec, isSystem)
	return arg.Bool(0), arg.Error(1)
}

func (m *MockDB) GetTodaySystemBannedLog(memberId int) ([]BannedLog, error) {
	arg := m.Called(memberId)
	return arg.Get(0).([]BannedLog), arg.Error(1)
}

func (m *MockDB) UpdateIsMessage(memberId int, isMessage bool) (bool, error) {
	arg := m.Called(memberId, isMessage)
	return arg.Bool(0), arg.Error(1)
}
