package member

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"testing"
)

const (
	_uid = "0d641b03d4d548dbb3a73a2197811261"
)

func TestRemoveBannedNotMember(t *testing.T) {
	m := newMockRemoveBannedMember(false, false, false)
	err := m.RemoveBanned(_uid)

	assert.Equal(t, errors.ErrNoMember, err)

	m.assertExpectations(t)
}

func TestRemoveBanned(t *testing.T) {
	m := newMockRemoveBannedMember(true, true, false)
	err := m.RemoveBanned(_uid)

	assert.Nil(t, err)

	m.assertExpectations(t)
}

func TestRemoveBannedIsNotBanned(t *testing.T) {
	m := newMockRemoveBannedMember(true, false, true)
	err := m.RemoveBanned(_uid)

	assert.Nil(t, err)

	m.assertExpectations(t)
}

func TestRemoveBannedIsNotBannedCache(t *testing.T) {
	m := newMockRemoveBannedMember(true, false, false)
	err := m.RemoveBanned(_uid)

	assert.Nil(t, err)

	m.assertExpectations(t)
}

type mockRemoveBannedMember struct {
	mockMember
}

func newMockRemoveBannedMember(isMemberExist, isBannedCache, memberIsMessage bool) mockRemoveBannedMember {
	return mockRemoveBannedMember{
		newMemberMockFunc(func(cache *MockCache, db *models.MockDB) {
			find := db.On("Find", _uid)
			if isMemberExist {
				find.Return(&models.Member{IsMessage: memberIsMessage}, nil)

				cache.On("delBanned", _uid).Return(isBannedCache, nil)

				if memberIsMessage == false {
					db.On("UpdateIsMessage", 0, true).Return(true, nil)
					cache.On("set", &models.Member{IsMessage: true}).Return(true, nil)
				}
			} else {
				find.Return(&models.Member{}, sql.ErrNoRows)
			}
		}, nil),
	}
}

func (m *mockRemoveBannedMember) assertExpectations(t *testing.T) {
	m.mCache.AssertExpectations(t)
	m.mDb.AssertExpectations(t)
}
