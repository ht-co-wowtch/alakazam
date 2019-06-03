package permission

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"testing"
)

func TestRoomInt(t *testing.T) {
	Convey("房間權限", t, func() {
		Convey("當房間只有發話權限", func() {
			actual := ToRoomInt(store.Room{
				IsMessage: true,
			})
			expected := Message

			Convey(fmt.Sprintf("權限只有%d", expected), func() {
				So(actual, ShouldEqual, expected)
			})
		})
		Convey("當房間只有跟注權限", func() {
			actual := ToRoomInt(store.Room{
				IsFollow: true,
			})

			expected := getFollow + sendFollow

			Convey(fmt.Sprintf("權限只有%d", expected), func() {
				So(actual, ShouldEqual, expected)
			})
		})

		Convey("當房間只有發紅包權限", func() {
			actual := ToRoomInt(store.Room{
				IsBonus: true,
			})

			expected := getBonus + sendBonus

			Convey(fmt.Sprintf("權限只有%d", expected), func() {
				So(actual, ShouldEqual, expected)
			})
		})

		Convey("當房間只有金額發話限制", func() {
			actual := ToRoomInt(store.Room{
				Limit: store.Limit{
					Day: 1,
					Dml: 1000,
				},
			})

			expected := money

			Convey(fmt.Sprintf("權限只有%d", expected), func() {
				So(actual, ShouldEqual, expected)
			})
		})

		Convey("當房間有金額發話限制與跟注權限", func() {
			actual := ToRoomInt(store.Room{
				IsFollow: true,
				Limit: store.Limit{
					Day: 1,
					Dml: 1000,
				},
			})

			expected := money + getFollow + sendFollow

			Convey(fmt.Sprintf("權限只有%d", expected), func() {
				So(actual, ShouldEqual, expected)
			})
		})

		Convey("當房間有金額發話限制沒有設定打碼量天數", func() {
			actual := ToRoomInt(store.Room{
				Limit: store.Limit{
					Dml: 1000,
				},
			})

			Convey("權限只有0", func() {
				So(actual, ShouldEqual, 0)
			})
		})

		Convey("當房間有金額發話限制沒有設定金額", func() {
			actual := ToRoomInt(store.Room{
				Limit: store.Limit{
					Day: 1,
				},
			})

			Convey("權限只有0", func() {
				So(actual, ShouldEqual, 0)
			})
		})
	})
}
