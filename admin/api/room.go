package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	"net/http"
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
		"room_id": roomId,
	})
	return nil
}

func (s *httpServer) UpdateRoom(c *gin.Context) error {
	var params room.Status
	params.Id = c.Param("id")
	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}
	if err := s.room.Update(params); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

type Rid struct {
	Id string `form:"id" binding:"required"`
}

func (s *httpServer) GetRoom(c *gin.Context) error {
	rid := Rid{
		Id: c.Param("id"),
	}
	if err := binding.Validator.ValidateStruct(&rid.Id); err != nil {
		return err
	}
	r, err := s.room.Get(rid.Id)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         r.Id,
		"is_message": r.IsMessage,
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
	room := Rid{
		Id: c.Param("id"),
	}
	if err := binding.Validator.ValidateStruct(&room.Id); err != nil {
		return err
	}
	err := s.room.Delete(room.Id)
	if err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
