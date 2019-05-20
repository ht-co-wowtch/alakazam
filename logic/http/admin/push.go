package admin

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	response "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"net/http"
)

// 所有房間推送
func (s *Server) pushAll(c *gin.Context) {
	p := new(logic.PushRoomAllForm)
	if err := c.ShouldBindJSON(p); err != nil {
		response.ErrorE(c, errors.DataError)
		return
	}
	if err := s.logic.PushAll(p); err != nil {
		response.Errors(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
