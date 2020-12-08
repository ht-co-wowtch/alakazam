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

func (s *httpServer) setBanned(c *gin.Context) error {
	var (
		err    error
		roomId int
		uid    string
		exp    = struct {
			Expired int `json:"expired"`
		}{}
	)

	// 當request為 /banned/:uid , c.Param("id")為空字串,因此strconv.Atoi會丟error,這裡忽略掉error,因為底下roomId允許為 0
	// 當request為 /banned/:uid/room/:id , roomId 會得到一個roomid
	roomId, _ = strconv.Atoi(c.Param("id"))

	uid = c.Param("uid")

	if err = c.ShouldBind(&exp); err != nil {
		log.Error("setBanned Error", zap.Error(err))
	} else {
		log.Debug("setBanned expired", zap.Int("expired", exp.Expired))
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
		log.Debug("setBanned SendDisplay", zap.String("name", name))
		if err != nil {
			log.Error("setBanned SendDisplay Err", zap.Error(err))
		}
		log.Debug("SendDisplay", zap.Int64("msg_id", xid))
	}

	c.Status(http.StatusNoContent)
	return nil
}

// old 設定禁言
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

	log.Debug("setBanned", zap.Int("RoomId", params.RoomId), zap.String("Uid", params.Uid), zap.Int("Expired", params.Expired))
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
	roomId, err := getId(c)
	if err != nil {
		return err
	}

	log.Debug("removeBanned ", zap.Int("getId(c)", roomId))

	params := struct {
		RoomId int    `json:"room_id"`
		Uid    string `json:"uid" binding:"required,len=32"`
	}{
		RoomId: roomId,
		Uid:    c.Param("uid"),
	}

	log.Debug("removeBanned ", zap.Int("RoomId", params.RoomId), zap.String("uid", params.Uid))

	if err := binding.Validator.ValidateStruct(&params); err != nil {
		log.Error("removeBanned Validate", zap.Error(err))
		return err
	}

	if roomId == 0 {
		if err := s.member.RemoveBannedAll(params.Uid); err != nil {
			log.Error("removeBanned RemoveBannedAll", zap.Error(err))
			return err
		}
	} else if err := s.member.RemoveBanned(params.Uid, params.RoomId); err != nil {
		log.Error("removeBanned RemoveBanned", zap.Error(err))
		return err
	}
	/**************************/
	/*
		name, _ := s.member.GetUserName(params.Uid)
		log.Debug("removeBanned ", zap.String("name", name))
		if params.RoomId > 0 {
			_, _ = s.message.SendDisplay(
				[]int32{int32(params.RoomId)},
				scheme.NewRoot(),
				scheme.DisplayBySetBanned(name, 0, false),
			)
		}
	*/
	/**************************/
	c.Status(http.StatusNoContent)
	return nil
}
