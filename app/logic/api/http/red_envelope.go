package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type giveRedEnvelopeReq struct {
	RoomId int `json:"room_id" binding:"required,max=10000"`

	// user push message
	Message string `json:"message" binding:"required,max=20"`

	// 單包金額 or 總金額 看Type種類決定
	Amount int `json:"amount" binding:"required"`

	// 數量
	Count int `json:"count" binding:"required"`

	// 紅包種類 拼手氣 or 普通
	Type string `json:"type" binding:"required"`

	uid string `json:"-"`

	token string `json:"-"`
}

func (s *httpServer) giveRedEnvelope(c *gin.Context) error {
	var arg giveRedEnvelopeReq
	if err := c.ShouldBindJSON(&arg); err != nil {
		return err
	}

	arg.token = c.GetString("token")
	arg.uid = c.GetString("uid")

	id, reply, err := s.msg.redEnvelope(arg)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"id":         reply.Order,
			"token":      reply.Token,
			"message_id": id,
		})
	}

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
