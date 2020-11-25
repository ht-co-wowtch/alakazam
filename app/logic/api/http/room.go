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
	"go.uber.org/zap"
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
	// 會收到前端傳入的參數有
	//   id 指的是房間id
	//   uid 指的是要被banned 的 user id
	var (
		err     error
		roomId  int
		uid     string
		expired = `{"expired":600}` //預設禁言 30秒
	)

	roomId, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	uid = c.Param("uid")
	l := len(uid)
	//uid 必須是32個字元的字串
	if l < 32 || l > 32 {
		return errors.New("[set banned] invalid user id")
	}

	// c.GetString("uid") 是從一開始的middleware 就設定,表示驗證過的當前使用者
	if err := s.isManage(roomId, c.GetString("uid")); err != nil {
		return err
	}
	// s.adminBannedUrlf 參考 logic/api/conf/conf.go
	// 格式為 "127.0.0.1:3112/banned/%%s/room/%%d"
	adminBannedUrl := fmt.Sprintf(s.adminBannedUrlf, uid, toomId)
	log.Debug("DEBUG adminBannedUrl", zap.String("adminBannedUrl", adminBannedUrl), zap.String("RoomId/id", c.Param("id")), zap.String("uid", c.Param("uid")), zap.String("expired", expired))

	resp, err := http.Post(adminBannedUrl, "application/json", strings.NewReader(expired))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	//if status code not in HTTP 200 serial
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.New("admin service banned error: " + resp.Status)
	}

	log.Debugf("recv admin setBanned status: %s, code: %d", resp.Status, resp.StatusCode)

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
