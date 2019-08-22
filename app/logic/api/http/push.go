package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
)

type PushRoom struct {
	member.User

	// user push message
	Message string `json:"message" binding:"required,max=100"`
}

// 單一房間推送訊息
func (s *httpServer) pushRoom(c *gin.Context) error {
	p := new(PushRoom)
	if err := c.ShouldBindJSON(p); err != nil {
		return err
	}
	if err := s.member.Auth(&p.User); err != nil {
		return err
	}
	if err := s.room.Auth(&p.User); err != nil {
		return err
	}
	if err := s.room.IsMessage(p.H.Room, p.RoomStatus, p.Uid, c.GetString("token")); err != nil {
		return err
	}

	msg := message.Messages{
		Rooms:   []int32{int32(p.H.Room)},
		Mid:     int64(p.H.Mid),
		Uid:     p.Uid,
		Name:    p.H.Name,
		Message: p.Message,
	}
	id, err := s.message.Send(msg)
	if err != nil {
		if err == errors.ErrRateSameMsg {
			isBlockade, err := s.member.SetBannedForSystem(p.Uid, 10*60)
			if err != nil {
				log.Error("set banned for rate same message", zap.Error(err), zap.String("uid", p.Uid))
			}
			if isBlockade {
				keys, err := s.member.Kick(p.User.Uid)
				if err != nil {
					log.Error("kick member for push room", zap.Error(err), zap.String("uid", p.User.Uid))
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
		}
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}
