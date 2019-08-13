package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

// 設定禁言
func (s *Server) setBanned(c *gin.Context) error {
	params := struct {
		Uid     string `form:"uid" binding:"required,len=32"`
		Expired int    `json:"expired" binding:"required"`
	}{
		Uid: c.Param("uid"),
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}
	if err := s.logic.SetBanned(params.Uid, params.Expired); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

// 解除禁言
func (s *Server) removeBanned(c *gin.Context) error {
	params := struct {
		Uid string `form:"uid" binding:"required,len=32"`
	}{
		Uid: c.Param("uid"),
	}
	if err := binding.Validator.ValidateStruct(&params); err != nil {
		return err
	}
	if err := s.logic.RemoveBanned(params.Uid); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
