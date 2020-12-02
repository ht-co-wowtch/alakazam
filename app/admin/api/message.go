package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
)

type customReq struct {
	RoomId  []int32        `json:"room_id" binding:"required"`
	Message scheme.Message `json:"message" binding:"required"`
	IsRaw   bool           `json:"is_raw"`
}

// 客制訊息內容(需依照格式填資料)
func (s *httpServer) custom(c *gin.Context) error {
	var p customReq
	if err := c.ShouldBindJSON(&p); err != nil {
		return err
	}

	id, err := s.message.SendMessage(p.RoomId, p.Message, p.IsRaw)
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

	if !p.IsTop && !p.IsBulletin {
		if id, err = s.message.SendAdmin(p.RoomId, p.Message); err != nil {
			return err
		}
	}
	if p.IsTop {
		if id, err = s.message.SendTop(p.RoomId, p.Message); err != nil {
			return err
		}
	}
	if p.IsBulletin {
		if id, err = s.message.SendSystem(p.RoomId, p.Message); err != nil {
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
	RoomId       []int32           `json:"room_id" binding:"required"`
	Uid          string            `json:"uid" binding:"required"`
	GameId       int               `json:"game_id" binding:"required"`
	GameName     string            `json:"game_name"`
	PeriodNumber int               `json:"period_number" binding:"required"`
	Bets         []scheme.BetOrder `json:"bets" binding:"required"`
	Count        int               `json:"count" binding:"required"`
	TotalAmount  int               `json:"total_amount" binding:"required"`
}

// 跟投
func (s *httpServer) bets(c *gin.Context) error {
	req := new(betsReq)
	if err := c.ShouldBindJSON(req); err != nil {
		return err
	}

	m, err := s.member.GetSession(req.Uid)
	if err != nil {
		return err
	}

	m.Uid = req.Uid
	user := scheme.NewUser(*m)

	bet := scheme.Bet{
		GameId:       req.GameId,
		GameName:     req.GameName,
		PeriodNumber: req.PeriodNumber,
		Count:        req.Count,
		TotalAmount:  req.TotalAmount,
		Orders:       req.Bets,
	}

	id, err := s.message.SendBets(req.RoomId, user, bet)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}

type betsPayReq struct {
	RoomId     int32         `json:"room_id" binding:"required"`
	Uid        string        `json:"uid" binding:"required"`
	GameName   string        `json:"game_name" binding:"required"`
	OpenReward OpenRewardReq `json:"open_chat_reward"`
}

type OpenRewardReq struct {
	Amount     float64 `json:"amount" binding:"required"`
	ButtonName string  `json:"button_name" binding:"required"`
}

// 投注中獎
func (s *httpServer) betsWin(c *gin.Context) error {
	req := new(betsPayReq)
	if err := c.ShouldBindJSON(req); err != nil {
		return err
	}

	m, err := s.member.GetSession(req.Uid)
	if err != nil {
		return err
	}

	m.Uid = req.Uid
	user := scheme.NewUser(*m)

	ws, err := s.member.GetWs(req.Uid)
	if err != nil {
		return err
	}

	var isSend bool
	keys := []string{}
	rid := strconv.Itoa(int(req.RoomId))

	for key, id := range ws {
		if id == rid {
			isSend = true
		}

		keys = append(keys, key)
	}

	if !isSend {
		return nil
	}

	id, err := s.message.SendBetsWin([]int32{req.RoomId}, user, req.GameName)
	if err != nil {
		return err
	}

	wid, err := s.message.SendBetsWinReward(keys, user, req.OpenReward.Amount, req.OpenReward.ButtonName)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        id,
		"reward_id": wid,
	})
	return nil
}

type giftReq struct {
	RoomId      int32   `json:"room_id" binding:"required"`
	Uid         string  `json:"uid" binding:"required"`
	Id          int     `json:"id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Combo       int     `json:"combo"`
	Amount      float64 `json:"amount" binding:"required"`
	TotalAmount float64 `json:"total_amount"`
	UserName    string  `json:"user_name"`
	UserAvatar  string  `json:"user_avatar"`
}

// 禮物
func (s *httpServer) gift(c *gin.Context) error {
	var req giftReq
	var user scheme.User
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	m, err := s.member.GetSession(req.Uid)
	//若member 為guest, GetSession就會拋出err
	if err != nil {
		return err
	}

	memberType := scheme.ToType(m.Type)
	log.Debug("Gift",
		zap.String("uid", req.Uid),
		zap.String("name", m.Name),
		zap.Int("(member)type", m.Type),
		zap.String("memberType", memberType),
		zap.Bool("isMessage", m.IsMessage),
		zap.Bool("isBlockade", m.IsBlockade))

	if req.UserName == "" {
		m.Uid = req.Uid
		user = scheme.NewUser(*m)
	} else {
		user = scheme.User{
			Name:   req.UserName,
			Uid:    req.Uid,
			Avatar: req.UserAvatar,
			Type:   memberType,
		}
	}

	id, err := s.message.SendGift(req.RoomId, user, scheme.Gift{
		Id:          req.Id,
		Name:        req.Name,
		Amount:      req.Amount,
		TotalAmount: req.TotalAmount,
		Combo: scheme.NullCombo{
			Count:      req.Combo,
			DurationMs: 3000,
		},
		Message: "送出" + req.Name,
	})
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}

type rewardReq struct {
	RoomId      int32   `json:"room_id" binding:"required"`
	Uid         string  `json:"uid" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	TotalAmount float64 `json:"total_amount"`
	UserName    string  `json:"user_name"`
	UserAvatar  string  `json:"user_avatar"`
}

// 打賞
func (s *httpServer) reward(c *gin.Context) error {
	var req rewardReq
	var user scheme.User
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	if req.UserName == "" {
		m, err := s.member.GetSession(req.Uid)
		if err != nil {
			return err
		}

		m.Uid = req.Uid
		user = scheme.NewUser(*m)
	} else {
		user = scheme.User{
			Name:   req.UserName,
			Uid:    req.Uid,
			Avatar: req.UserAvatar,
			Type:   "player",
		}
	}

	id, err := s.message.SendReward(req.RoomId, user, req.Amount, req.TotalAmount)
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

	u := scheme.NewRoot()
	redEnvelope := scheme.RedEnvelope{
		Id:      result.Order,
		Token:   result.Token,
		Expired: result.ExpireAt.Format(time.RFC3339),
	}

	var msgId int64
	rid := []int32{int32(o.RoomId)}
	if o.PublishAt.IsZero() {
		if msgId, err = s.message.SendRedEnvelope(rid, o.Message, u, redEnvelope); err != nil {
			return err
		}
	} else if result.PublishAt.Before(time.Now()) {
		return errors.ErrPublishAt
	} else if msgId, err = s.delayMessage.SendDelayRedEnvelopeForAdmin(rid, o.Message, u, redEnvelope, result.PublishAt); err != nil {
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

type followReq struct {
	RoomId int32  `json:"room_id" binding:"required"`
	Uid    string `json:"uid" binding:"required"`
	Total  int    `json:"total" binding:"required"`
}

func (s *httpServer) follow(c *gin.Context) error {
	var req followReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	m, err := s.member.GetSession(req.Uid)
	if err != nil {
		return err
	}

	id, err := s.message.SendFollow(req.RoomId, scheme.NewUser(*m), req.Total)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}
