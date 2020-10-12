package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
	"net/http"
	"strconv"
)

func (s *httpServer) profile(c *gin.Context) error {
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
		"type":        scheme.ToType(m.Type),
		"avatar":      scheme.ToAvatarName(m.Gender),
		"is_banned":   m.Banned(),
		"is_blockade": m.Blockade(),
		"is_manage":   m.Permission.IsManage,
	})

	return nil
}

func (s *httpServer) renew(c *gin.Context) error {
	user, err := s.nidoran.Auth(c.Param("token"))
	if err != nil {
		return err
	}
	if err := s.member.Update(user.Uid, user.Name, user.Gender); err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"uid": user.Uid,
	})
	return nil
}
