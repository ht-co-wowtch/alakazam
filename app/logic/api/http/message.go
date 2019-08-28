package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
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

func (s *httpServer) getTopMessage(c *gin.Context) error {
	rid, err := strconv.Atoi(c.Param("room"))
	if err != nil {
		return err
	}

	var msg interface{}
	msg, err = s.room.GetTopMessage(rid)
	if err != nil {
		if err != errors.ErrNoRows {
			return err
		}
		msg = ""
	}
	c.JSON(http.StatusOK, gin.H{
		"data": msg,
	})
	return nil
}
