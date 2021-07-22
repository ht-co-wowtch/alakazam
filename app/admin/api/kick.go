package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// 將會員踢出房間
func (s *httpServer) kick(c *gin.Context) error {
	keys, err := s.member.Kick(c.Param("uid"))
	if err != nil {
		return err
	}

	kLen := len(keys)
	if kLen > 0 {
		// 發送踢出房間訊息
		err = s.message.Kick("因为某些原因你被踢出房间", keys)
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
