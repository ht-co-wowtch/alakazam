package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"strings"
)

func AuthenticationHandler(c *gin.Context) {
	authorization := c.GetHeader("Authorization")
	token := strings.Split(authorization, " ")

	if len(token) != 2 || token[0] != "Bearer" || token[1] == "" {
		e := errdefs.Err(errors.ErrAuthorization)
		c.AbortWithStatusJSON(e.Status, e)
		return
	}

	c.Set("token", token[1])
	c.Next()
}
