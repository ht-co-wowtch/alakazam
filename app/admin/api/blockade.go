package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func (s *httpServer) setBlockade(c *gin.Context) error {
	keys := []string{}
	uid := c.Param("uid")
	id, err := getId(c)
	if err != nil {
		return err
	}

	if id == 0 {
		if err := s.member.SetBlockadeAll(uid, true); err != nil {
			return err
		}
		if keys, err = s.member.Kick(uid); err != nil {
			return err
		}
	} else {
		if err := s.member.SetBlockade(uid, id, true); err != nil {
			return err
		}
		if keys, err = s.member.GetRoomKeys(uid, id); err != nil {
			return err
		}
	}

	var msg string
	if len(keys) == 0 {
		msg = "封锁成功"
	} else {
		err = s.message.Kick("你被踢出房间，因为被封锁", keys)

		if err != nil {
			log.Error("kick member message for set blockade", zap.Error(err), zap.String("uid", uid))
			msg = "封锁成功，但执行聊天室踢人失败"
		} else {
			msg = fmt.Sprintf("封锁成功，將執行中断该用户所在的%d个连线", len(keys))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": msg,
	})
	return nil
}

func (s *httpServer) removeBlockade(c *gin.Context) error {
	id, err := getId(c)
	if err != nil {
		return err
	}

	if id == 0 {
		if err := s.member.SetBlockadeAll(c.Param("uid"), false); err != nil {
			return err
		}
	} else if err := s.member.SetBlockade(c.Param("uid"), id, false); err != nil {
		return err
	}

	c.Status(http.StatusNoContent)
	return nil
}

func getId(c *gin.Context) (int, error) {
	id := 0
	idr := c.Param("id")

	if idr != "" {
		var err error
		id, err = strconv.Atoi(idr)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}
