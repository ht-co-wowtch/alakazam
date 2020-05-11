package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	"net/http"
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
	rid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}
	if _, err := s.room.Get(rid); err != nil {
		return err
	}
	if err := s.room.Update(rid, params); err != nil {
		return err
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

func (s *httpServer) online(c *gin.Context) error {
	o, err := s.room.Online()
	if err == nil {
		c.JSON(http.StatusOK, o)
	}
	return err
}
