package member

import (
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"testing"
	"time"
)

func TestGetUserName(t *testing.T) {
	m, db, cache := mockMember()
	uids := []string{"1", "2", "3"}

	cache.m.On("getName", uids).
		Return(map[string]string{
			"1": "1",
			"3": "3",
		}, nil)

	db.m.On("GetMembersByUid", []string{"2"}).Return([]models.Member{
		models.Member{Uid: "2", Name: "2"},
	}, nil)

	cache.m.On("setName", map[string]string{"2": "2"}).Return(nil)

	name, err := m.GetUserName(uids)

	assert.Nil(t, err)
	assert.Equal(t, map[string]string{
		"1": "1",
		"3": "3",
		"2": "2",
	}, name)
}

func TestGetUserNameNoRows(t *testing.T) {
	m, db, cache := mockMember()
	cache.m.On("getName", mock.Anything).Return(map[string]string{}, redis.Nil)
	db.m.On("GetMembersByUid", []string{}).Once().Return([]models.Member{}, nil)
	cache.m.On("setName", mock.Anything).Return(nil)

	_, err := m.GetUserName([]string{})

	db.m.AssertExpectations(t)

	assert.Nil(t, err)
}

func TestGetUserNameError(t *testing.T) {
	m, _, cache := mockMember()
	cache.m.On("getName", mock.Anything).Return(map[string]string{}, errors.New(""))
	_, err := m.GetUserName([]string{})

	assert.Equal(t, errors.New(""), err)
}

func mockMember() (*Member, *mockDb, *mockCache) {
	db := new(mockDb)
	cache := new(mockCache)
	return &Member{
		db: db,
		c:  cache,
	}, db, cache
}

type mockDb struct {
	m mock.Mock
}

func (m *mockDb) Find(uid string) (*models.Member, error) {
	arg := m.m.Called(uid)
	return arg.Get(0).(*models.Member), arg.Error(1)
}

func (m *mockDb) CreateUser(member *models.Member) (bool, error) {
	arg := m.m.Called(member)
	return arg.Bool(0), arg.Error(1)
}

func (m *mockDb) UpdateUser(member *models.Member) (bool, error) {
	arg := m.m.Called(member)
	return arg.Bool(0), arg.Error(1)
}

func (m *mockDb) GetMembers(ids []int) ([]models.Member, error) {
	arg := m.m.Called(ids)
	return arg.Get(0).([]models.Member), arg.Error(1)
}

func (m *mockDb) GetMembersByUid(uid []string) ([]models.Member, error) {
	arg := m.m.Called(uid)
	return arg.Get(0).([]models.Member), arg.Error(1)
}

func (m *mockDb) SetBlockade(uid string) (int64, error) {
	arg := m.m.Called(uid)
	return arg.Get(0).(int64), arg.Error(1)
}

func (m *mockDb) DeleteBanned(uid string) (int64, error) {
	arg := m.m.Called(uid)
	return arg.Get(0).(int64), arg.Error(1)
}

func (m *mockDb) SetBannedLog(memberId int, sec time.Duration, isSystem bool) (bool, error) {
	arg := m.m.Called(memberId, sec, isSystem)
	return arg.Bool(0), arg.Error(1)
}

func (m *mockDb) GetTodaySystemBannedLog(memberId int) ([]models.BannedLog, error) {
	arg := m.m.Called(memberId)
	return arg.Get(0).([]models.BannedLog), arg.Error(1)
}

func (m *mockDb) UpdateIsMessage(memberId int, isMessage bool) (bool, error) {
	arg := m.m.Called(memberId, isMessage)
	return arg.Bool(0), arg.Error(1)
}
