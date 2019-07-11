package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
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
func (s *Context) giveRedEnvelope(c *gin.Context) error {
	arg := new(GiveRedEnvelope)
	if err := c.ShouldBindJSON(arg); err != nil {
		return err
	}
	if err := s.logic.Auth(&arg.User); err != nil {
		return err
	}
	if !models.IsRedEnvelope(int(arg.Status)) {
		return errors.ErrLogin
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

func (s *Context) takeRedEnvelope(c *gin.Context) error {
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

	name, err := s.logic.GetUserName([]string{reply.Uid})
	if err != nil {
		return err
	}
	reply.Name = name[0]

	switch reply.Status {
	case "success":
		reply.StatusMessage = "获得红包"
	case "received":
		reply.StatusMessage = "已经抢过了"
	case "gone":
		reply.StatusMessage = "手慢了，红包派完了"
	default:
		reply.StatusMessage = "不存在的红包"
	}
	c.JSON(http.StatusOK, reply)
	return nil
}

func (s *Context) getRedEnvelope(c *gin.Context) error {
	reply, err := s.client.GetRedEnvelope(c.Param("id"), c.GetString("token"))
	if err != nil {
		return err
	}
	name := make([]string, len(reply.Members)+1)
	name[0] = reply.Uid
	for i, v := range reply.Members {
		name[i+1] = v.Uid
	}
	n, err := s.logic.GetUserName(name)
	if err != nil {
		return err
	}
	reply.Name = n[0]
	for i, v := range n[1:] {
		reply.Members[i].Name = v
	}
	c.JSON(http.StatusOK, reply)
	return nil
}
