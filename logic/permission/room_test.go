package permission

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"strconv"
	"testing"
)

func TestRoomInt(t *testing.T) {
	testCases := []struct {
		room     store.Room
		expected int
	}{
		{
			room: store.Room{
				IsMessage: true,
			},
			expected: Message,
		},
		{
			room: store.Room{
				IsFollow: true,
			},
			expected: getFollow + sendFollow,
		},
		{
			room: store.Room{
				DayLimit: 1,
				DmlLimit: 1000,
			},
			expected: money,
		},
		{
			room: store.Room{
				IsFollow: true,
				DayLimit: 1,
				DmlLimit: 1000,
			},
			expected: money + getFollow + sendFollow,
		},
		{
			room: store.Room{
				DmlLimit: 1000,
			},
			expected: 0,
		},
		{
			room: store.Room{
				DayLimit: 1,
			},
			expected: 0,
		},
	}

	for i, v := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual := ToRoomInt(v.room)
			assert.Equal(t, v.expected, actual)
		})
	}
}
