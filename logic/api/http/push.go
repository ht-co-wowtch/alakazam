package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"net/http"
	"time"
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
	if err := s.room.IsMessage(p.RoomId, p.RoomStatus, p.Uid, c.GetString("token")); err != nil {
		return err
	}
	msg := message.Message{
		Uid:     p.Uid,
		Name:    p.Name,
		Avatar:  "",
		Message: p.Message,
		Time:    time.Now().Format("15:04:05"),
	}
	if err := s.message.Send(p.RoomId, msg); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
