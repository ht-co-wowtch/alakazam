package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"net/http"
)

// 設定禁言
func (s *Server) setBanned(c *gin.Context) error {
	var params struct {
		Uid     string `json:"uid" binding:"required"`
		Expired int    `json:"expired" binding:"required"`
		Remark  string `json:"remark" binding:"required"`
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		return errors.DataError
	}
	if err := s.logic.SetBanned(params.Uid, params.Remark, params.Expired); err != nil {
		return errors.FailureError
	}
	c.Status(http.StatusNoContent)
	return nil
}

// 解除禁言
func (s *Server) removeBanned(c *gin.Context) error {
	var params struct {
		Uid string `form:"uid" binding:"required"`
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		return errors.DataError
	}
	if err := s.logic.RemoveBanned(params.Uid); err != nil {
		return errors.FailureError
	}
	c.Status(http.StatusNoContent)
	return nil
}
