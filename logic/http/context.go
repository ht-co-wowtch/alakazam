package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"time"
)

type Context struct {
	logic *logic.Logic

	client *client.Client
}

func NewContext(l *logic.Logic, client *client.Client) *Context {
	return &Context{
		logic:  l,
		client: client,
	}
}

func (s *Context) InitRoute(e *gin.Engine) {
	c := cors.Config{
		AllowCredentials: true,
		AllowOrigins:     conf.Conf.HTTPServer.Cors.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT"},
		AllowHeaders:     conf.Conf.HTTPServer.Cors.Headers,
		MaxAge:           time.Minute * 5,
	}

	e.Use(cors.New(c), AuthenticationHandler)

	e.POST("/push/room", Handler(s.pushRoom))
	e.POST("/red-envelope", Handler(s.giveRedEnvelope))
	e.PUT("/red-envelope", Handler(s.takeRedEnvelope))
	e.GET("/red-envelope/:id", Handler(s.getRedEnvelopeDetail))
	e.GET("/red-envelope-consume/:id", Handler(s.getRedEnvelope))
}
