package cache

import (
	"github.com/rafaeljusto/redigomock"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
)

func TestSetUser(t *testing.T) {
	uid := "82ea16cd2d6a49d887440066ef739669"
	key := "0b7f8111-8781-4574-8cb8-2eda0adb7598"
	roomId := "1000"
	name := "test"
	mockSetUser(uid, key, roomId, name)

	err := d.SetUser(uid, key, roomId, name, "", permission.PlayDefaultPermission)
	assert.Nil(t, err)
}

func TestRefreshUserExpire(t *testing.T) {
	ok, err := mockRefreshUserExpire("82ea16cd2d6a49d887440066ef739669")

	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestDeleteUser(t *testing.T) {
	uid := "82ea16cd2d6a49d887440066ef739669"
	key := "0b7f8111-8781-4574-8cb8-2eda0adb7598"

	mockDeleteUser(uid, key)
	ok, err := d.DeleteUser(uid, key)

	assert.Nil(t, err)
	assert.True(t, ok)
}

func mockDeleteUser(uid string, key string) *redigomock.Cmd {
	return mock.Command("HDEL", keyUidInfo(uid), key).
		Expect([]byte(`true`))
}

func mockSetUser(uid string, key string, roomId string, name string) {
	mock.Command("HMSET", keyUidInfo(uid), key, roomId, hashNameKey, name, hashStatusKey, permission.PlayDefaultPermission, hashServerKey, "").
		Expect("")
	mock.Command("EXPIRE", keyUidInfo(uid), expireSec).
		Expect("")
}

func mockRefreshUserExpire(uid string) (bool, error) {
	mock.Command("EXPIRE", keyUidInfo(uid), expireSec).
		Expect([]byte(`true`))
	ok, err := d.RefreshUserExpire(uid)
	return ok, err

}
