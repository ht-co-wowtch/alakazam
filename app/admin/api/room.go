package api

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strconv"
)

func (s *httpServer) CreateRoom(c *gin.Context) error {
	var params room.Status
	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}
	roomId, err := s.room.Create(params)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"id": roomId,
	})
	return nil
}

func (s *httpServer) UpdateRoom(c *gin.Context) error {
	var params room.Status
	var r models.Room
	rid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}
	if r, err = s.room.Get(rid); err != nil {
		return err
	}
	if err := s.room.Update(rid, params); err != nil {
		return err
	}

	if len(s.noticeUrl) > 0 {
		for _, u := range s.noticeUrl {
			b, _ := json.Marshal(gin.H{
				"id":         r.Id,
				"is_message": r.IsMessage,
				"is_bets":    r.IsBets,
				"limit": room.Limit{
					Day:     r.DayLimit,
					Deposit: r.DepositLimit,
					Dml:     r.DmlLimit,
				},
				"status":    r.Status,
				"create_at": r.CreateAt,
				"update_at": r.UpdateAt,
			})

			if _, err := http.Post(u, "application/json", bytes.NewReader(b)); err != nil {
				log.Error("notice", zap.Int("rid", rid), zap.Error(err))
			}
		}
	}

	c.Status(http.StatusNoContent)
	return nil
}

type Rid struct {
	Id int `form:"id" binding:"required"`
}

func (s *httpServer) GetRoom(c *gin.Context) error {
	rid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	if err := binding.Validator.ValidateStruct(&Rid{Id: rid}); err != nil {
		return err
	}
	r, err := s.room.Get(rid)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         r.Id,
		"is_message": r.IsMessage,
		"is_bets":    r.IsBets,
		"limit": room.Limit{
			Day:     r.DayLimit,
			Deposit: r.DepositLimit,
			Dml:     r.DmlLimit,
		},
		"status":    r.Status,
		"create_at": r.CreateAt,
		"update_at": r.UpdateAt,
	})
	return nil
}

func (s *httpServer) DeleteRoom(c *gin.Context) error {
	rid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	if err := binding.Validator.ValidateStruct(&Rid{Id: rid}); err != nil {
		return err
	}
	if err := s.room.Delete(rid); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (s *httpServer) AddManage(c *gin.Context) error {
	var params struct {
		RoomId int    `json:"room_id" binding:"required"`
		Uid    string `json:"uid" binding:"required,len=32"`
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}

	if err := s.member.SetManage(params.Uid, params.RoomId, true); err != nil {
		return err
	}

	keys, err := s.member.GetKeys(params.Uid)
	if err != nil {
		return err
	}
	m, _ := s.member.Get(params.Uid)
	r, _ := s.room.Get(int(params.RoomId))
	connect := room.NewPbConnect(m, r, "", 0)

	_, _ = s.message.SendPermission(keys, m, *connect.Permission)

	c.Status(http.StatusNoContent)
	return nil
}

func (s *httpServer) DeleteManage(c *gin.Context) error {
	rid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	params := struct {
		RoomId int    `json:"room_id" binding:"required"`
		Uid    string `form:"uid" binding:"required,len=32"`
	}{
		RoomId: rid,
		Uid:    c.Param("uid"),
	}
	if err := binding.Validator.ValidateStruct(&params); err != nil {
		return err
	}

	if err := s.member.SetManage(params.Uid, params.RoomId, false); err != nil {
		return err
	}
	keys, err := s.member.GetKeys(params.Uid)
	if err != nil {
		return err
	}
	m, _ := s.member.Get(params.Uid)
	r, _ := s.room.Get(int(params.RoomId))
	connect := room.NewPbConnect(m, r, "", 0)

	_, _ = s.message.SendPermission(keys, m, *connect.Permission)

	c.Status(http.StatusNoContent)
	return nil
}

func (s *httpServer) online(c *gin.Context) error {
	o, err := s.room.Online()
	if err == nil {
		c.JSON(http.StatusOK, o)
	}
	return err
}

func (s *httpServer) notice(c *gin.Context) error {
	var u struct {
		Name string `json:"key" binding:"required"`
		Url  string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&u); err != nil {
		return err
	}

	_, err := url.Parse(u.Url)
	if err != nil {
		return err
	}

	if _, ok := s.noticeUrl[u.Name]; !ok {
		s.noticeUrl[u.Name] = u.Url
	}

	return nil
}
