package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type messageReq struct {
	RoomId int `json:"room_id" binding:"required,max=10000"`

	// user push message
	Message string `json:"message" binding:"required,max=100"`

	ToUid string `json:"to_uid"`

	Uid string `json:"-"`

	Token string `json:"-"`
}

// 單一房間推送訊息
func (s *httpServer) pushRoom(c *gin.Context) error {
	var p messageReq
	if err := c.ShouldBindJSON(&p); err != nil {
		return err
	}

	p.Token = c.GetString("token")
	p.Uid = c.GetString("uid")

	var id int64
	var err error

	if p.ToUid != "" {
		id, err = s.msg.private(p)
	} else {
		id, err = s.msg.user(p)
	}

	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"id": id,
		})
	}

	return err
}

type privateReq struct {
	Keys []string `json:"keys" binding:"required"`

	// user push message
	Message string `json:"message" binding:"required,max=100"`
}

// 私密
func (s *httpServer) pushKey(c *gin.Context) error {
	var p privateReq
	if err := c.ShouldBindJSON(&p); err != nil {
		return err
	}

	user, err := s.member.GetSession(c.GetString("uid"))

	if err != nil {
		return err
	}

	id, err := s.msg.message.SendKey(p.Keys, p.Message, user)

	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"id": id,
		})
	}

	return err
}
