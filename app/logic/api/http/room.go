package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/ht-co/micro/log"
	"gitlab.com/ht-co/wowtch/live/alakazam/errors"
	"gitlab.com/ht-co/wowtch/live/alakazam/message/scheme"

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

// 禁言
// 透過聊天室Admin去設定禁言
func (s *httpServer) setBanned(c *gin.Context) error {
	// 會收到前端傳入的參數有
	//   id 指的是房間id(roomId)
	//   uid 指的是要被banned 的 user id
	var (
		err     error
		roomId  int
		uid     string
		expired = `{"expired":600}` //預設禁言10分鐘 (600/60)
	)

	log.Debug("[room.go]setBanned expired", zap.String("c.Param(id)", c.Param("id")), zap.String("c.Param(uid)", c.Param("uid")))

	roomId, err = strconv.Atoi(c.Param("id"))

	if err != nil {
		log.Error("[room.go]setBanned-strconv.Atoi(c.Param(id))", zap.Error(err))
		return errors.ErrNoRoom
	}

	uid = c.Param("uid")

	log.Debug("[room.go]setBanned", zap.String("expired", expired), zap.Int("roomId", roomId), zap.String("c.GetString(uid)", c.GetString("uid")))

	l := len(uid)

	//uid 必須是32個字元的字串
	if l < 32 || l > 32 {
		//return errors.New("[set banned] invalid user id")
		log.Error("[room.go]setBanned-len(uid)", zap.Error(err))
		return errors.ErrLogin
	}

	// c.GetString("uid") 是從一開始的middleware 就設定,表示驗證過的當前使用者
	if err := s.isManage(roomId, c.GetString("uid")); err != nil {
		log.Error("[room.go]setBanned-isManage", zap.Error(err))
		return errors.ErrForbidden
	}

	// 透過聊天室Admin去設定禁言
	// s.adminBannedUrlf 參考 logic/api/conf/conf.go
	// 格式為 "127.0.0.1:3112/banned/%%s/room/%%d"
	adminBannedUrl := fmt.Sprintf(s.adminBannedUrlf, uid, roomId)
	//log.Debug("setBanned", zap.String("admin url", adminBannedUrl))
	resp, err := http.Post(adminBannedUrl, "application/json", strings.NewReader(expired))

	if err != nil {
		log.Error("setBanned send admin Error", zap.Error(err))
		return err
	}

	defer resp.Body.Close()
	//if status code not in HTTP 200 serial
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Debug("setBanned admin response", zap.String("repStatus", resp.Status))
		return errors.ErrForbidden
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

	if !member.Permission.IsManage {
		return errors.ErrForbidden
	}

	return nil
}

// 取得用戶資料
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

// 房管名單
func (s *httpServer) manageList(c *gin.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	ms, err := s.msg.room.GetManages(id)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, ms)
	return nil
}

// 封鎖名單
func (s *httpServer) blockadeList(c *gin.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	ms, err := s.msg.room.GetBlockades(id)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, ms)
	return nil
}

// 聊天觀看者結構
type viewer struct {
	Uid     string `json:"uid"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Banned  bool   `json:"is_banned"`
	Type    string `json:"type"`
	Manager bool   `json:"is_manage"`
	Lv      int    `json:"lv"`
}

// 觀看名單
func (s *httpServer) onlineViewers(c *gin.Context) error {
	roomId, _ := strconv.Atoi(c.Param("id"))
	viewerList, err := s.msg.room.GetOnlineViewer()

	if err != nil {
		log.Errorf("取得聊天室在線名單錯誤, %+v", err.Error())
		return err
	}

	d := []viewer{}
	if viewers, ok := viewerList[int32(roomId)]; ok {
		var viewerUids []string
		for _, uid := range viewers {
			log.Infof("viewe uid:%+v", uid)
			viewerUids = append(viewerUids, uid)
		}
		members, err := s.member.BatchGetMembersByUid(viewerUids)
		if err != nil {
			return err
		}

		for _, m := range members {
			d = append(d, viewer{
				Uid:     m.Uid,
				Name:    m.Name,
				Type:    scheme.ToType(m.Type),
				Avatar:  scheme.ToAvatarName(m.Gender),
				Banned:  m.Banned(),
				Manager: m.Permission.IsManage,
				Lv:      m.Lv,
			})
		}
	}

	log.Infof("room %d 's viewers: ", roomId, d)
	c.JSON(http.StatusOK, d)

	return nil
}
