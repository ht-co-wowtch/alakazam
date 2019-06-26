package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetBlockade(t *testing.T) {
	assert.NoError(t, prepareTestDatabase())

	aff, err := s.SetBlockade(uidA, "")

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)

	m := new(Member)
	ok, err := x.Where("uid = ? AND is_blockade = 1", uidA).Get(m)

	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestDeleteBanned(t *testing.T) {
	assert.NoError(t, prepareTestDatabase())

	aff, err := s.DeleteBanned(uidB)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)

	m := new(Member)
	ok, err := x.Where("uid = ? AND is_blockade = 0", uidA).Get(m)

	assert.Nil(t, err)
	assert.True(t, ok)
}
