package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"strings"
)

func AuthenticationHandler(c *gin.Context) {
	authorization := c.GetHeader("Authorization")
	token := strings.Split(authorization, " ")

	if len(token) != 2 || token[0] != "Bearer" || token[1] == "" {
		c.AbortWithStatusJSON(errors.AuthorizationError.Status, errors.AuthorizationError)
		return
	}

	c.Set("token", token[1])
	c.Next()
}
