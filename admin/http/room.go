package http

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"net/http"
)

func (s *Server) CreateRoom(c *gin.Context) error {
	var params logic.Room
	if err := bindRoom(c, &params); err != nil {
		return err
	}
	roomId, err := s.logic.CreateRoom(params)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"room_id": roomId,
	})
	return nil
}

func (s *Server) UpdateRoom(c *gin.Context) error {
	var params logic.Room
	params.Id = c.Param("id")
	if err := bindRoom(c, &params); err != nil {
		return err
	}
	if err := s.logic.UpdateRoom(params); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

type Rid struct {
	Id string `form:"id" binding:"required"`
}

func (s *Server) GetRoom(c *gin.Context) error {
	room := Rid{
		Id: c.Param("id"),
	}
	if err := binding.Validator.ValidateStruct(&room.Id); err != nil {
		return err
	}
	r, err := s.logic.GetRoom(room.Id)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         r.Id,
		"is_message": r.IsMessage,
		"limit": logic.Limit{
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

func (s *Server) DeleteRoom(c *gin.Context) error {
	room := Rid{
		Id: c.Param("id"),
	}
	if err := binding.Validator.ValidateStruct(&room.Id); err != nil {
		return err
	}
	err := s.logic.DeleteRoom(room.Id)
	if err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func bindRoom(c *gin.Context, params *logic.Room) error {
	if err := c.ShouldBindJSON(params); err != nil {
		return err
	}
	if params.Limit.Day != 0 {
		if (params.Limit.Deposit < 0 && params.Limit.Dml < 0) || (params.Limit.Deposit+params.Limit.Dml <= 0) {
			return errdefs.InvalidParameter(errors.New("需设定打码or充值量"), 2)
		}
	}
	return nil
}
