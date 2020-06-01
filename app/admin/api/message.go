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

// 客制訊息內容(需依照格式填資料)
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

	// 訊息是否為頂置
	IsTop bool `json:"is_top"`

	// 訊息是否為公告
	IsBulletin bool `json:"is_bulletin"`
}

// 訊息
func (s *httpServer) push(c *gin.Context) error {
	p := new(pushRoomReq)
	if err := c.ShouldBindJSON(p); err != nil {
		return err
	}

	var id int64
	var err error
	u := message.NewRoot()
	msg := message.ProducerMessage{
		Rooms: p.RoomId,
		User:  u,
	}

	if !p.IsTop && !p.IsBulletin {
		msg.Display = message.DisplayByAdmin(u, p.Message)
		msg := message.ProducerMessage{
			Rooms:   p.RoomId,
			Display: message.DisplayByAdmin(u, p.Message),
			User:    u,
		}

		if id, err = s.message.SendForAdmin(msg); err != nil {
			return err
		}
	}

	if p.IsTop {
		msg := message.ProducerMessage{
			Rooms:   p.RoomId,
			Display: message.DisplayBySystem(p.Message),
			User:    u,
		}

		if id, err = s.message.SendTop(msg); err != nil {
			return err
		}
	}
	if p.IsBulletin {
		msg := message.ProducerMessage{
			Rooms:   p.RoomId,
			Display: message.DisplayBySystem(p.Message),
			User:    u,
		}

		if id, err = s.message.Send(msg); err != nil {
			return err
		}
	}

	var ts []int
	if p.IsTop {
		ts = append(ts, models.TOP_MESSAGE)
	}
	if p.IsBulletin {
		ts = append(ts, models.BULLETIN_MESSAGE)
	}

	if len(ts) > 0 {
		if err := s.room.AddTopMessage(p.RoomId, id, p.Message, ts); err != nil {
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
	GameName     string             `json:"game_name"`
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
		Display: message.DisplayByBets(user, req.GameName, req.TotalAmount),
	}

	bet := message.Bet{
		GameId:       req.GameId,
		GameName:     req.GameName,
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

type betsPayReq struct {
	RoomId   int32  `json:"room_id" binding:"required"`
	Uid      string `json:"uid" binding:"required"`
	GameName string `json:"game_name" binding:"required"`
}

// 注單派彩
func (s *httpServer) betsPay(c *gin.Context) error {
	req := new(betsPayReq)
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
		Rooms:   []int32{req.RoomId},
		User:    user,
		Display: message.DisplayByBetsPay(user, req.GameName),
	}

	id, err := s.message.Send(msg)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}

type giftReq struct {
	RoomId      int32   `json:"room_id" binding:"required"`
	Uid         string  `json:"uid" binding:"required"`
	Id          int     `json:"id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Combo       int     `json:"combo" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	TotalAmount float64 `json:"total_amount" binding:"required"`
	UserName    string  `json:"user_name"`
	UserAvatar  string  `json:"user_avatar"`
}

// 禮物
func (s *httpServer) gift(c *gin.Context) error {
	var req giftReq
	var user message.User
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	if req.Name == "" {
		m, err := s.member.Fetch(req.Uid)
		if err != nil {
			return err
		}

		user = message.User{
			Name:   m.Name,
			Uid:    req.Uid,
			Avatar: message.ToAvatarName(m.Gender),
		}
	} else {
		user = message.User{
			Name:   req.UserName,
			Uid:    req.Uid,
			Avatar: req.UserAvatar,
		}
	}

	id, err := s.message.SendGift(req.RoomId, user, message.Gift{
		Id:          req.Id,
		Name:        req.Name,
		Amount:      req.Amount,
		TotalAmount: req.TotalAmount,
		Combo:       req.Combo,
	})
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}

type redEnvelopeReq struct {
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
	var o redEnvelopeReq
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
		IsSave:  true,
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

var delMessageType = map[string]int{
	"top":      models.TOP_MESSAGE,
	"bulletin": models.BULLETIN_MESSAGE,
}

// 取消置頂訊息
func (s *httpServer) deleteTopMessage(c *gin.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	var rid []int32
	var msg string
	msgId := int64(id)
	t, ok := delMessageType[c.Query("type")]

	if ok {
		var topMsg models.Message
		rid, topMsg, err = s.room.GetTopMessage(msgId, t)
		if err != nil {
			goto Err
		}
		if t == models.TOP_MESSAGE {
			if err := s.message.CloseTop(msgId, rid); err != nil {
				return err
			}
		}
		if err := s.room.DeleteTopMessage(rid, msgId, t); err != nil {
			return err
		}

		msg = topMsg.Message
	} else {
		err = errors.ErrNoRows
	}

	// TODO 因為後台UI介面的因素導致沒有處理`没有资料`情況，所以將`没有资料`情況視為正常
	// 參考 http://mantis.jetfuel.com.tw/view.php?id=3004
Err:
	if err != nil {
		if err == errors.ErrNoRows {
			rid = []int32{}
			msg = "没有资料"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      msgId,
		"message": msg,
		"room_id": rid,
	})
	return nil
}
