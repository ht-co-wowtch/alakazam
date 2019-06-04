package front

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/activity"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	response "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"gopkg.in/go-playground/validator.v8"
	"net/http"
)

type LuckyMoney struct {
	logic.User

	activity.GiveMoney
}

// 發紅包
func (s *Server) giveLuckyMoney(c *gin.Context) {
	arg := new(LuckyMoney)
	if err := validatorLuckyMoney(c, arg); err != nil {
		response.Errors(c, err)
		return
	}

	if err := s.logic.Auth(&arg.User); err != nil {
		response.Errors(c, err)
		return
	}

	arg.Token = c.GetString("token")

	id, err := s.money.Give(&arg.GiveMoney)

	if err != nil {
		response.Errors(c, err)
		return
	}

	if err := s.logic.PushMoney(id, arg.Message, &arg.User); err != nil {
		response.Errors(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func validatorLuckyMoney(c *gin.Context, arg *LuckyMoney) error {
	var err error
	if err = c.ShouldBindJSON(arg); err != nil {
		v := err.(validator.ValidationErrors)
		for _, e := range v {
			switch e.Field {
			case "Amount":
				err = errors.DataError.Mes("红包金额最低0.01")
			case "Count":
				err = errors.DataError.Mes("红包最大数量是500")
			case "Message":
				err = errors.DataError.Mes("限制文字长度为1到20个字")
			default:
				err = errors.DataError
			}
		}
	}

	return err
}
