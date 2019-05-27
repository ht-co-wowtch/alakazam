package store

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetBlockade(t *testing.T) {
	uid := "82ea16cd2d6a49d887440066ef739669"
	mockSetBlockade(uid)

	aff, err := store.SetBlockade(uid, "")

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)
}

func TestStore_DeleteBanned(t *testing.T) {
	uid := "82ea16cd2d6a49d887440066ef739669"
	mockDeleteBanned(uid)

	aff, err := store.DeleteBanned(uid)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)
}

func mockSetBlockade(uid string) *sqlmock.ExpectedExec {
	return mockBanned(uid, true)
}

func mockDeleteBanned(uid string) *sqlmock.ExpectedExec {
	return mockBanned(uid, false)
}

func mockBanned(uid string, isBlockade bool) *sqlmock.ExpectedExec {
	return mock.ExpectExec("UPDATE members SET is_blockade = \\? WHERE uid = \\?").
		WithArgs(isBlockade, uid).
		WillReturnResult(sqlmock.NewResult(1, 1))
}
