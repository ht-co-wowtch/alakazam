package admin

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/server/errors"
	response "gitlab.com/jetfueltw/cpw/alakazam/server/logic/http"
	"net/http"
)

// 所有房間推送
func (s *Server) setBanned(c *gin.Context) {
	var params struct {
		Uid     string `json:"uid"`
		Expired int    `json:"expired"`
		Remark  string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		response.ErrorE(c, errors.DataError)
		return
	}
	if err := s.logic.SetBanned(params.Uid, params.Remark, params.Expired); err != nil {
		response.ErrorE(c, errors.FailureError)
		return
	}
	c.Status(http.StatusNoContent)
}
