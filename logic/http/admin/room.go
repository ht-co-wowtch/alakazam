package admin

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	response "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"net/http"
)

func (s *Server) CreateRoom(c *gin.Context) {
	var params store.Room
	if err := bindRoom(c, &params); err != nil {
		response.Errors(c, err)
		return
	}

	roomId, err := s.logic.CreateRoom(params)

	if err != nil {
		response.ErrorE(c, errors.FailureError)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"room_id": roomId,
	})
}

func (s *Server) UpdateRoom(c *gin.Context) {
	var params store.Room
	if err := bindRoom(c, &params); err != nil {
		response.Errors(c, err)
		return
	}

	if !s.logic.UpdateRoom(c.Param("id"), params) {
		response.ErrorE(c, errors.FailureError)
		return
	}

	c.Status(http.StatusNoContent)
}

func (s *Server) GetRoom(c *gin.Context) {
	r, ok := s.logic.GetRoom(c.Param("id"))
	if !ok {
		response.ErrorE(c, errors.NoRowsError)
		return
	}
	c.JSON(http.StatusOK, r)
}

func bindRoom(c *gin.Context, params *store.Room) error {
	if err := c.ShouldBindJSON(params); err != nil {
		return errors.DataError
	}
	if params.Limit.Day > 0 {
		if params.Limit.Dml+params.Limit.Amount <= 0 {
			return errors.SetRoomError.Mes("储值或打码量不可都小于等于0")
		}
		if params.Limit.Day > 30 {
			return errors.SetRoomError.Mes("储值跟打码量聊天限制天数不能大于30")
		}
	} else if params.Limit.Day == 0 && params.Limit.Dml+params.Limit.Amount > 0 {
		return errors.SetRoomError.Mes("储值跟打码量都需是0")
	} else if params.Limit.Day < 0 || params.Limit.Amount < 0 || params.Limit.Dml < 0 {
		return errors.FailureError
	}
	return nil
}
