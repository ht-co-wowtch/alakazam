package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"net/http"
)

// 所有房間推送
func (s *Server) push(c *gin.Context) error {
	p := new(logic.PushRoomAllForm)
	if err := c.ShouldBindJSON(p); err != nil {
		return err
	}
	id, err := s.logic.PushMessage(p)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}

// TODO 待完成
func (s *Server) deleteTopMessage(c *gin.Context) error {
	c.Status(http.StatusNoContent)
	return nil
}
