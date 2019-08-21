package models

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"testing"
	"time"
)

func TestMemberTableName(t *testing.T) {
	m := new(Member)

	assert.Equal(t, "members", m.TableName())
}

func TestCreateUser(t *testing.T) {
	uid := id.UUid32()
	member := &Member{
		Uid:        uid,
		Name:       "test",
		Avatar:     "/",
		IsBlockade: true,
	}

	aff, err := s.CreateUser(member)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)

	a := new(Member)
	ok, err := x.Where("uid = ?", uid).Get(a)

	assert.Nil(t, err)
	assert.True(t, ok)
	assert.False(t, a.CreateAt.IsZero())

	member.CreateAt = time.Time{}
	a.CreateAt = time.Time{}

	assert.Equal(t, member, a)
}

func TestUpdateUser(t *testing.T) {
	assert.NoError(t, prepareTestDatabase())

	member := &Member{
		Uid:    uidA,
		Name:   "test",
		Avatar: "/test",
	}

	aff, err := s.UpdateUser(member)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)

	a := new(Member)
	ok, err := x.Where("uid = ?", uidA).Get(a)

	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, member.Name, a.Name)
	assert.Equal(t, member.Avatar, a.Avatar)
}

func TestFindMember(t *testing.T) {
	assert.NoError(t, prepareTestDatabase())

	m, ok, err := s.Find(uidA)

	at, _ := time.ParseInLocation("2006-01-02 15:04:05", "2019-06-26 13:52:32", time.Local)

	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, &Member{
		Id:         1,
		Uid:        uidA,
		Name:       "testA",
		Avatar:     "/",
		Type:       Player,
		CreateAt:   at,
	}, m)
}
