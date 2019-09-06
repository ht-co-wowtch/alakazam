package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"net/http"
	"strconv"
	"time"
)

// TODO 先兼容msg_id 跟 timestamp，待前端轉換完成在移除msg_id
func (s *httpServer) getMessage(c *gin.Context) error {
	rid, err := strconv.Atoi(c.Param("room"))
	if err != nil {
		return err
	}
	msg_id, err := strconv.Atoi(c.DefaultQuery("msg_id", "0"))
	if err != nil {
		return err
	}

	var msg interface{}

	if msg_id != 0 {
		if msg, err = s.history.Get(rid, msg_id); err != nil {
			return err
		}
	} else {
		var timestamp int64
		if timestampStr := c.Query("timestamp"); timestampStr == "" {
			timestamp = time.Now().Unix()
		} else {
			timestamp, err = strconv.ParseInt(c.Query("timestamp"), 10, 0)
			if err != nil {
				return errdefs.InvalidParameter(4000, "时间格式错误", nil)
			}
		}
		if msg, err = s.history.GetV2(int32(rid), time.Unix(timestamp, 0)); err != nil {
			return err
		}
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
