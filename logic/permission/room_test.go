package permission

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"strconv"
	"testing"
)

func TestRoomInt(t *testing.T) {
	testCases := []struct {
		room     models.Room
		expected int
	}{
		{
			room: models.Room{
				IsMessage: true,
			},
			expected: Message,
		},
		{
			room: models.Room{
				IsFollow: true,
			},
			expected: getFollow + sendFollow,
		},
		{
			room: models.Room{
				DayLimit: 1,
				DmlLimit: 1000,
			},
			expected: money,
		},
		{
			room: models.Room{
				IsFollow: true,
				DayLimit: 1,
				DmlLimit: 1000,
			},
			expected: money + getFollow + sendFollow,
		},
		{
			room: models.Room{
				DmlLimit: 1000,
			},
			expected: 0,
		},
		{
			room: models.Room{
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
