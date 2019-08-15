package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"net/http"
)

// 多房間推送
func (s *httpServer) push(c *gin.Context) error {
	p := new(message.PushRoomForm)
	if err := c.ShouldBindJSON(p); err != nil {
		return err
	}
	id, err := s.message.PushMessage(p)
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
