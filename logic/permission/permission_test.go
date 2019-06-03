package permission

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBanned(t *testing.T) {
	Convey("禁言", t, func() {
		Convey("有禁言", func() {
			So(IsBanned(PlayDefaultPermission-Message), ShouldBeTrue)
		})
		Convey("沒禁言", func() {
			So(IsBanned(PlayDefaultPermission), ShouldBeFalse)
		})
	})
}

func TestSendBonus(t *testing.T) {
	Convey("發紅包", t, func() {
		Convey("有權限", func() {
			So(IsSendBonus(sendBonus+Message), ShouldBeTrue)
		})
		Convey("沒有權限", func() {
			So(IsSendBonus(Message), ShouldBeFalse)
		})
	})
}

func TestGetBonus(t *testing.T) {
	Convey("搶紅包", t, func() {
		Convey("有權限", func() {
			So(IsGetBonus(getBonus+Message), ShouldBeTrue)
		})
		Convey("沒有權限", func() {
			So(IsGetBonus(Message), ShouldBeFalse)
		})
	})
}

func TestSendFollow(t *testing.T) {
	Convey("發跟注", t, func() {
		Convey("有權限", func() {
			So(IsSendFollow(sendFollow+Message), ShouldBeTrue)
		})
		Convey("沒有權限", func() {
			So(IsSendFollow(Message), ShouldBeFalse)
		})
	})
}

func TestGetFollow(t *testing.T) {
	Convey("跟注", t, func() {
		Convey("有權限", func() {
			So(IsGetFollow(getFollow+Message), ShouldBeTrue)
		})
		Convey("沒有權限", func() {
			So(IsGetFollow(Message), ShouldBeFalse)
		})
	})
}

func TestIsMoney(t *testing.T) {
	Convey("後台金額發話限制", t, func() {
		Convey("有限制", func() {
			So(IsMoney(money+look), ShouldBeTrue)
		})
		Convey("沒有限制", func() {
			So(IsMoney(look), ShouldBeFalse)
		})
	})
}

func TestNewPermission(t *testing.T) {
	p := NewPermission(253)
	assert.Equal(t, &Permission{
		Message:    false,
		SendFollow: true,
		GetFollow:  true,
		SendBonus:  true,
		GetBonus:   true,
	}, p)
}
