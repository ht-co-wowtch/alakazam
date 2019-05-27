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
		if params.Limit.Day > 31 {
			return errors.SetRoomError.Mes("限制充值聊天天数不能大于31")
		}
		if params.Limit.Dml <= 0 {
			return errors.SetRoomError.Mes("打码量不可小于等于0")
		}
	} else if params.Limit.Day == 0 && params.Limit.Dml > 0 {
		return errors.SetRoomError.Mes("需设定充值天数")
	} else if params.Limit.Day < 0 || params.Limit.Amount < 0 || params.Limit.Dml < 0 {
		return errors.FailureError
	}
	return nil
}
