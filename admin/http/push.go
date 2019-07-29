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
	if err := s.logic.PushAll(p); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
