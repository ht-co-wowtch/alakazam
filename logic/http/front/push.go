package front

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	response "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"net/http"
)

// 單一房間推送訊息
func (s *Server) pushRoom(c *gin.Context) {
	arg := new(logic.PushRoomForm)
	if err := c.ShouldBindJSON(arg); err != nil {
		response.ErrorE(c, errors.DataError)
		return
	}
	if err := s.logic.PushRoom(arg); err != nil {
		response.Errors(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
