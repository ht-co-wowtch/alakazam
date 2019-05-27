package front

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"time"
)

type Server struct {
	logic *logic.Logic
}

func New(l *logic.Logic) *Server {
	return &Server{
		logic: l,
	}
}

func (s *Server) InitRoute(e *gin.Engine) {
	c := cors.Config{
		AllowCredentials: true,
		AllowOrigins:     conf.Conf.HTTPServer.Cors.Origins,
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     conf.Conf.HTTPServer.Cors.Headers,
		MaxAge:           time.Minute * 5,
	}
	e.Use(cors.New(c))

	e.POST("/push/room", s.pushRoom)
}
