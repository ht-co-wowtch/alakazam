package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (s *httpServer) getMessage(c *gin.Context) error {
	rid, err := strconv.Atoi(c.Param("room"))
	if err != nil {
		return err
	}
	msg_id, err := strconv.Atoi(c.DefaultQuery("msg_id", "0"))
	if err != nil {
		return err
	}
	msg, err := s.history.Get(rid, msg_id)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"data": msg,
	})
	return nil
}
