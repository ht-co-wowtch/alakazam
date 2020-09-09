package http

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
	"net/http"
	"strconv"
)

// 設定禁言
func (s *httpServer) setBanned(c *gin.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	params := struct {
		RoomId  int    `json:"room_id" binding:"required"`
		Uid     string `json:"uid" binding:"required,len=32"`
		Expired int    `json:"expired" binding:"required"`
	}{
		RoomId: id,
		Uid:    c.Param("uid"),
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}

	if err := s.isManage(params.RoomId, params.Uid); err != nil {
		return err
	}

	if err := s.member.SetBanned(params.Uid, params.RoomId, params.Expired, false); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

// 解除禁言
func (s *httpServer) removeBanned(c *gin.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	params := struct {
		RoomId int    `json:"room_id" binding:"required"`
		Uid    string `json:"uid" binding:"required,len=32"`
	}{
		RoomId: id,
		Uid:    c.Param("uid"),
	}
	if err := binding.Validator.ValidateStruct(&params); err != nil {
		return err
	}

	if err := s.isManage(params.RoomId, params.Uid); err != nil {
		return err
	}

	if err := s.member.RemoveBanned(params.Uid, params.RoomId); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (s *httpServer) isManage(rid int, uid string) error {
	member, err := s.member.GetByRoom(uid, rid)
	if err != nil {
		return err
	}

	if !member.IsManage {
		return errors.ErrForbidden
	}

	return nil
}

func (s *httpServer) user(c *gin.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	params := struct {
		RoomId int    `json:"room_id" binding:"required"`
		Uid    string `json:"uid" binding:"required,len=32"`
	}{
		RoomId: id,
		Uid:    c.Param("uid"),
	}
	if err := binding.Validator.ValidateStruct(&params); err != nil {
		return err
	}

	m, err := s.member.GetStatus(params.Uid, params.RoomId)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"uid":         m.Uid,
		"name":        m.Name,
		"type":        m.Type,
		"avatar":      scheme.ToAvatarName(m.Gender),
		"is_banned":   m.IsBanned,
		"is_blockade": m.IsBlockade,
		"is_manage":   m.IsManage,
	})

	return nil
}

func (s *httpServer) manageList(c *gin.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	ms, err := s.msg.room.GetManages(id)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, ms)
	return nil
}

func (s *httpServer) blockadeList(c *gin.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	ms, err := s.msg.room.GetBlockades(id)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, ms)
	return nil
}
