package front

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	response "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"gopkg.in/go-playground/validator.v8"
	"net/http"
)

type LuckyMoney struct {
	logic.User

	// 單包金額 or 總金額 看Type種類決定
	// 最少0.01元
	Amount float32 `json:"amount" binding:"required,min=0.01"`

	// 紅包數量,範圍1 ~ 500包
	Count int `json:"count" binding:"required,min=1,max=500"`

	// 紅包說明 下限1字元, 上限20字元
	Message string `json:"message" binding:"required,min=1,max=20"`

	// 紅包種類 拼手氣 or 普通
	Type int `json:"type" binding:"required"`
}

// 發紅包
func (s *Server) giveLuckyMoney(c *gin.Context) {
	arg := new(LuckyMoney)
	if err := validatorLuckyMoney(c, arg); err != nil {
		response.Errors(c, err)
	}

	if err := s.money.Give(arg.Amount, arg.Count, arg.Message, arg.Type); err != nil {
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
