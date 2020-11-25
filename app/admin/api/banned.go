package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
	"gitlab.com/jetfueltw/cpw/micro/log"

	"go.uber.org/zap"
)

// New
func (s *httpServer) setBanned(c *gin.Context) error {
	var (
		err    error
		roomId int
		uid    string
		exp    = struct {
			Expired int `json:"expired"`
		}{}
	)
	roomId, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	uid = c.Param("uid")

	if err := c.ShouldBind(&exp); err != nil {
		log.Error("setBanned ShouldBind", zap.Error(err))
	} else {
		log.Debug("setBanned ShoudBind", zap.Int("expired", exp.Expired))
	}
	log.Debug("setBanned", zap.Int("roomid", roomId), zap.String("uid", uid))

	if roomId == 0 {
		if err := s.member.SetBannedAll(uid, exp.Expired); err != nil {
			return err
		}
	} else if err := s.member.SetBanned(uid, roomId, exp.Expired, false); err != nil {
		return err
	}

	name, _ := s.member.GetUserName(uid)
	if roomId > 0 {
		xid, err := s.message.SendDisplay(
			[]int32{int32(roomId)},
			scheme.NewRoot(),
			scheme.DisplayBySetBanned(name, exp.Expired, true),
		)
		log.Debug("setBanned SendDisplay", zap.Int("name", name))
		if err != nil {
			log.Error("SendDisplayErr", zap.Error(err))
		}
		log.Debug("SendDisplay", zap.Int64("xid", xid))
	}

	log.Debugf("setBanned response %d", http.StatusNoContent)
	c.Status(http.StatusNoContent)
	return nil
}

// 設定禁言
/*
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

	name, _ := s.member.GetUserName(params.Uid)
	if params.RoomId > 0 {
		_, _ = s.message.SendDisplay(
			[]int32{int32(params.RoomId)},
			scheme.NewRoot(),
			scheme.DisplayBySetBanned(name, params.Expired, true),
		)
	}
	c.Status(http.StatusNoContent)
	return nil
}
*/

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
