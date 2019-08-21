package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"net/http"
)

type PushRoomForm struct {
	// 要廣播的房間
	RoomId []int32 `json:"room_id" binding:"required"`

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

	msg := message.AdminMessage{
		Rooms:   p.RoomId,
		Message: p.Message,
		IsTop:   p.Top,
	}
	id, err := s.message.SendForAdmin(msg)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}

// TODO 待完成
func (s *httpServer) deleteTopMessage(c *gin.Context) error {
	c.Status(http.StatusNoContent)
	return nil
}
