package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type messageReq struct {
	RoomId int `json:"room_id" binding:"required,max=10000"`

	// user push message
	Message string `json:"message" binding:"required,max=100"`

	uid string `json:"-"`

	token string `json:"-"`
}

// 單一房間推送訊息
func (s *httpServer) pushRoom(c *gin.Context) error {
	var p messageReq
	if err := c.ShouldBindJSON(&p); err != nil {
		return err
	}

	p.token = c.GetString("token")
	p.uid = c.GetString("uid")

	id, err := s.msg.user(p)

	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"id": id,
		})
	}

	return err
}
