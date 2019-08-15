package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"net/http"
)

// 單一房間推送訊息
func (s *httpServer) pushRoom(c *gin.Context) error {
	p := new(message.PushRoom)
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
	if err := s.message.PushRoom(c, p); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
