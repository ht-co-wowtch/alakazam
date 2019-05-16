package admin

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/server/errors"
	response "gitlab.com/jetfueltw/cpw/alakazam/server/logic/http"
	"net/http"
)

// 所有房間推送
func (s *Server) pushAll(c *gin.Context) {
	var arg struct {
		// 要廣播的房間
		RoomId []string `json:"room_id" binding:"required"`

		// user push message
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&arg); err != nil {
		response.ErrorE(c, errors.DataError)
		return
	}
	if err := s.logic.PushAll(arg.RoomId, arg.Message); err != nil {
		response.Errors(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
