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

	// 當request為 /banned/:uid , c.Param("id")為空字串,因此strconv.Atoi會丟error,這裡忽略掉error,因為底下roomId允許為0 (表示所有房間)
	// 當request為 /banned/:uid/room/:id , roomId 會得到一個roomid
	roomId, _ = strconv.Atoi(c.Param("id"))

	uid = c.Param("uid")

	if err = c.ShouldBind(&exp); err != nil {
		log.Error("[banned.go]setBanned ShouldBind Error", zap.Error(err))
	} else {
		log.Debug("[banned.go]setBanned expired", zap.Int("expired", exp.Expired))
	}
	log.Debug("[banned.go]setBanned", zap.Int("roomid", roomId), zap.String("uid", uid))

	if roomId == 0 {
		if err := s.member.SetBannedAll(uid, exp.Expired); err != nil {
			return err
		}
	} else if err := s.member.SetBanned(uid, roomId, exp.Expired, false); err != nil {
		return err
	}

	name, _ := s.member.GetUserName(uid)
	if roomId > 0 {
		_, err := s.message.SendDisplay(
			[]int32{int32(roomId)},
			scheme.NewRoot(),
			scheme.DisplayBySetBanned(name, exp.Expired, true),
		)
		if err != nil {
			log.Error("[banned.go]SendDisplay Err", zap.Error(err))
		}
	}

	c.Status(http.StatusNoContent)
	return nil
}

// 解除禁言
func (s *httpServer) removeBanned(c *gin.Context) error {
	roomId, err := getId(c)
	if err != nil {
		return err
	}

	params := struct {
		RoomId int    `json:"room_id"`
		Uid    string `json:"uid" binding:"required,len=32"`
	}{
		RoomId: roomId,
		Uid:    c.Param("uid"),
	}

	log.Debug("[banned.go]removeBanned", zap.Int("RoomId", params.RoomId), zap.String("uid", params.Uid))

	if err := binding.Validator.ValidateStruct(&params); err != nil {
		log.Error("[banned.go]removeBanned Validate", zap.Error(err))
		return err
	}

	if roomId == 0 { // roomId表示所有房間
		if err := s.member.RemoveBannedAll(params.Uid); err != nil {
			log.Error("[banned.go]RemoveBannedAll", zap.Error(err))
			return err
		}
	} else if err := s.member.RemoveBanned(params.Uid, params.RoomId); err != nil {
		log.Error("[banned.go]removeBanned", zap.Error(err))
		return err
	}
	/**************************/
	/*
		name, _ := s.member.GetUserName(params.Uid)
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
