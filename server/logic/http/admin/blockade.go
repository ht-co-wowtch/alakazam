package admin

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	response "gitlab.com/jetfueltw/cpw/alakazam/server/logic/http"
	"net/http"
)

func (s *Server) setBlockade(c *gin.Context) {
	var params struct {
		Uid    string `json:"uid" binding:"required"`
		Remark string `json:"remark" binding:"required"`
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		response.ErrorE(c, errors.DataError)
		return
	}
	if err := s.logic.SetBlockade(params.Uid, params.Remark); err != nil {
		response.ErrorE(c, errors.FailureError)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) removeBlockade(c *gin.Context) {
	var params struct {
		Uid string `form:"uid" binding:"required"`
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		response.ErrorE(c, errors.DataError)
		return
	}
	if err := s.logic.RemoveBlockade(params.Uid); err != nil {
		response.ErrorE(c, errors.FailureError)
		return
	}
	c.Status(http.StatusNoContent)
}
