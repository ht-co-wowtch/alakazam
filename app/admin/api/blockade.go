package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
)

func (s *httpServer) setBlockade(c *gin.Context) error {
	uid := c.Param("uid")
	ok, err := s.member.SetBlockade(uid)
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNoRows
	}

	keys, err := s.member.Kick(uid)
	if err != nil {
		return err
	}

	err = s.message.Kick(message.KickMessage{
		Message: "你被踢出房间，因为被封锁",
		Keys:    keys,
	})

	if err != nil {
		log.Error("kick member message for set blockade", zap.Error(err), zap.String("uid", uid))
		c.JSON(http.StatusOK, gin.H{
			"msg": "封锁成功，但执行聊天室踢人失败",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": fmt.Sprintf("封锁成功，將執行中断该用户所在的%d个连线", len(keys)),
		})
	}
	return nil
}

func (s *httpServer) removeBlockade(c *gin.Context) error {
	ok, err := s.member.RemoveBlockade(c.Param("uid"))
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNoRows
	}
	c.Status(http.StatusNoContent)
	return nil
}
