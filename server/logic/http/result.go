package http

import (
	"github.com/gin-gonic/gin"
	Err "gitlab.com/jetfueltw/cpw/alakazam/server/errors"
)

const (
	// OK ok
	OK = 0

	// RequestErr request error
	RequestErr = -400

	contextErrCode = "code"
)

type resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func errors(c *gin.Context, err error) {
	e, ok := err.(Err.Error)
	if !ok {
		e = Err.TypeError
	}
	errorE(c, e)
}

func errorE(c *gin.Context, e Err.Error) {
	c.Set(contextErrCode, e.Code)
	c.JSON(e.Status, e)
}

func result(c *gin.Context, data interface{}, code int) {
	c.Set(contextErrCode, code)
	c.JSON(200, resp{
		Code: code,
		Data: data,
	})
}
