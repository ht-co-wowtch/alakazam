package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
	"net/http"
)

// 設定禁言
func (s *httpServer) setBanned(c *gin.Context) error {
	id, err := getId(c)
	if err != nil {
		return err
	}

	params := struct {
		RoomId  int    `json:"room_id"`
		Uid     string `json:"uid" binding:"required,len=32"`
		Expired int    `json:"expired" binding:"required"`
	}{
		RoomId: id,
		Uid:    c.Param("uid"),
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}

	if id == 0 {
		if err := s.member.SetBannedAll(params.Uid, params.Expired); err != nil {
			return err
		}
	} else if err := s.member.SetBanned(params.Uid, params.RoomId, params.Expired, false); err != nil {
		return err
	}

	m, _ := s.member.Get(params.Uid)
	_, _ = s.message.SendDisplay(
		[]int32{int32(params.RoomId)},
		scheme.NewRoot(),
		scheme.DisplayBySetBanned(m.Name, params.Expired, true),
	)

	c.Status(http.StatusNoContent)
	return nil
}

// 解除禁言
func (s *httpServer) removeBanned(c *gin.Context) error {
	id, err := getId(c)
	if err != nil {
		return err
	}

	params := struct {
		RoomId int    `json:"room_id"`
		Uid    string `json:"uid" binding:"required,len=32"`
	}{
		RoomId: id,
		Uid:    c.Param("uid"),
	}
	if err := binding.Validator.ValidateStruct(&params); err != nil {
		return err
	}

	if id == 0 {
		if err := s.member.RemoveBannedAll(params.Uid); err != nil {
			return err
		}
	} else if err := s.member.RemoveBanned(params.Uid, params.RoomId); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
