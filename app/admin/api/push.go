package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"net/http"
	"strconv"
)

type pushRoomForm struct {
	// 要廣播的房間
	RoomId []int32 `json:"room_id" binding:"required"`

	// user push message
	Message string `json:"message" binding:"required,max=250"`

	// 訊息是否頂置
	Top bool `json:"top"`
}

// 多房間推送
func (s *httpServer) push(c *gin.Context) error {
	p := new(pushRoomForm)
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

func (s *httpServer) deleteTopMessage(c *gin.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	msgId := int64(id)
	rid, msg, err := s.room.GetTopMessage(msgId)
	if err != nil {
		return err
	}
	if err := s.message.CloseTop(msgId, rid); err != nil {
		return err
	}
	if err := s.room.DeleteTopMessage(msgId); err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"id":      msgId,
		"message": msg.Message,
		"room_id": rid,
	})
	return nil
}
