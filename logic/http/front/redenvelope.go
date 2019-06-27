package front

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gopkg.in/go-playground/validator.v8"
	"net/http"
)

type GiveRedEnvelope struct {
	logic.User

	// 單包金額 or 總金額 看Type種類決定
	Amount int `json:"amount" binding:"required"`

	Count int `json:"count" binding:"required"`

	// 紅包說明
	Message string `json:"message" binding:"required,max=20"`

	// 紅包種類 拼手氣 or 普通
	Type string `json:"type" binding:"required"`
}

// 發紅包
func (s *Server) giveRedEnvelope(c *gin.Context) error {
	arg := new(GiveRedEnvelope)
	if err := c.ShouldBindJSON(arg); err != nil {
		return err
	}
	if err := s.logic.Auth(&arg.User); err != nil {
		return err
	}

	give := client.RedEnvelope{
		RoomId:    arg.RoomId,
		Message:   arg.Message,
		Type:      arg.Type,
		Amount:    arg.Amount,
		Count:     arg.Count,
		ExpireMin: 60,
	}
	reply, err := s.client.GiveRedEnvelope(give, c.GetString("token"))
	if err != nil {
		return err
	}
	if err := s.logic.PushRedEnvelope(reply, arg.User); err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"id":    reply.Order,
		"token": reply.Token,
	})
	return nil
}

type TakeRedEnvelope struct {
	logic.User

	Token string `json:"token" binding:"required"`
}

func (s *Server) takeRedEnvelope(c *gin.Context) error {
	arg := new(TakeRedEnvelope)
	if err := c.ShouldBindJSON(arg); err != nil {
		return err
	}
	if err := s.logic.Auth(&arg.User); err != nil {
		return err
	}
	reply, err := s.client.TakeRedEnvelope(arg.Token, c.GetString("token"))
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, reply)
	return nil
}

func validatorLuckyMoney(c *gin.Context, arg *GiveRedEnvelope) error {
	var err error
	if err = c.ShouldBindJSON(arg); err != nil {
		v := err.(validator.ValidationErrors)
		for _, e := range v {
			switch e.Field {
			case "Amount":
				err = errors.DataError.Mes("红包金额最低0.01")
			case "Count":
				err = errors.DataError.Mes("红包最大数量是500")
			case "Message":
				err = errors.DataError.Mes("限制文字长度为1到20个字")
			default:
				err = errors.DataError
			}
		}
	}

	return err
}
