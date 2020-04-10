package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
)

func (s *httpServer) giveRedEnvelope(c *gin.Context) error {
	var arg member.RedEnvelope
	if err := c.ShouldBindJSON(&arg); err != nil {
		return err
	}

	user, reply, err := s.member.GiveRedEnvelope(c.GetString("uid"), c.GetString("token"), arg)
	if err != nil {
		return err
	}

	msg := message.ProducerRedEnvelopeMessage{
		ProducerMessage: message.ProducerMessage{
			Rooms: []int32{int32(arg.RoomId)},
			User:  toUserMessage(user),
			Display: message.Display{
				Message: message.DisplayText{
					Text: arg.Message,
				},
			},
		},
		RedEnvelopeId: reply.Order,
		Token:         reply.Token,
		Expired:       reply.ExpireAt,
	}

	msgId, err := s.message.SendRedEnvelope(msg)

	if err != nil {
		log.Error("send red envelope message error",
			zap.Error(err),
			zap.String("uid", user.Uid),
			zap.String("order", reply.Order),
		)
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

	reply, err := s.member.TakeRedEnvelope(c.GetString("uid"), c.GetString("token"), arg.Token)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, reply)
	return nil
}

func (s *httpServer) getRedEnvelopeDetail(c *gin.Context) error {
	detail, err := s.member.GetRedEnvelopeDetail(c.Param("id"), c.GetString("token"))
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, detail)
	return nil
}
