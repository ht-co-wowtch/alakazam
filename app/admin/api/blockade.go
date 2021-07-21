package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
)

// 封鎖
// 如果有帶入id(roomId)就針對個別房間進行封鎖，反之則是全站封鎖
func (s *httpServer) setBlockade(c *gin.Context) error {
	keys := []string{}
	uid := c.Param("uid")
	//id 就是roomId
	id, err := getId(c)
	if err != nil {
		return err
	}

	if id == 0 { //roomId表示全站所有房間
		// 變更db中狀態為封鎖
		if err := s.member.SetBlockadeAll(uid, true); err != nil {
			return err
		}

		if keys, err = s.member.Kick(uid); err != nil {
			return err
		}
	} else { // 單一房間
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
		err = s.message.Kick("你被踢出房间，因为被封锁", keys) // 發送踢人訊息到kafka producer

		if err != nil {
			log.Error("[blockade.go]kick member message for set blockade", zap.Error(err), zap.String("uid", uid))
			msg = "封锁成功，但执行聊天室踢人失败"
		} else {
			msg = fmt.Sprintf("封锁成功，將執行中断该用户所在的%d个连线", len(keys))
		}
	}
	/**/

	name, _ := s.member.GetUserName(uid)
	log.Debug("[blockade.go]setBlockade", zap.Int("RoomId", id), zap.String("uid", uid), zap.String("name", name))
	if id > 0 {
		// 發送封鎖訊息到kafka producer
		_, _ = s.message.SendDisplay(
			[]int32{int32(id)},
			scheme.NewRoot(),                       // 訊息發送人
			scheme.DisplayByUnBlock(name, 0, true), // 封鎖/解除封鎖訊息
		)
	}

	/**/
	c.JSON(http.StatusOK, gin.H{
		"msg": msg,
	})
	return nil
}

// 解除封鎖狀態
// 如果有帶入id(roomId)就針對個別房間進行解除封鎖，反之則是全站解除封鎖
func (s *httpServer) removeBlockade(c *gin.Context) error {
	roomId, err := getId(c)
	if err != nil {
		return err
	}

	uid := c.Param("uid")
	if roomId == 0 {
		if err := s.member.SetBlockadeAll(uid, false); err != nil {
			return err
		}
	} else if err := s.member.SetBlockade(uid, roomId, false); err != nil {
		return err
	}

	name, _ := s.member.GetUserName(uid)
	log.Debug("[blockade.go]removeBlockade ", zap.Int("RoomId", roomId), zap.String("uid", uid), zap.String("name", name))

	if roomId > 0 {
		// 發送封鎖訊息到kafka producer
		_, _ = s.message.SendDisplay(
			[]int32{int32(roomId)},
			scheme.NewRoot(),
			scheme.DisplayByUnBlock(name, 0, false),
		)
	}
	c.Status(http.StatusNoContent)
	return nil
}

// 取得roomId
func getId(c *gin.Context) (int, error) {
	id := 0  // roomId允許為0 (表示所有房間)
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
