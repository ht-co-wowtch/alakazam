package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic"
)

// 以room id來推送訊息
func (s *Server) pushRoom(c *gin.Context) {
	arg := new(logic.PushRoomForm)
	if err := c.Bind(arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	if err := s.logic.PushRoom(c, arg); err != nil {
		errors(c, ServerErr, err.Error())
		return
	}
	result(c, nil, OK)
}

// 所有房間推送
func (s *Server) pushAll(c *gin.Context) {
	arg := new(logic.PushRoomAllForm)
	if err := c.Bind(arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	if err := s.logic.PushAll(c, arg); err != nil {
		errors(c, ServerErr, err.Error())
		return
	}
	result(c, nil, OK)
}
