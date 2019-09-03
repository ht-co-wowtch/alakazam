package room

import (
	"database/sql"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gopkg.in/go-playground/validator.v8"
	"testing"
)

var jsonb = []byte(`{"token":"123","room_id":1}`)

func TestConnect(t *testing.T) {
	chat, _, cache, member := makeMock()

	cache.m.On("getChat", 1).
		Return(models.Room{Status: true}, nil)

	member.m.On("Login", 1, "123", "test").
		Return(&models.Member{}, "", nil)

	_, err := chat.Connect("test", jsonb)

	assert.Nil(t, err)
}

func TestConnectRoom(t *testing.T) {
	chat, _, cache, member := makeMock()

	cache.m.On("getChat", mock.Anything).
		Return(models.Room{Id: 1, Status: true, IsMessage: true, HeaderMessage: []byte(`message`)}, nil)

	member.m.On("Login", mock.Anything, mock.Anything, mock.Anything).
		Return(&models.Member{
			Uid:        "1",
			Name:       "test",
			IsBlockade: true,
			IsMessage:  true,
			Type:       models.Player,
		}, "key", nil)

	connect, _ := chat.Connect("test", jsonb)

	assert.Equal(t, &pb.ConnectReply{
		Uid:           "1",
		Key:           "key",
		Name:          "test",
		RoomID:        1,
		Heartbeat:     10,
		IsBlockade:    true,
		IsMessage:     true,
		IsRedEnvelope: true,
		HeaderMessage: []byte(`message`),
	}, connect)
}

func TestConnectNotRoomCache(t *testing.T) {
	chat, db, cache, member := makeMock()

	cache.m.On("getChat", mock.Anything).
		Return(models.Room{}, redis.Nil)

	db.m.On("GetChat", 1).
		Return(models.Room{Status: true}, models.RoomTopMessage{}, nil)

	cache.m.On("set", models.Room{Status: true}).
		Return(nil)

	member.m.On("Login", mock.Anything, mock.Anything, mock.Anything).
		Return(&models.Member{}, "key", nil)

	_, err := chat.Connect("", jsonb)

	assert.Nil(t, err)
}

func TestConnectCacheRoomMessage(t *testing.T) {
	chat, db, cache, member := makeMock()

	cache.m.On("getChat", mock.Anything).
		Return(models.Room{}, redis.Nil)

	db.m.On("GetChat", 1).
		Return(models.Room{Status: true}, models.RoomTopMessage{MsgId: 1, RoomId: 1}, nil)

	cache.m.On("setChat", models.Room{Status: true}, `{"id":1,"uid":"root","type":"top","name":"管理员","avatar":"","message":"","time":"00:00:00"}`).
		Return(nil)

	member.m.On("Login", mock.Anything, mock.Anything, mock.Anything).
		Return(&models.Member{}, "key", nil)

	_, err := chat.Connect("", jsonb)

	assert.Nil(t, err)
}

func TestConnectNotRoom(t *testing.T) {
	chat, db, cache, _ := makeMock()

	cache.m.On("getChat", mock.Anything).
		Return(models.Room{}, redis.Nil)

	db.m.On("GetChat", 1).
		Return(models.Room{Status: true}, models.RoomTopMessage{}, sql.ErrNoRows)

	_, err := chat.Connect("", jsonb)

	assert.Equal(t, errors.ErrNoRoom, err)
}

func TestConnectRoomClose(t *testing.T) {
	chat, _, cache, _ := makeMock()

	cache.m.On("getChat", mock.Anything).
		Return(models.Room{Status: false}, nil)

	_, err := chat.Connect("", jsonb)

	assert.Equal(t, errors.ErrRoomClose, err)
}

func TestConnectNotData(t *testing.T) {
	chat, _, _, _ := makeMock()

	_, err := chat.Connect("", []byte(""))
	e, ok := err.(*json.SyntaxError)

	assert.True(t, ok)
	assert.Equal(t, "unexpected end of JSON input", e.Error())
}

func TestConnectJson(t *testing.T) {
	chat, _, _, _ := makeMock()

	_, err := chat.Connect("", []byte(`{"token":"123","room_id":0}`))
	e, ok := err.(validator.ValidationErrors)

	assert.True(t, ok)
	assert.Equal(t, "Key: '.RoomID' Error:Field validation for 'RoomID' failed on the 'required' tag", e.Error())

	_, err = chat.Connect("", []byte(`{"token":"","room_id":1}`))
	e, ok = err.(validator.ValidationErrors)

	assert.True(t, ok)
	assert.Equal(t, "Key: '.Token' Error:Field validation for 'Token' failed on the 'required' tag", e.Error())
}

func TestReloadChatMessage(t *testing.T) {
	c.c.FlushAll()

	chat, db, cache, _ := makeMock()

	db.m.On("GetChat", 1).
		Return(models.Room{}, models.RoomTopMessage{RoomId: 1}, nil)

	cache.m.On("setChat", mock.Anything, mock.Anything).
		Return(nil)

	room, err := chat.reloadChat(roomTest.Id)

	assert.Nil(t, err)
	assert.Equal(t, []byte(`{"id":0,"uid":"root","type":"top","name":"管理员","avatar":"","message":"","time":"00:00:00"}`), room.HeaderMessage)
}

func makeMock() (*chat, *mockDb, *mockCache, *mockMember) {
	db := new(mockDb)
	member := new(mockMember)
	cache := new(mockCache)
	return &chat{
		db:               db,
		member:           member,
		cache:            cache,
		heartbeatNanosec: 10,
	}, db, cache, member
}

type mockDb struct {
	m mock.Mock
}

func (m *mockDb) GetChat(id int) (models.Room, models.RoomTopMessage, error) {
	arg := m.m.Called(id)
	return arg.Get(0).(models.Room), arg.Get(1).(models.RoomTopMessage), arg.Error(2)
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

func (m *mockMember) GetSession(uid string) (*models.Member, error) {
	arg := m.m.Called(uid)
	return arg.Get(0).(*models.Member), arg.Error(1)
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

func (m *mockCache) setChat(room models.Room, message []byte) error {
	arg := m.m.Called(room, message)
	return arg.Error(0)
}

func (m *mockCache) getChat(id int) (models.Room, error) {
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

func (m *mockCache) setChatTopMessage(rids []int32, message []byte) error {
	arg := m.m.Called(rids, message)
	return arg.Error(0)
}

func (m *mockCache) getChatTopMessage(rid int) ([]byte, error) {
	arg := m.m.Called(rid)
	return arg.Get(0).([]byte), arg.Error(1)
}

func (m *mockCache) deleteChatTopMessage(rids []int32) error {
	arg := m.m.Called(rids)
	return arg.Error(0)
}
