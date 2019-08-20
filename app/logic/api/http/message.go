package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *httpServer) getMessage(c *gin.Context) error {
	room := c.Param("room")
	r, err := s.room.Get(room)
	if err != nil {
		return err
	}
	msg, err := s.history.Get(r.Id)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"data": msg,
	})
	return nil
}
