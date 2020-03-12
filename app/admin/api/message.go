package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type pushRoomForm struct {
	// 要廣播的房間
	RoomId []int32 `json:"room_id" binding:"required"`

	// user push message
	Message string `json:"message" binding:"required,max=250"`

	// 訊息是否頂置
	Top bool `json:"top"`
}

// 多房間推送
func (s *httpServer) push(c *gin.Context) error {
	p := new(pushRoomForm)
	if err := c.ShouldBindJSON(p); err != nil {
		return err
	}

	msg := message.ProducerAdminMessage{
		Rooms:   p.RoomId,
		Message: p.Message,
		IsTop:   p.Top,
	}
	id, err := s.message.SendForAdmin(msg)
	if err != nil {
		return err
	}
	if p.Top {
		now := time.Now()
		m := message.Message{
			Id:        id,
			Uid:       member.RootUid,
			Type:      message.TopType,
			Name:      member.RootName,
			Message:   p.Message,
			Time:      now.Format("15:04:05"),
			Timestamp: now.Unix(),
		}
		if err := s.room.AddTopMessage(p.RoomId, m); err != nil {
			if err == errors.ErrNoRoom {
				return err
			}
			log.Error("add top message for admin push api", zap.Error(err), zap.Int32s("rids", p.RoomId))
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}

type betsReq struct {
	RoomId       []int32       `json:"room_id" binding:"required"`
	Uid          string        `json:"uid" binding:"required"`
	GameId       int           `json:"game_id" binding:"required"`
	PeriodNumber int           `json:"period_number" binding:"required"`
	Bets         []message.Bet `json:"bets" binding:"required"`
	Count        int           `json:"count" binding:"required"`
	TotalAmount  int           `json:"total_amount" binding:"required"`
}

// 跟投
func (s *httpServer) bets(c *gin.Context) error {
	req := new(betsReq)
	if err := c.ShouldBindJSON(req); err != nil {
		return err
	}

	m, err := s.member.Fetch(req.Uid)
	if err != nil {
		return err
	}
	msg := message.ProducerBetsMessage{
		Rooms:        req.RoomId,
		Mid:          int64(m.Id),
		Uid:          m.Uid,
		Name:         m.Name,
		Avatar:       m.Gender,
		GameId:       req.GameId,
		PeriodNumber: req.PeriodNumber,
		Bets:         req.Bets,
		Count:        req.Count,
		TotalAmount:  req.TotalAmount,
	}

	id, err := s.message.SendBets(msg)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}

// 取消置頂訊息
func (s *httpServer) deleteTopMessage(c *gin.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	msgId := int64(id)
	rid, msg, err := s.room.GetTopMessage(msgId)

	if err == nil {
		if err := s.message.CloseTop(msgId, rid); err != nil {
			return err
		}
		if err := s.room.DeleteTopMessage(rid, msgId); err != nil {
			return err
		}

		// TODO 因為後台UI介面的因素導致沒有處理`没有资料`情況，所以將`没有资料`情況視為正常
		// 參考 http://mantis.jetfuel.com.tw/view.php?id=3004
	} else if err == errors.ErrNoRows {
		rid = []int32{}
		msg = models.Message{
			Message: "没有资料",
		}
	} else {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      msgId,
		"message": msg.Message,
		"room_id": rid,
	})
	return nil
}

type redEnvelope struct {
	// 房間id
	RoomId int `json:"room_id" binding:"required"`

	// 紅包訊息
	Message string `json:"message" binding:"required,max=20"`

	// 紅包種類	Name string `json:"name"`
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

// 紅包
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

	msg := message.ProducerAdminRedEnvelopeMessage{
		ProducerAdminMessage: message.ProducerAdminMessage{
			Rooms:   []int32{int32(o.RoomId)},
			Name:    member.RootName,
			Message: o.Message,
		},
		RedEnvelopeId: result.Order,
		Token:         result.Token,
		Expired:       result.ExpireAt,
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
		Token       string    `json:"token"`
	}{
		Id:          result.Order,
		MsgId:       msgId,
		TotalAmount: result.TotalAmount,
		PublishAt:   result.PublishAt,
		ExpireAt:    result.ExpireAt,
		Token:       result.Token,
	}

	c.JSON(http.StatusOK, json)
	return nil
}

type giftReq struct {
	RoomId int32 `json:"room_id" binding:"required"`

	Message string `json:"message" binding:"required,max=250"`

	Animation string `json:"animation" binding:"required"`

	AnimationId int `json:"animation_id" binding:"required"`
}

// 發禮物
func (s *httpServer) gift(c *gin.Context) error {
	p := new(giftReq)
	if err := c.ShouldBindJSON(p); err != nil {
		return err
	}

	msg := message.ProducerGiftMessage{
		Room:        p.RoomId,
		Name:        member.System,
		Message:     p.Message,
		Animation:   p.Animation,
		AnimationId: p.AnimationId,
	}

	id, err := s.message.SendGift(msg)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}
