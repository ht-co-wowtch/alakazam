package room

// 項目中的單元測試中對於不同會員種類有不同權限限制關係請參考 https://gitlab.com/jetfueltw/cpw/alakazam#permission

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"net/http"
	"testing"
)

func TestVisitorConnectionRoom(t *testing.T) {
	reply, err := connectChat(true, newVisitorMember())

	assert.Equal(t, err.(errdefs.Causer).Code, errors.NoLogin)
	assert.Nil(t, reply)
}

func TestGuestConnectionRoom(t *testing.T) {
	reply, err := connectChat(true, newGuestMember(true))

	assert.Nil(t, err)
	assert.True(t, reply.Connect.Status)
	assert.False(t, reply.Connect.Permission.IsMessage)
	assert.False(t, reply.Connect.Permission.IsRedEnvelope)
	assert.Equal(t, reply.Connect.PermissionMessage.IsRedEnvelope, "请先登入会员")
	assert.Equal(t, reply.Connect.PermissionMessage.IsRedEnvelope, "请先登入会员")
}

func TestMemberConnectionRoom(t *testing.T) {
	reply, err := connectChat(true, newPlayMember(true, true, false))

	assert.Nil(t, err)
	assert.True(t, reply.Connect.Status)
	assert.True(t, reply.Connect.Permission.IsMessage)
	assert.True(t, reply.Connect.Permission.IsRedEnvelope)
}

func TestMarketConnectionRoom(t *testing.T) {
	reply, err := connectChat(true, newMarketMember(true, false, true))

	assert.Nil(t, err)
	assert.True(t, reply.Connect.Status)
	assert.True(t, reply.Connect.Permission.IsMessage)
	assert.True(t, reply.Connect.Permission.IsRedEnvelope)
}

func TestVisitorConnectionCloseRoom(t *testing.T) {
	_, err := connectChat(false, newVisitorMember())

	assert.Equal(t, err, errors.ErrRoomClose)
}

func TestGuestConnectionCloseRoom(t *testing.T) {
	_, err := connectChat(false, newGuestMember(true))

	assert.Equal(t, err, errors.ErrRoomClose)
}

func TestMemberConnectionCloseRoom(t *testing.T) {
	_, err := connectChat(false, newPlayMember(true, true, false))

	assert.Equal(t, err, errors.ErrRoomClose)
}

func TestMarketConnectionCloseRoom(t *testing.T) {
	_, err := connectChat(false, newMarketMember(true, true, false))

	assert.Equal(t, err, errors.ErrRoomClose)
}

func connectChat(chatStatus bool, member member.Chat) (*pb.ConnectReply, error) {
	return connectNewChat(chatStatus, true, member)
}

func connectNewChat(chatStatus, chatIsMessage bool, member member.Chat) (*pb.ConnectReply, error) {
	c := newChat(chatStatus, chatIsMessage, member)
	return c.Connect("", []byte(`{"token":"test", "room_id":1}`))
}

func newChat(status, isMessage bool, member member.Chat) chat {
	cache := &mockCache{}
	room := models.Room{Status: status, IsMessage: isMessage}
	cache.On("getChat", 1).Return(room, nil)
	return chat{
		cache:  cache,
		member: member,
	}
}

func newVisitorMember() *member.MockMember {
	return newMember(false, false, false, 0)
}

func newGuestMember(isLogin bool) *member.MockMember {
	return newMember(isLogin, false, false, models.Guest)
}

func newPlayMember(isLogin, isMessage, isBlockade bool) *member.MockMember {
	return newMember(isLogin, isBlockade, isMessage, models.Player)
}

func newMarketMember(isLogin, isBlockade, isMessage bool) *member.MockMember {
	return newMember(isLogin, isBlockade, isMessage, models.Market)
}

func newMember(isLogin, isBlockade, isMessage bool, t int) *member.MockMember {
	var err error
	m := &member.MockMember{}
	user := models.Member{
		Type:       t,
		IsMessage:  isMessage,
		IsBlockade: isBlockade,
	}

	if !isLogin {
		err = errdefs.Causer{
			Status: http.StatusNotFound, Code: errors.NoLogin,
		}
	}

	m.On("Login", mock.Anything, mock.Anything, mock.Anything).Return(&user, "", err)
	return m
}
