package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"net/http"
	"strconv"
	"time"
)

func (s *httpServer) getMessage(c *gin.Context) error {
	rid, err := strconv.Atoi(c.Param("room"))
	if err != nil {
		return err
	}

	var timestamp int64
	timestampStr := c.Query("timestamp")
	if timestampStr == "" {
		timestamp = time.Now().Unix()
	} else {
		timestamp, err = strconv.ParseInt(timestampStr, 10, 0)
		if err != nil {
			return errdefs.InvalidParameter(4000, "时间格式错误", nil)
		}
	}

	msg, err := s.history.Get(int32(rid), time.Unix(timestamp, 0))
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
