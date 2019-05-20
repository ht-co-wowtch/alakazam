package dao

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"testing"
)

func AddUser(d *cache.Cache, t *testing.T, uid, key, roomId, name string, status int) {
	err := d.SetUser(uid, key, roomId, name, "", status)
	assert.Nil(t, err)
}

func GetUser(d *cache.Cache, t *testing.T, uid, key string) (string, string, int) {
	roomId, name, status, err := d.GetUser(uid, key)
	assert.Nil(t, err)
	return roomId, name, status
}
