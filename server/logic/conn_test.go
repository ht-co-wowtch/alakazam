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

func TestDisconnect(t *testing.T) {
	err := d.AddMapping("123", "", "", "", "", 0)
	assert.Nil(t, err)

	has, err := l.Disconnect("123", "", "")

	assert.True(t, has)
	assert.Nil(t, err)
}

func TestChangeRoom(t *testing.T) {
	err := d.AddMapping("456", "", "1000", "", "", 0)
	assert.Nil(t, err)

	err = l.ChangeRoom("456", "", "1001")
	assert.Nil(t, err)

	r, _, _, err := d.UserData("456", "")
	assert.Nil(t, err)
	assert.Equal(t, "1001", r)
}

func TestHeartbeat(t *testing.T) {
	err := d.AddMapping("789", "", "1000", "", "", 0)
	assert.Nil(t, err)

	err = l.Heartbeat("789", "", "", "", "")
	assert.Nil(t, err)
}
