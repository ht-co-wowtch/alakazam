package dao

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBanned(t *testing.T) {
	now := time.Now().Add(time.Second * 5)
	err := d.SetBanned("123", 5)
	assert.Nil(t, err)

	ex, ok, err := d.GetBanned("123")
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, ex.Format(time.RFC3339), now.Format(time.RFC3339))
}

func TestExpiredBanned(t *testing.T) {
	err := d.SetBanned("123", 1)
	assert.Nil(t, err)
	time.Sleep(time.Second)

	ex, ok, err := d.GetBanned("123")
	assert.Nil(t, err)
	assert.False(t, ok)
	assert.True(t, ex.IsZero())
}

func TestDeleteBanned(t *testing.T) {
	err := d.AddMapping("123", "", "", "", "", 2)
	assert.Nil(t, err)

	err = d.SetBanned("123", 10)
	assert.Nil(t, err)
	time.Sleep(time.Second)

	err = d.DelBanned("123")
	assert.Nil(t, err)

	_, ok, err := d.GetBanned("123")
	assert.Nil(t, err)
	assert.False(t, ok)

	_, _, s, err := d.GetUser("123", "")
	assert.Nil(t, err)
	assert.Equal(t, 2, s)
}
