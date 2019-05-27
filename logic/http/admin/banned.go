package admin

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	response "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"net/http"
)

// 設定禁言
func (s *Server) setBanned(c *gin.Context) {
	var params struct {
		Uid     string `json:"uid" binding:"required"`
		Expired int    `json:"expired" binding:"required"`
		Remark  string `json:"remark" binding:"required"`
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

// 解除禁言
func (s *Server) removeBanned(c *gin.Context) {
	var params struct {
		Uid string `form:"uid" binding:"required"`
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		response.ErrorE(c, errors.DataError)
		return
	}
	if err := s.logic.RemoveBanned(params.Uid); err != nil {
		response.ErrorE(c, errors.FailureError)
		return
	}
	c.Status(http.StatusNoContent)
}
