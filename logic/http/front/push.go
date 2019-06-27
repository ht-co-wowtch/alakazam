package front

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"net/http"
)

// 單一房間推送訊息
func (s *Server) pushRoom(c *gin.Context) error {
	arg := new(logic.PushRoom)
	if err := c.ShouldBindJSON(arg); err != nil {
		return errors.DataError
	}
	if err := s.logic.PushRoom(c, arg); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
