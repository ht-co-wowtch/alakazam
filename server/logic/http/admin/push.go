package admin

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/server/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic"
	response "gitlab.com/jetfueltw/cpw/alakazam/server/logic/http"
	"net/http"
)

// 所有房間推送
func (s *Server) pushAll(c *gin.Context) {
	arg := new(logic.PushRoomAllForm)
	if err := c.ShouldBind(arg); err != nil {
		response.ErrorE(c, errors.DataError)
		return
	}
	if err := s.logic.PushAll(arg); err != nil {
		response.Errors(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
