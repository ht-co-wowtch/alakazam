package api

import (
	"github.com/gin-gonic/gin"
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

func (s *httpServer) bets(c *gin.Context) error {
	req := new(betsReq)
	if err := c.ShouldBindJSON(req); err != nil {
		return err
	}

	member, err := s.member.Fetch(req.Uid)
	if err != nil {
		return err
	}
	msg := message.ProducerBetsMessage{
		Rooms:        req.RoomId,
		Mid:          int64(member.Id),
		Uid:          member.Uid,
		Name:         member.Name,
		Avatar:       member.Gender,
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
