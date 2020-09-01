package http

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"net/http"
	"strconv"
)

type bannedReq struct {
	RoomId  int    `json:"room_id" binding:"required"`
	Uid     string `json:"uid" binding:"required,len=32"`
	Expired int    `json:"expired" binding:"required"`
}

// 設定禁言
func (s *httpServer) setBanned(c *gin.Context) error {
	var req bannedReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	if err := s.isManage(req.RoomId, req.Uid); err != nil {
		return err
	}

	if err := s.member.SetBanned(req.Uid, req.RoomId, req.Expired, false); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

// 解除禁言
func (s *httpServer) removeBanned(c *gin.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	req := struct {
		RoomId int    `json:"room_id" binding:"required"`
		Uid    string `json:"uid" binding:"required,len=32"`
	}{
		RoomId: id,
		Uid:    c.Param("uid"),
	}
	if err := binding.Validator.ValidateStruct(&req); err != nil {
		return err
	}

	if err := s.isManage(req.RoomId, req.Uid); err != nil {
		return err
	}

	if err := s.member.RemoveBanned(req.Uid, req.RoomId); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

// 封鎖
func (s *httpServer) setBlockade(c *gin.Context) error {
	//if err := s.isManage(c.GetString("uid")); err != nil {
	//	return err
	//}
	//
	//uid := c.Param("uid")
	//ok, err := s.member.SetBlockade(uid)
	//if err != nil {
	//	return err
	//}
	//if !ok {
	//	return errors.ErrNoRows
	//}
	//
	//keys, err := s.member.Kick(uid)
	//if err != nil {
	//	return err
	//}
	//
	//var msg string
	//if len(keys) == 0 {
	//	msg = "封锁成功"
	//} else {
	//	err = s.msg.message.Kick("你被踢出房间，因为被封锁", keys)
	//
	//	if err != nil {
	//		log.Error("kick member message for set blockade", zap.Error(err), zap.String("uid", uid))
	//		msg = "封锁成功，但执行聊天室踢人失败"
	//	} else {
	//		msg = fmt.Sprintf("封锁成功，將執行中断该用户所在的%d个连线", len(keys))
	//	}
	//}
	//
	//c.JSON(http.StatusOK, gin.H{
	//	"msg": msg,
	//})
	return nil
}

// 解除封鎖
func (s *httpServer) removeBlockade(c *gin.Context) error {
	//if err := s.isManage(c.GetString("uid")); err != nil {
	//	return err
	//}
	//
	//ok, err := s.member.RemoveBlockade(c.Param("uid"))
	//if err != nil {
	//	return err
	//}
	//if !ok {
	//	return errors.ErrNoRows
	//}
	//c.Status(http.StatusNoContent)
	return nil
}

func (s *httpServer) isManage(rid int, uid string) error {
	m, _, err := s.msg.room.GetMessageSession(uid, rid)
	if err != nil {
		return err
	}

	switch m.Type {
	case models.STREAMER, models.MANAGE:
	default:
		return errors.ErrForbidden
	}

	return nil
}
