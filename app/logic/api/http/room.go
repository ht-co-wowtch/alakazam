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
