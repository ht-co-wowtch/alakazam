package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type systemReq struct {
	Messages []message.RawMessage `json:"messages" binding:"required"`
	IsRaw    bool                 `json:"is_raw"`
}

func (s *httpServer) system(c *gin.Context) error {
	var p systemReq
	var id int64
	var err error
	if err = c.ShouldBindJSON(&p); err != nil {
		return err
	}

	if len(p.Messages) > 1 {
		id, err = s.message.SendRaws(p.Messages, p.IsRaw)
	} else {
		id, err = s.message.SendRaw(p.Messages[0].RoomId, []byte(p.Messages[0].Body), p.IsRaw)
	}

	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}

type pushRoomReq struct {
	// 要廣播的房間
	RoomId []int32 `json:"room_id" binding:"required"`

	// user push message
	Message string `json:"message" binding:"required,max=250"`

	// 訊息是否頂置
	Top bool `json:"top"`
}

func (s *httpServer) push(c *gin.Context) error {
	p := new(pushRoomReq)
	if err := c.ShouldBindJSON(p); err != nil {
		return err
	}

	u := message.NewRoot()
	msg := message.ProducerMessage{
		Rooms:   p.RoomId,
		Display: message.DisplayByAdmin(u, p.Message),
		User:    u,
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
			Uid:       u.Uid,
			Type:      message.TopType,
			Name:      u.Name,
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
	RoomId       []int32            `json:"room_id" binding:"required"`
	Uid          string             `json:"uid" binding:"required"`
	GameId       int                `json:"game_id" binding:"required"`
	PeriodNumber int                `json:"period_number" binding:"required"`
	Bets         []message.BetOrder `json:"bets" binding:"required"`
	Count        int                `json:"count" binding:"required"`
	TotalAmount  int                `json:"total_amount" binding:"required"`
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

	user := message.User{
		Name:   m.Name,
		Uid:    req.Uid,
		Avatar: message.ToAvatarName(m.Gender),
	}

	msg := message.ProducerMessage{
		Rooms:   req.RoomId,
		User:    user,
		Display: message.DisplayByBets(user, "六合彩", req.TotalAmount),
	}

	bet := message.Bet{
		GameId:       req.GameId,
		PeriodNumber: req.PeriodNumber,
		Count:        req.Count,
		TotalAmount:  req.TotalAmount,
		Orders:       req.Bets,
	}

	id, err := s.message.SendBets(msg, bet)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
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

	u := message.NewRoot()

	msg := message.ProducerMessage{
		Rooms:   []int32{int32(o.RoomId)},
		Display: message.DisplayByAdmin(u, o.Message),
		User:    u,
	}

	redEnvelope := message.RedEnvelope{
		Id:      result.Order,
		Token:   result.Token,
		Expired: result.ExpireAt.Format(time.RFC3339),
	}

	var msgId int64
	if o.PublishAt.IsZero() {
		if msgId, err = s.message.SendRedEnvelope(msg, redEnvelope); err != nil {
			return err
		}
	} else if result.PublishAt.Before(time.Now()) {
		return errors.ErrPublishAt
	} else if msgId, err = s.delayMessage.SendDelayRedEnvelopeForAdmin(msg, redEnvelope, result.PublishAt); err != nil {
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
