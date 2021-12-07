package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/ht-co/cpw/micro/errdefs"
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
			return errdefs.InvalidParameter(4000, "时间格式错误")
		}
	}

	msg, err := s.history.Get(int32(rid), time.Unix(timestamp, 0))
	if err != nil {
		return err
	}

	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.JSON(http.StatusOK, gin.H{
		"data": msg,
	})
	return nil
}
