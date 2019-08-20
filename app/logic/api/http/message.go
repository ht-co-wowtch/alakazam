package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (s *httpServer) getMessage(c *gin.Context) error {
	room := c.Param("room")
	r, err := s.room.Get(room)
	if err != nil {
		return err
	}
	msg_id, err := strconv.Atoi(c.DefaultQuery("msg_id", "0"))
	if err != nil {
		return err
	}
	msg, err := s.history.Get(r.Id, msg_id)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"data": msg,
	})
	return nil
}
