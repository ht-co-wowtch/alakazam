package front

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	response "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
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
	Type int `json:"type" binding:"required,len=1"`
}

// 發紅包
func (s *Server) giveLuckyMoney(c *gin.Context) {
	arg := new(LuckyMoney)
	if err := c.ShouldBindJSON(arg); err != nil {
		response.ErrorE(c, errors.DataError)
		return
	}

	if err := s.money.Give(arg.Amount, arg.Count, arg.Message, arg.Type); err != nil {
		response.Errors(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
