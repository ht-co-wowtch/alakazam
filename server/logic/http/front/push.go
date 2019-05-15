package front

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/server/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic"
	response "gitlab.com/jetfueltw/cpw/alakazam/server/logic/http"
	"net/http"
)

// 單一房間推送訊息
func (s *Server) pushRoom(c *gin.Context) {
	arg := new(logic.PushRoomForm)
	if err := c.ShouldBind(arg); err != nil {
		response.ErrorE(c, errors.PushRoomDataError)
		return
	}
	if err := s.logic.PushRoom(c, arg); err != nil {
		response.Errors(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
