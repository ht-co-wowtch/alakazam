package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"net/http"
)

type giveRedEnvelopeReq struct {
	RoomId int `json:"room_id" binding:"required"`

	// 單包金額 or 總金額 看Type種類決定
	Amount int `json:"amount" binding:"required"`

	Count int `json:"count" binding:"required"`

	// 紅包說明
	Message string `json:"message" binding:"required,max=20"`

	// 紅包種類 拼手氣 or 普通
	Type string `json:"type" binding:"required"`
}

func (s *httpServer) giveRedEnvelope(c *gin.Context) error {
	arg := new(giveRedEnvelopeReq)
	if err := c.ShouldBindJSON(arg); err != nil {
		return err
	}
	user, err := s.member.GetSession(c.GetString("uid"))
	if err != nil {
		return err
	}
	give := client.RedEnvelope{
		RoomId:    arg.RoomId,
		Message:   arg.Message,
		Type:      arg.Type,
		Amount:    arg.Amount,
		Count:     arg.Count,
		ExpireMin: 120,
	}

	reply, err := s.client.GiveRedEnvelope(give, c.GetString("token"))
	if err != nil {
		return err
	}

	msg := message.RedEnvelopeMessage{
		Messages: message.Messages{
			Rooms:   []int32{int32(arg.RoomId)},
			Mid:     int64(user.Id),
			Uid:     user.Uid,
			Name:    user.Name,
			Message: arg.Message,
		},
		RedEnvelopeId: reply.Uid,
		Token:         reply.Token,
		Expired:       reply.ExpireAt,
	}

	msgId, err := s.message.SendRedEnvelope(msg)

	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"id":         reply.Order,
		"token":      reply.Token,
		"message_id": msgId,
	})
	return nil
}

type takeRedEnvelope struct {
	Token string `json:"token" binding:"required"`
}

func (s *httpServer) takeRedEnvelope(c *gin.Context) error {
	arg := new(takeRedEnvelope)
	if err := c.ShouldBindJSON(arg); err != nil {
		return err
	}

	_, err := s.member.GetSession(c.GetString("uid"))
	if err != nil {
		return err
	}

	reply, err := s.client.TakeRedEnvelope(arg.Token, c.GetString("token"))
	if err != nil {
		return err
	}
	if reply.Name, err = s.member.GetUserName(reply.Uid); err != nil {
		return err
	}

	switch reply.Status {
	case client.TakeEnvelopeSuccess:
		reply.StatusMessage = "获得红包"
	case client.TakeEnvelopeReceived:
		reply.StatusMessage = "已经抢过了"
	case client.TakeEnvelopeGone:
		reply.StatusMessage = "手慢了，红包派完了"
	case client.TakeEnvelopeExpired:
		reply.StatusMessage = "红包已过期，不能抢"
	default:
		reply.StatusMessage = "不存在的红包"
	}
	c.JSON(http.StatusOK, reply)
	return nil
}

func (s *httpServer) getRedEnvelopeDetail(c *gin.Context) error {
	reply, err := s.client.GetRedEnvelopeDetail(c.Param("id"), c.GetString("token"))
	if err != nil {
		return err
	}
	name := make([]string, len(reply.Members)+1)
	name[0] = reply.Uid
	for i, v := range reply.Members {
		name[i+1] = v.Uid
	}

	names, err := s.member.GetUserNames(name)
	if err != nil {
		return err
	}

	reply.Name = names[reply.Uid]

	for i, v := range reply.Members {
		reply.Members[i].Name = names[v.Uid]
	}

	c.JSON(http.StatusOK, reply)
	return nil
}
