package front

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"time"
)

type Server struct {
	logic *logic.Logic

	client *client.Client
}

func New(l *logic.Logic, client *client.Client) *Server {
	return &Server{
		logic:  l,
		client: client,
	}
}

func (s *Server) InitRoute(e *gin.Engine) {
	c := cors.Config{
		AllowCredentials: true,
		AllowOrigins:     conf.Conf.HTTPServer.Cors.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT"},
		AllowHeaders:     conf.Conf.HTTPServer.Cors.Headers,
		MaxAge:           time.Minute * 5,
	}

	e.Use(cors.New(c), http.AuthenticationHandler)

	e.POST("/push/room", http.Handler(s.pushRoom))
	e.POST("/red-envelope", http.Handler(s.giveRedEnvelope))
	e.PUT("/red-envelope", http.Handler(s.takeRedEnvelope))
	e.GET("/red-envelope/:id", http.Handler(s.getRedEnvelope))
}
