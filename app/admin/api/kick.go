package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"net/http"
)

func (s *httpServer) kick(c *gin.Context) error {
	keys, err := s.member.Kick(c.Param("uid"))
	if err != nil {
		return err
	}

	kLen := len(keys)
	if kLen > 0 {
		err = s.message.Kick(message.KickMessage{
			Message: "因为某些原因你被踢出房间",
			Keys:    keys,
		})
		c.JSON(http.StatusOK, gin.H{
			"msg": fmt.Sprintf("该用户所在的%d个连线将中断", kLen),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "该用户不在线",
		})
	}
	return nil
}
