package admin

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	response "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"net/http"
	"strconv"
)

func (s *Server) SetRoom(c *gin.Context) {
	var params store.Room
	if err := c.ShouldBindJSON(&params); err != nil {
		response.ErrorE(c, errors.DataError)
		return
	}

	if params.Limit.Day > 0 {
		if params.Limit.Dml+params.Limit.Amount <= 0 {
			response.ErrorE(c, errors.SetRoomError.Mes("储值或打码量不可都小于等于0"))
			return
		}
		if params.Limit.Day > 30 {
			response.ErrorE(c, errors.SetRoomError.Mes("储值跟打码量聊天限制天数不能大于30"))
			return
		}
	} else if params.Limit.Day == 0 && params.Limit.Dml+params.Limit.Amount > 0 {
		response.ErrorE(c, errors.SetRoomError.Mes("储值跟打码量都需是0"))
		return
	} else if params.Limit.Day < 0 || params.Limit.Amount < 0 || params.Limit.Dml < 0 {
		response.ErrorE(c, errors.FailureError)
		return
	}
	if !s.logic.SetRoom(params) {
		response.ErrorE(c, errors.FailureError)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) GetRoom(c *gin.Context) {
	id := c.Param("id")

	i, err := strconv.Atoi(id)
	if err != nil {
		response.ErrorE(c, errors.FailureError)
		return
	}

	r, ok := s.logic.GetRoom(i)

	if !ok {
		response.ErrorE(c, errors.NoRowsError)
		return
	}
	c.JSON(http.StatusOK, r)
}
