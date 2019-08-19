package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
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
	r, err := s.room.Get(p.H.Room)
	if err != nil {
		return err
	}

	msg := message.Messages{
		Rooms:   []string{p.H.Room},
		Rids:    []int64{int64(r.Id)},
		Mid:     int64(p.H.Mid),
		Uid:     p.Uid,
		Name:    p.H.Name,
		Message: p.Message,
	}
	id, err := s.message.Send(msg)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}
