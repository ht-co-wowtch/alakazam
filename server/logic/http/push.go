package http

import (
	"github.com/gin-gonic/gin"
	Err "gitlab.com/jetfueltw/cpw/alakazam/server/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic"
	"net/http"
)

// 單一房間推送訊息
func (s *Server) pushRoom(c *gin.Context) {
	arg := new(logic.PushRoomForm)
	if err := c.ShouldBind(arg); err != nil {
		errorE(c, Err.PushRoomDataError)
		return
	}
	if err := s.logic.PushRoom(c, arg); err != nil {
		errors(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// 所有房間推送
func (s *Server) pushAll(c *gin.Context) {
	arg := new(logic.PushRoomAllForm)
	if err := c.ShouldBind(arg); err != nil {
		errorE(c, Err.PushRoomDataError)
		return
	}
	if err := s.logic.PushAll(c, arg); err != nil {
		errors(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
