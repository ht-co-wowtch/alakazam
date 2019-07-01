package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"net/http"
)

func (s *Server) CreateRoom(c *gin.Context) error {
	var params logic.Room
	if err := bindRoom(c, &params); err != nil {
		return err
	}

	roomId, err := s.logic.CreateRoom(params)

	if err != nil {
		return errors.FailureError
	}
	c.JSON(http.StatusOK, gin.H{
		"room_id": roomId,
	})
	return nil
}

func (s *Server) UpdateRoom(c *gin.Context) error {
	var params logic.Room
	if err := bindRoom(c, &params); err != nil {
		return err
	}

	if !s.logic.UpdateRoom(params) {
		return errors.FailureError
	}

	c.Status(http.StatusNoContent)
	return nil
}

func (s *Server) GetRoom(c *gin.Context) error {
	r, ok := s.logic.GetRoom(c.Param("id"))
	if !ok {
		return errors.NoRowsError
	}
	c.JSON(http.StatusOK, gin.H{
		"id":                  r.Id,
		"is_message":          r.IsMessage,
		"day_limit":           r.DayLimit,
		"deposit_limit":       r.DepositLimit,
		"dml_limit":           r.DmlLimit,
		"red_envelope_expire": r.RedEnvelopeExpire,
		"create_at":           r.CreateAt,
		"update_at":           r.UpdateAt,
	})
	return nil
}

func bindRoom(c *gin.Context, params *logic.Room) error {
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
	} else if params.Limit.Day < 0 || params.Limit.Deposit < 0 || params.Limit.Dml < 0 {
		return errors.FailureError
	} else if params.Limit.Day == 0 && params.Limit.Dml > 0 {
		return errors.SetRoomError.Mes("需设定充值天数")
	} else if params.Limit.Day == 0 && params.Limit.Deposit > 0 {
		return errors.SetRoomError.Mes("需设定充值天数")
	}
	return nil
}
