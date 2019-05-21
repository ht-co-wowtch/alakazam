package admin

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	response "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"net/http"
)

func (s *Server) SetRoom(c *gin.Context) {
	var params store.Room
	if err := c.ShouldBindJSON(&params); err != nil {
		response.ErrorE(c, errors.DataError)
		return
	}
	if !s.logic.SetRoom(params) {
		response.ErrorE(c, errors.FailureError)
		return
	}
	c.Status(http.StatusNoContent)
}
