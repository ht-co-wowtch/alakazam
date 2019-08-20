package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *httpServer) getMessage(c *gin.Context) error {
	room := c.Param("room")
	msg, err := s.room.GetMessage(room)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"data": msg,
	})
	return nil
}
