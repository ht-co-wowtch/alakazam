package store

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateRoom(t *testing.T) {
	mock.ExpectExec("INSERT INTO rooms \\(room_id, is_message, is_bonus, is_follow, day_limit, amount_limit, dml_limit\\) VALUES (.+)").
		WithArgs(sqlmock.AnyArg(), true, false, false, 5, 1000, 100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	aff, err := store.CreateRoom(Room{
		IsMessage: true,
		Limit: Limit{
			Day:    5,
			Amount: 1000,
			Dml:    100,
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetRoom(t *testing.T) {
	//mock.ExpectQuery("SELECT \\* FORM rooms WHERE room_id = \\?").
	//	WithArgs(1000).
	//	WillReturnRows(sql.ErrNoRows)
	//
	//r, err := store.GetRoom(1000)
	//
	//assert.Equal(t, sql.ErrNoRows, err)
	//assert.Empty(t, r)
	//
	//if err := mock.ExpectationsWereMet(); err != nil {
	//	t.Errorf("there were unfulfilled expectations: %s", err)
	//}
}
