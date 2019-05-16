package logic

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnect(t *testing.T) {
	c, err := l.Connect("", []byte(`{"token":"","room_id":"1000"}`))
	assert.Nil(t, err)

	assert.NotEmpty(t, c.Uid)
	assert.NotEmpty(t, c.Key)
	assert.NotEmpty(t, c.Name)
	assert.NotEmpty(t, c.Permission)
	assert.Equal(t, "1000", c.RoomId)

	r, n, s, err := d.UserData(c.Uid, c.Key)

	assert.Nil(t, err)
	assert.Equal(t, c.RoomId, r)
	assert.Equal(t, c.Name, n)
	assert.Equal(t, c.Permission, s)
}
