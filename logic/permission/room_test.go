package permission

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"testing"
)

func TestRoomInt(t *testing.T) {
	actual := ToRoomInt(store.Room{
		IsMessage: true,
		IsFollow:  true,
		IsBonus:   false,
		Limit: store.Limit{
			Day:    1,
			Amount: 1000,
		},
	})

	expected := RoomDefaultPermission - sendBonus - getBonus + dml + recharge

	assert.Equal(t, expected, actual)
}
