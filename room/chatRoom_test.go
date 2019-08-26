package room

import (
	"database/sql"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gopkg.in/go-playground/validator.v8"
	"testing"
)

var jsonb = []byte(`{"token":"123","room_id":1}`)

func TestConnect(t *testing.T) {
	chat, _, cache, member := makeMock()

	cache.m.On("get", 1).
		Return(models.Room{Status: true}, nil)

	member.m.On("Login", 1, "123", "test").
		Return(&models.Member{}, "key", nil)

	user, key, rid, err := chat.Connect("test", jsonb)

	assert.Equal(t, &models.Member{}, user)
	assert.Equal(t, "key", key)
	assert.Equal(t, 1, rid)
	assert.Nil(t, err)
}

func TestConnectNotRoomCache(t *testing.T) {
	chat, db, cache, member := makeMock()

	cache.m.On("get", mock.Anything).
		Return(models.Room{}, redis.Nil)

	db.m.On("GetRoom", 1).
		Return(models.Room{Status: true}, nil)

	cache.m.On("set", models.Room{Status: true}).
		Return(nil)

	member.m.On("Login", mock.Anything, mock.Anything, mock.Anything).
		Return(&models.Member{}, "key", nil)

	_, _, _, err := chat.Connect("", jsonb)

	assert.Nil(t, err)
}

func TestConnectNotRoom(t *testing.T) {
	chat, db, cache, _ := makeMock()

	cache.m.On("get", mock.Anything).
		Return(models.Room{}, redis.Nil)

	db.m.On("GetRoom", 1).
		Return(models.Room{}, sql.ErrNoRows)

	_, _, _, err := chat.Connect("", jsonb)

	assert.Equal(t, errNoRoom, err)
}

func TestConnectRoomClose(t *testing.T) {
	chat, _, cache, _ := makeMock()

	cache.m.On("get", mock.Anything).
		Return(models.Room{Status: false}, nil)

	_, _, _, err := chat.Connect("", jsonb)

	assert.Equal(t, errRoomClose, err)
}

func TestConnectNotData(t *testing.T) {
	chat, _, _, _ := makeMock()

	_, _, _, err := chat.Connect("", []byte(""))
	e, ok := err.(*json.SyntaxError)

	assert.True(t, ok)
	assert.Equal(t, "unexpected end of JSON input", e.Error())
}

func TestConnectJson(t *testing.T) {
	chat, _, _, _ := makeMock()

	_, _, _, err := chat.Connect("", []byte(`{"token":"123","room_id":0}`))
	e, ok := err.(validator.ValidationErrors)

	assert.True(t, ok)
	assert.Equal(t, "Key: '.RoomID' Error:Field validation for 'RoomID' failed on the 'required' tag", e.Error())

	_, _, _, err = chat.Connect("", []byte(`{"token":"","room_id":1}`))
	e, ok = err.(validator.ValidationErrors)

	assert.True(t, ok)
	assert.Equal(t, "Key: '.Token' Error:Field validation for 'Token' failed on the 'required' tag", e.Error())
}

func makeMock() (*chat, *mockDb, *mockCache, *mockMember) {
	db := new(mockDb)
	member := new(mockMember)
	cache := new(mockCache)
	return &chat{
		db:     db,
		member: member,
		cache:  cache,
	}, db, cache, member
}

type mockDb struct {
	m mock.Mock
}

func (m *mockDb) GetRoom(id int) (models.Room, error) {
	arg := m.m.Called(id)
	return arg.Get(0).(models.Room), arg.Error(1)
}

type mockMember struct {
	m mock.Mock
}

func (m *mockMember) Login(rid int, token, server string) (*models.Member, string, error) {
	arg := m.m.Called(rid, token, server)
	return arg.Get(0).(*models.Member), arg.String(1), arg.Error(2)
}

func (m *mockMember) Logout(uid, key string) (bool, error) {
	arg := m.m.Called(uid, key)
	return arg.Bool(0), arg.Error(1)
}

func (m *mockMember) Heartbeat(uid string) error {
	arg := m.m.Called(uid)
	return arg.Error(0)
}

type mockCache struct {
	m mock.Mock
}

func (m *mockCache) set(room models.Room) error {
	arg := m.m.Called(room)
	return arg.Error(0)
}

func (m *mockCache) get(id int) (models.Room, error) {
	arg := m.m.Called(id)
	return arg.Get(0).(models.Room), arg.Error(1)
}

func (m *mockCache) addOnline(server string, online *Online) error {
	arg := m.m.Called(server, online)
	return arg.Error(0)
}

func (m *mockCache) getOnline(server string) (*Online, error) {
	arg := m.m.Called(server)
	return arg.Get(0).(*Online), arg.Error(1)
}
