package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"net/http"
	"strconv"
	"time"
)

type redEnvelope struct {
	// 房間id
	RoomId int `json:"room_id" binding:"required"`

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
			RoomId:    strconv.Itoa(o.RoomId),
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

	msg := message.AdminRedEnvelopeMessage{
		AdminMessage: message.AdminMessage{
			Rooms:   []int32{int32(o.RoomId)},
			Message: o.Message,
		},
		RedEnvelopeId: result.Order,
		Token:         result.Token,
		Expired:       result.ExpireAt.Unix(),
	}
	var msgId int64
	if o.PublishAt.IsZero() {
		if msgId, err = s.message.SendRedEnvelopeForAdmin(msg); err != nil {
			return err
		}
	} else if result.PublishAt.Before(time.Now()) {
		return errors.ErrPublishAt
	} else if msgId, err = s.delayMessage.SendDelayRedEnvelopeForAdmin(msg, result.PublishAt); err != nil {
		return err
	}

	json := struct {
		Id          string    `json:"id"`
		MsgId       int64     `json:"message_id"`
		TotalAmount int       `json:"total_amount"`
		PublishAt   time.Time `json:"publish_at"`
		ExpireAt    time.Time `json:"expire_at"`
	}{
		Id:          result.Order,
		MsgId:       msgId,
		TotalAmount: result.TotalAmount,
		PublishAt:   result.PublishAt,
		ExpireAt:    result.ExpireAt,
	}

	c.JSON(http.StatusOK, json)
	return nil
}
