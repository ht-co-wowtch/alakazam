package store

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetRoom(t *testing.T) {
	mock.ExpectExec("INSERT INTO rooms \\(room_id, is_message, is_bonus, is_follow, day_limit, amount_limit, dml_limit\\) VALUES (.+)").
		WithArgs(1000, true, false, false, 5, 1000, 100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	aff, err := store.SetRoom(Room{
		RoomId:    1000,
		IsMessage: true,
		Limit: Limit{
			Day:    5,
			Amount: 1000,
			Dml:    100,
		},
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)
}
