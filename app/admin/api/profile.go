package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *httpServer) renew(c *gin.Context) error {
	user, err := s.nidoran.Auth(c.Param("token"))
	if err != nil {
		return err
	}
	if err := s.member.Update(user.Uid, user.Name, user.Gender); err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"uid": user.Uid,
	})
	return nil
}
