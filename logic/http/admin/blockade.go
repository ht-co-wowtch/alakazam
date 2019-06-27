package admin

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"net/http"
)

func (s *Server) setBlockade(c *gin.Context) error {
	var params struct {
		Uid    string `json:"uid" binding:"required"`
		Remark string `json:"remark" binding:"required"`
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		return errors.DataError
	}

	if !s.logic.SetBlockade(params.Uid, params.Remark) {
		return errors.FailureError
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (s *Server) removeBlockade(c *gin.Context) error {
	var params struct {
		Uid string `form:"uid" binding:"required"`
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		return errors.DataError
	}
	if !s.logic.RemoveBlockade(params.Uid) {
		return errors.FailureError
	}
	c.Status(http.StatusNoContent)
	return nil
}
