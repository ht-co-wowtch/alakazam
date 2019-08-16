package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"net/http"
	"time"
)

type redEnvelope struct {
	// 房間id
	RoomId string `json:"room_id" binding:"required"`

	// 紅包訊息
	Message string `json:"message" binding:"required,max=20"`

	// 紅包種類
	Type string `json:"type" binding:"required"`

	// 紅包金額 看種類決定
	Amount int `json:"amount" binding:"required,min=1"`

	// 紅包數量
	Count int `json:"count" binding:"required,min=1,max=100"`

	// 紅包多久過期(分鐘)
	ExpireMin int `json:"expire_min" binding:"required,min=1,max=120"`

	// 什麼時候發佈
	PublishAt time.Time `json:"publish_at"`
}

func (s *httpServer) giveRedEnvelope(c *gin.Context) error {
	var o redEnvelope
	if err := c.ShouldBindJSON(&o); err != nil {
		return err
	}
	result, err := s.nidoran.GiveRedEnvelopeForAdmin(client.RedEnvelopeAdmin{
		RedEnvelope: client.RedEnvelope{
			RoomId:    o.RoomId,
			Message:   o.Message,
			Type:      o.Type,
			Amount:    o.Amount,
			Count:     o.Count,
			ExpireMin: o.ExpireMin,
		},
		PublishAt: o.PublishAt,
	})
	if err != nil {
		return err
	}
	msg := message.Message{
		Name:    "管理员",
		Message: o.Message,
		Time:    time.Now().Format("15:04:05"),
	}
	redEnvelope := message.RedEnvelope{
		Id:      result.Uid,
		Token:   result.Token,
		Expired: result.ExpireAt.Unix(),
	}
	if o.PublishAt.IsZero() {
		if err := s.message.SendRedEnvelope(o.RoomId, msg, redEnvelope); err != nil {
			return err
		}
	} else if result.PublishAt.After(time.Now()) {
		if err := s.message.SendDelayRedEnvelope(o.RoomId, msg, redEnvelope, result.PublishAt); err != nil {
			return err
		}
	}
	c.JSON(http.StatusOK, result)
	return nil
}
