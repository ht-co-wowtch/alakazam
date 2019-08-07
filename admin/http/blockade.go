package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Server) setBlockade(c *gin.Context) error {
	if err := s.logic.SetBlockade(c.Param("uid")); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (s *Server) removeBlockade(c *gin.Context) error {
	if _, err := s.logic.RemoveBlockade(c.Param("uid")); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
