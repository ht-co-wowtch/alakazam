package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"net/http"
)

func (s *httpServer) setBlockade(c *gin.Context) error {
	ok, err := s.member.SetBlockade(c.Param("uid"))
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNoRows
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (s *httpServer) removeBlockade(c *gin.Context) error {
	ok, err := s.member.RemoveBlockade(c.Param("uid"))
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNoRows
	}
	c.Status(http.StatusNoContent)
	return nil
}
