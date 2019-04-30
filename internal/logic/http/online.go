package http

import (
	"github.com/gin-gonic/gin"
)

// 根據房間type與room id取房間在線人數
func (s *Server) onlineRoom(c *gin.Context) {
	var arg struct {
		Type  string   `form:"type" binding:"required"`
		Rooms []string `form:"rooms" binding:"required"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	res, err := s.logic.OnlineRoom(c, arg.Type, arg.Rooms)
	if err != nil {
		result(c, nil, RequestErr)
		return
	}
	result(c, res, OK)
}