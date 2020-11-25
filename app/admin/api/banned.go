package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
	"gitlab.com/jetfueltw/cpw/micro/log"

	"go.uber.org/zap"
)

// 設定禁言
func (s *httpServer) setBanned(c *gin.Context) error {

	id, err := getId(c)
	if err != nil {
		return err
	}

	// start
	exp := struct {
		Expired int `json:"expired"`
	}{}
	if err := c.BodyParser(&exp); err != nil {
		log.Error("BodyParser", zap.Error(err))
	} else {
		log.Debug("setBanned", zap.Int("expired", exp.Expired))
	}
	log.Debug("setBanned", zap.Int("roomid", id), zap.String("uid", c.Param("uid")))
	// end

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
	// start
	log.Debug("setBanned", zap.Int("params.RoomId", params.RoomId), zap.String("params.Uid", params.Uid), zap.Int("params.Expired", params.Expired))
	// end

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

	log.Debugf("setBanned response %d", http.StatusNoContent)
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
