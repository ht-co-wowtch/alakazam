package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"net/http"
)

type PushRoomForm struct {
	// 要廣播的房間
	RoomId []string `json:"room_id" binding:"required"`

	// user push message
	Message string `json:"message" binding:"required,max=250"`

	// 訊息是否頂置
	Top bool `json:"top"`
}

// 多房間推送
func (s *httpServer) push(c *gin.Context) error {
	p := new(PushRoomForm)
	if err := c.ShouldBindJSON(p); err != nil {
		return err
	}

	var rid []int64
	for _, v := range p.RoomId {
		r, err := s.room.Get(v)
		if err != nil {
			return err
		}
		rid = append(rid, int64(r.Id))
	}

	msg := message.AdminMessage{
		Rooms:   p.RoomId,
		Rids:    rid,
		Message: p.Message,
		IsTop:   p.Top,
	}
	if err := s.message.SendForAdmin(msg); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

// TODO 待完成
func (s *httpServer) deleteTopMessage(c *gin.Context) error {
	c.Status(http.StatusNoContent)
	return nil
}
