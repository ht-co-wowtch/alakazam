package store

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestCreateUser(t *testing.T) {
	mock.ExpectExec("^INSERT INTO members \\(uid, permission, create_at\\) VALUES \\(\\?, \\?, CURRENT_TIMESTAMP\\)").
		WithArgs("1", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	aff, err := store.CreateUser("1", 1)

	assert.Nil(t, err)
	assert.Equal(t, aff, int64(1))
}

func TestFindUserPermission(t *testing.T) {
	mock.ExpectQuery("^SELECT permission, is_blockade FROM members WHERE uid = \\?").
		WithArgs("1").
		WillReturnRows(
			sqlmock.NewRows([]string{"permission", "is_blockade"}).
				AddRow(1, 0),
		)
	permission, isBlockade, err := store.FindUserPermission("1")

	assert.Nil(t, err)
	assert.Equal(t, 1, permission)
	assert.False(t, isBlockade)
}
