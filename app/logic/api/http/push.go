package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
)

type messageReq struct {
	RoomId int `json:"room_id" binding:"required,max=10000"`

	// user push message
	Message string `json:"message" binding:"required,max=100"`
}

// 單一房間推送訊息
func (s *httpServer) pushRoom(c *gin.Context) error {
	p := new(messageReq)
	if err := c.ShouldBindJSON(p); err != nil {
		return err
	}
	user, err := s.member.GetMessageSession(c.GetString("uid"))
	if err != nil {
		return err
	}
	if err := s.room.IsMessage(p.RoomId, user.Uid); err != nil {
		return err
	}

	msg := message.Messages{
		Rooms:   []int32{int32(p.RoomId)},
		Mid:     int64(user.Id),
		Uid:     user.Uid,
		Name:    user.Name,
		Message: p.Message,
	}

	id, err := s.message.Send(msg)
	switch err {
	case errors.ErrRateSameMsg:
		isBlockade, err := s.member.SetBannedForSystem(user.Uid, 10*60)
		if err != nil {
			log.Error("set banned for rate same message", zap.Error(err), zap.String("uid", user.Uid))
		}
		if isBlockade {
			keys, err := s.member.Kick(user.Uid)
			if err != nil {
				log.Error("kick member for push room", zap.Error(err), zap.String("uid", user.Uid))
			}
			if len(keys) > 0 {
				err = s.message.Kick(message.KickMessage{
					Message: "你被踢出房间，因为自动禁言达五次",
					Keys:    keys,
				})
				if err == nil {
					log.Error("kick member set message for push room", zap.Error(err))
				}
			}
		}
	case nil:
		c.JSON(http.StatusOK, gin.H{
			"id": id,
		})
	}
	return err
}
