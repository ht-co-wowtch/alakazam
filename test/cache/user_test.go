package cache

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"testing"
)

func TestGetUser(t *testing.T) {
	uid := "fc5d9a0855bf429dbd2f08af9be9efd8"
	key := "011eab06-4f86-4a78-8b89-65633fe77559"
	expectedRoomId := "6318a4f786e64c6487a30687e9df3a13"
	expectedName := "test"
	expectedStatus := permission.PlayDefaultPermission
	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1NTg2ODgwMTcsImlzcyI6ImNwdyIsImF1ZCI6ImNoYXQiLCJzZXNzaW9uX3Rva2VuIjoiY2MwZGEwNjMwMzg2NGFjNWJlZGJhMzViNWQ1NWNkZTEiLCJ1aWQiOiI5ODQxNjQyNmU0OTQ0ZWUyODhkOTQ3NWNkODBiYzUwMSJ9.sfIKY2nZ6b4pWGrAmNUV8ndkQRmnv2fKdg80cW3FS9Y"

	err := d.SetUser(uid, key, expectedRoomId, expectedName, "", expectedStatus)
	assert.Nil(t, err)

	rId, name, status, err := d.GetUser(uid, key)

	assert.Nil(t, err)
	assert.Equal(t, expectedRoomId, rId)
	assert.Equal(t, expectedName, name)
	assert.Equal(t, expectedStatus, status)
}
