package dao

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/dao"
	"testing"
)

func AddUser(d *dao.Dao, t *testing.T, uid, key, roomId, name string, status int) {
	err := d.AddMapping(uid, key, roomId, name, "", status)
	assert.Nil(t, err)
}

func GetUser(d *dao.Dao, t *testing.T, uid, key string) (string, string, int) {
	roomId, name, status, err := d.UserData(uid, key)
	assert.Nil(t, err)
	return roomId, name, status
}
