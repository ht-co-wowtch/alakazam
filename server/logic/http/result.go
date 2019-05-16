package http

import (
	"github.com/gin-gonic/gin"
	Err "gitlab.com/jetfueltw/cpw/alakazam/server/errors"
)

const (
	contextErrCode = "code"
)

type resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Errors(c *gin.Context, err error) {
	e, ok := err.(Err.Error)
	if !ok {
		e = Err.TypeError
	}
	ErrorE(c, e)
}

func ErrorE(c *gin.Context, e Err.Error) {
	c.Set(contextErrCode, e.Code)
	c.JSON(e.Status, e)
}

func Result(c *gin.Context, data interface{}, code int) {
	c.Set(contextErrCode, code)
	c.JSON(200, resp{
		Code: code,
		Data: data,
	})
}
