package member

import (
	"bytes"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func init() {
	log.Default()
}

func TestVisitorSendMessage(t *testing.T) {
	m := newMockNoMessageMember(false, 99)
	_, err := m.GetMessageSession("123")

	assert.Equal(t, err, errors.ErrLogin)
}

func TestGuestSendMessage(t *testing.T) {
	m := newMockNoMessageMember(true, models.Guest)
	_, err := m.GetMessageSession("123")

	assert.Equal(t, err, errors.ErrLogin)
}

func TestMemberSendMessage(t *testing.T) {
	m := newMockMember(true, true, false, models.Player)
	_, err := m.GetMessageSession("123")

	assert.Nil(t, err)
}

func TestMarketSendMessage(t *testing.T) {
	m := newMockMember(true, true, false, models.Market)
	_, err := m.GetMessageSession("123")

	assert.Nil(t, err)
}

func TestMemberIsBannedSendMessage(t *testing.T) {
	m := newMockMember(true, true, true, models.Player)
	_, err := m.GetMessageSession("123")

	assert.Equal(t, err, errors.ErrMemberBanned)
}

func TestMarketIsBannedSendMessage(t *testing.T) {
	m := newMockMember(true, true, true, models.Market)
	_, err := m.GetMessageSession("123")

	assert.Equal(t, err, errors.ErrMemberBanned)
}

func TestMemberBlockadeSendMessage(t *testing.T) {
	m := newMockBlockadeMember(models.Player)
	_, err := m.GetMessageSession("123")

	assert.Equal(t, err, errors.ErrBlockade)
}

func TestMarketBannedSendMessage(t *testing.T) {
	m := newMockBlockadeMember(models.Market)
	_, err := m.GetMessageSession("123")

	assert.Equal(t, err, errors.ErrBlockade)
}

func newMockNoMessageMember(isLogin bool, t int) Member {
	m, _ := newMockMemberCache(isLogin, false, false, t)
	return m
}

func newMockBlockadeMember(t int) Member {
	m, _ := newMockMemberCache(true, true, true, t)
	return m
}

func newMockMember(isLogin, isMessage, isBanned bool, t int) Member {
	m, c := newMockMemberCache(isLogin, isMessage, false, t)
	c.On("isBanned", mock.Anything).Return(isBanned, nil)
	return m
}

func newMockMemberCache(isLogin, isMessage, isBlockade bool, t int) (Member, *MockCache) {
	var err error
	cache := &MockCache{}
	member := Member{
		c: cache,
	}

	if !isLogin {
		err = errors.ErrLogin
	}

	cache.On("get", mock.Anything).Return(&models.Member{
		Type:       t,
		IsMessage:  isMessage,
		IsBlockade: isBlockade,
	}, err)
	return member, cache
}

func TestVisitorGiveRedEnvelope(t *testing.T) {
	m := newMockMember(false, false, false, 99)
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Equal(t, err, errors.ErrLogin)
}

func TestGuestGiveRedEnvelope(t *testing.T) {
	m := newMockMember(true, false, false, models.Guest)
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Equal(t, err, errors.ErrLogin)
}

func TestMemberGiveRedEnvelope(t *testing.T) {
	m := newMockMember(true, true, false, models.Player)
	m.cli = mockRedEnvelopeClient()

	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Nil(t, err)
}

func TestMarketGiveRedEnvelope(t *testing.T) {
	m := newMockMember(true, true, false, models.Market)
	m.cli = mockRedEnvelopeClient()

	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Nil(t, err)
}

func TestMemberBannedGiveRedEnvelope(t *testing.T) {
	m := newMockMember(true, true, true, models.Player)
	m.cli = mockRedEnvelopeClient()
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Nil(t, err)
}

func TestMarketBannedGiveRedEnvelope(t *testing.T) {
	m := newMockMember(true, true, true, models.Market)
	m.cli = mockRedEnvelopeClient()
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Nil(t, err)
}

func TestMemberBlockadeGiveRedEnvelope(t *testing.T) {
	m := newMockBlockadeMember(models.Player)
	m.cli = mockRedEnvelopeClient()
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Equal(t, err, errors.ErrBlockade)
}

func TestMarketBlockadeGiveRedEnvelope(t *testing.T) {
	m := newMockBlockadeMember(models.Market)
	m.cli = mockRedEnvelopeClient()
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Equal(t, err, errors.ErrBlockade)
}

func mockRedEnvelopeClient() *client.Client {
	return client.NewMockClient(func(req *http.Request) (resp *http.Response, err error) {
		body, err := json.Marshal(client.RedEnvelopeReply{})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
		}, nil
	})
}

func TestGetUserName(t *testing.T) {
	m, db, cache := mockMember()
	uids := []string{"1", "2", "3"}

	cache.On("getName", uids).
		Return(map[string]string{
			"1": "1",
			"3": "3",
		}, nil)

	db.m.On("GetMembersByUid", []string{"2"}).Return([]models.Member{
		models.Member{Uid: "2", Name: "2"},
	}, nil)

	cache.On("setName", map[string]string{"2": "2"}).Return(nil)

	name, err := m.GetUserNames(uids)

	assert.Nil(t, err)
	assert.Equal(t, map[string]string{
		"1": "1",
		"3": "3",
		"2": "2",
	}, name)
}

func TestGetUserNameNoRows(t *testing.T) {
	m, db, cache := mockMember()
	cache.On("getName", mock.Anything).Return(map[string]string{}, redis.Nil)
	db.m.On("GetMembersByUid", []string{}).Once().Return([]models.Member{}, nil)
	cache.On("setName", mock.Anything).Return(nil)

	_, err := m.GetUserNames([]string{})

	db.m.AssertExpectations(t)

	assert.Nil(t, err)
}

func TestGetUserNameError(t *testing.T) {
	m, _, cache := mockMember()
	cache.On("getName", mock.Anything).Return(map[string]string{}, errors.New(""))
	_, err := m.GetUserNames([]string{})

	assert.Equal(t, errors.New(""), err)
}

func mockMember() (*Member, *mockDb, *MockCache) {
	db := new(mockDb)
	cache := new(MockCache)
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
