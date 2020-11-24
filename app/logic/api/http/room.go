package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
	"gitlab.com/jetfueltw/cpw/micro/log"
)

// Original 設定禁言
/*
func (s *httpServer) setBanned(c *gin.Context) error {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	params := struct {
		RoomId  int
		Uid     string
		Expired int `form:"expired"`
	}{
		RoomId: id,
		Uid:    c.Param("uid"),
	}

	if c.ShouldBind(&params) != nil {
		// default expired is 30
		params.Expired = int(30)
	}

	//below GetString("uid") come from authenticationHandler middleware at request very first
	if err := s.isManage(params.RoomId, c.GetString("uid")); err != nil {
		return err
	}

	isSystem := false
	if err := s.member.SetBanned(params.Uid, params.RoomId, params.Expired, isSystem); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}*/

// 透過Admin去設定禁言
func (s *httpServer) setBanned(c *gin.Context) error {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	params := struct {
		RoomId  int
		Uid     string
		Expired int `form:"expired"`
	}{
		RoomId: id,
		Uid:    c.Param("uid"),
	}

	if c.ShouldBind(&params) != nil {
		// default expired is 30
		params.Expired = int(30)
	}

	//below GetString("uid") come from authenticationHandler middleware at request very first
	if err := s.isManage(params.RoomId, c.GetString("uid")); err != nil {
		return err
	}

	/*
		return $this->api('POST', "banned/{$uid}/room/{$chatId}", [
		            'expired' => $expired,
		        ]);
	*/
	adminBannedUrl := fmt.Sprintf(s.adminBannedUrlf, params.Uid, params.RoomId)
	log.Debugf("adminBannedUrl: %s", adminBannedUrl)
	resp, err := http.Post(adminBannedUrl, "application/json", strings.NewReader("{expired:30}"))

	if err != nil {
		log.Errorf("After admin request banned Error: %s", err)
		return err
	}
	defer resp.Body.Close()

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

	if !member.Permission.IsManage {
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
		"type":        scheme.ToType(m.Type),
		"avatar":      scheme.ToAvatarName(m.Gender),
		"is_banned":   m.Banned(),
		"is_blockade": m.Blockade(),
		"is_manage":   m.Permission.IsManage,
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
