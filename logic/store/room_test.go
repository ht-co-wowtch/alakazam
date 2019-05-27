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
	roomId := "82ea16cd2d6a49d887440066ef739669"
	room := Room{
		IsMessage: true,
		Limit: Limit{
			Day:    1,
			Amount: 1000,
			Dml:    100,
		},
	}
	mock.ExpectQuery("^SELECT \\* FROM rooms WHERE room_id = \\?").
		WithArgs(roomId).
		WillReturnRows(
			sqlmock.NewRows([]string{"room_id", "is_message", "is_bonus", "is_follow", "day_limit", "amount_limit", "dml_limit"}).
				AddRow(room.RoomId, room.IsMessage, false, false, room.Limit.Day, room.Limit.Amount, room.Limit.Dml),
		)

	r, err := store.GetRoom(roomId)

	assert.Nil(t, err)
	assert.Equal(t, room, r)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
