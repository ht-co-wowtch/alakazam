package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
)

type PushRoom struct {
	member.User

	// user push message
	Message string `json:"message" binding:"required,max=100"`
}

// 單一房間推送訊息
func (s *httpServer) pushRoom(c *gin.Context) error {
	p := new(PushRoom)
	if err := c.ShouldBindJSON(p); err != nil {
		return err
	}
	if err := s.member.Auth(&p.User); err != nil {
		return err
	}
	if err := s.room.Auth(&p.User); err != nil {
		return err
	}
	if err := s.room.IsMessage(p.H.Room, p.RoomStatus, p.Uid, c.GetString("token")); err != nil {
		return err
	}

	msg := message.Messages{
		Rooms:   []int32{int32(p.H.Room)},
		Mid:     int64(p.H.Mid),
		Uid:     p.Uid,
		Name:    p.H.Name,
		Message: p.Message,
	}
	id, err := s.message.Send(msg)
	if err != nil {
		if err == errors.ErrRateSameMsg {
			if err := s.member.SetBanned(p.Uid, 10*60, true); err != nil {
				log.Error("set banned for rate same message", zap.Error(err), zap.String("uid", p.Uid))
			}
		}
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}
