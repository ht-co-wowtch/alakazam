package member

// 項目中的單元測試中對於不同會員種類有不同權限限制關係請參考 https://gitlab.com/jetfueltw/cpw/alakazam#permission

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"io/ioutil"
	"net/http"
	"testing"
)

func init() {
	log.Default()
}

// 會員操作權限

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
	m := newMockMessagesMember(true, true, false, models.Player)
	_, err := m.GetMessageSession("123")

	assert.Nil(t, err)
}

func TestMarketSendMessage(t *testing.T) {
	m := newMockMessagesMember(true, true, false, models.Market)
	_, err := m.GetMessageSession("123")

	assert.Nil(t, err)
}

func TestMemberIsBannedSendMessage(t *testing.T) {
	m := newMockMessagesMember(true, true, true, models.Player)
	_, err := m.GetMessageSession("123")

	assert.Equal(t, err, errors.ErrMemberBanned)
}

func TestMarketIsBannedSendMessage(t *testing.T) {
	m := newMockMessagesMember(true, true, true, models.Market)
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

func TestVisitorGiveRedEnvelope(t *testing.T) {
	m := newMockMessagesMember(false, false, false, 99)
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Equal(t, err, errors.ErrLogin)
}

func TestGuestGiveRedEnvelope(t *testing.T) {
	m := newMockMessagesMember(true, false, false, models.Guest)
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Equal(t, err, errors.ErrLogin)
}

func TestMemberGiveRedEnvelope(t *testing.T) {
	m := newMockRedEnvelopeMemberStatus(true, true, false, models.Player)
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Nil(t, err)
}

func TestMarketGiveRedEnvelope(t *testing.T) {
	m := newMockRedEnvelopeMemberStatus(true, true, false, models.Market)
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Nil(t, err)
}

func TestMemberBannedGiveRedEnvelope(t *testing.T) {
	m := newMockRedEnvelopeMemberStatus(true, true, true, models.Player)
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Nil(t, err)
}

func TestMarketBannedGiveRedEnvelope(t *testing.T) {
	m := newMockRedEnvelopeMemberStatus(true, true, true, models.Market)
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Nil(t, err)
}

func TestMemberBlockadeGiveRedEnvelope(t *testing.T) {
	m := newMockRedEnvelopeBlockadeMember(models.Player)
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Equal(t, err, errors.ErrBlockade)
}

func TestMarketBlockadeGiveRedEnvelope(t *testing.T) {
	m := newMockRedEnvelopeBlockadeMember(models.Market)
	_, _, err := m.GiveRedEnvelope("", "", RedEnvelope{})

	assert.Equal(t, err, errors.ErrBlockade)
}

func TestVisitorTaskRedEnvelope(t *testing.T) {
	m := newMockRedEnvelopeMemberStatus(false, false, false, 99)
	_, err := m.TakeRedEnvelope("", "", "")

	assert.Equal(t, err, errors.ErrLogin)
}

func TestGuestTaskRedEnvelope(t *testing.T) {
	m := newMockRedEnvelopeMemberStatus(true, false, false, models.Guest)
	_, err := m.TakeRedEnvelope("", "", "")

	assert.Equal(t, err, errors.ErrLogin)
}

func TestMemberTaskRedEnvelope(t *testing.T) {
	m := newMockRedEnvelopeMemberStatus(true, true, false, models.Player)
	_, err := m.TakeRedEnvelope("", "", "")

	assert.Nil(t, err)
}

func TestMarketTaskRedEnvelope(t *testing.T) {
	m := newMockRedEnvelopeMemberStatus(true, true, false, models.Market)
	_, err := m.TakeRedEnvelope("", "", "")

	assert.Nil(t, err)
}

// 紅包

func TestGetRedEnvelopeDetailByMemberName(t *testing.T) {
	m := newMemberMockFunc(func(cache *MockCache, db *models.MockDB) {
		cache.On("getName", []string{"B", "C", "A"}).Return(map[string]string{
			"A": "test1",
			"B": "test2",
			"C": "test3",
		}, nil)

	}, func(req *http.Request) (resp *http.Response, err error) {
		body, err := json.Marshal(client.RedEnvelopeDetail{
			RedEnvelopeInfo: client.RedEnvelopeInfo{Uid: "A"},
			Members: []client.MemberDetail{
				client.MemberDetail{
					Uid: "B",
				},
				client.MemberDetail{
					Uid: "C",
				},
			},
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
		}, nil
	})

	detail, err := m.GetRedEnvelopeDetail("aa641b03d4d548d233a73a219781gy61", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NzI0Mjk4ODYsImlkIjoiNTE2OTQ3N2I3OGRhNDg1ZDlhZjU3YjE1NzNkYmY2NWYiLCJ1aWQiOiIwZDY0MWIwM2Q0ZDU0OGRiYjNhNzNhMjE5NzgxMTI2MSJ9.PBaDJp6e8lc7r75VdUV2sPgka7IR3ScfbvhMlvAiJvY")

	assert.Nil(t, err)
	assert.Equal(t, "test1", detail.Name)
	assert.Equal(t, "test2", detail.Members[0].Name)
	assert.Equal(t, "test3", detail.Members[1].Name)

	m.mCache.AssertExpectations(t)
	m.mDb.AssertExpectations(t)
}

func TestGetRedEnvelopeDetailByAdminName(t *testing.T) {
	m := newMemberMockFunc(func(cache *MockCache, db *models.MockDB) {
		cache.On("getName", []string{"B", "C"}).Return(map[string]string{
			"B": "test2",
			"C": "test3",
		}, nil)

	}, func(req *http.Request) (resp *http.Response, err error) {
		body, err := json.Marshal(client.RedEnvelopeDetail{
			RedEnvelopeInfo: client.RedEnvelopeInfo{Uid: "", IsAdmin: true},
			Members: []client.MemberDetail{
				client.MemberDetail{
					Uid: "B",
				},
				client.MemberDetail{
					Uid: "C",
				},
			},
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
		}, nil
	})

	detail, err := m.GetRedEnvelopeDetail("aa641b03d4d548d233a73a219781gy61", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NzI0Mjk4ODYsImlkIjoiNTE2OTQ3N2I3OGRhNDg1ZDlhZjU3YjE1NzNkYmY2NWYiLCJ1aWQiOiIwZDY0MWIwM2Q0ZDU0OGRiYjNhNzNhMjE5NzgxMTI2MSJ9.PBaDJp6e8lc7r75VdUV2sPgka7IR3ScfbvhMlvAiJvY")

	assert.Nil(t, err)
	assert.Equal(t, RootName, detail.Name)
	assert.Equal(t, "test2", detail.Members[0].Name)
	assert.Equal(t, "test3", detail.Members[1].Name)

	m.mCache.AssertExpectations(t)
	m.mDb.AssertExpectations(t)
}

// =====================================================================================================================

func newMockNoMessageMember(isLogin bool, t int) mockMember {
	return newMockMemberStatus(isLogin, false, false, t, nil)
}

func newMockBlockadeMember(t int) mockMember {
	return newMockMemberStatus(true, true, true, t, nil)
}

func newMockMessagesMember(isLogin, isMessage, isBanned bool, t int) mockMember {
	m := newMockMemberStatus(isLogin, isMessage, false, t, nil)
	m.mockCache(func(cache *MockCache) {
		cache.On("isBanned", mock.Anything).Return(isBanned, nil)
	})
	return m
}

var mockRedEnvelopeHttpFunc = func(req *http.Request) (resp *http.Response, err error) {
	body, err := json.Marshal(client.RedEnvelopeReply{})
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
	}, nil
}

func newMockRedEnvelopeMemberStatus(isLogin, isMessage, isBanned bool, t int) mockMember {
	m := newMockMessagesMember(isLogin, isMessage, isBanned, t)
	m.mockHttpFunc(mockRedEnvelopeHttpFunc)
	return m
}

func newMockRedEnvelopeBlockadeMember(t int) mockMember {
	m := newMockBlockadeMember(t)
	m.mockHttpFunc(mockRedEnvelopeHttpFunc)
	return m
}

func newMockMemberStatus(isLogin, isMessage, isBlockade bool, t int, httpFunc client.TransportFunc) mockMember {
	return newMemberMockFunc(func(cache *MockCache, db *models.MockDB) {
		var err error
		if !isLogin {
			err = errors.ErrLogin
		}

		cache.On("get", mock.Anything).Return(&models.Member{
			Type:       t,
			IsMessage:  isMessage,
			IsBlockade: isBlockade,
		}, err)

	}, httpFunc)
}

func newMemberMockFunc(m func(cache *MockCache, db *models.MockDB), httpFunc client.TransportFunc) mockMember {
	cache := &MockCache{}
	db := &models.MockDB{}
	member := newMockMember(cache, db, httpFunc)
	m(cache, db)
	return member
}

type mockMember struct {
	Member
	mCache *MockCache
	mDb    *models.MockDB
}

func newMockMember(cache *MockCache, db *models.MockDB, httpFunc client.TransportFunc) mockMember {
	return mockMember{
		Member: Member{
			c:   cache,
			db:  db,
			cli: client.NewMockClient(httpFunc),
		},
		mCache: cache,
		mDb:    db,
	}
}

func (m *mockMember) mockHttpFunc(httpFunc client.TransportFunc) {
	m.cli = client.NewMockClient(httpFunc)
}

func (m *mockMember) mockCache(mock func(cache *MockCache)) {
	mock(m.c.(*MockCache))
}
