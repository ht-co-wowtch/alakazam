package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"net/http"
)

func (s *Server) setBlockade(c *gin.Context) error {
	ok, err := s.logic.SetBlockade(c.Param("uid"))
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNoRows
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (s *Server) removeBlockade(c *gin.Context) error {
	ok, err := s.logic.RemoveBlockade(c.Param("uid"))
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNoRows
	}
	c.Status(http.StatusNoContent)
	return nil
}
