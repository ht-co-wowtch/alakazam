package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Server) setBlockade(c *gin.Context) error {
	params := struct {
		Uid    string `json:"uid" binding:"required,len=32"`
		Remark string `json:"remark" binding:"required,max=50"`
	}{
		Uid: c.Param("uid"),
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}
	if err := s.logic.SetBlockade(params.Uid, params.Remark); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (s *Server) removeBlockade(c *gin.Context) error {
	params := struct {
		Uid string `json:"uid" binding:"required,len=32"`
	}{
		Uid: c.Param("uid"),
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		return err
	}
	if _, err := s.logic.RemoveBlockade(params.Uid); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
