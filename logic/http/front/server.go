package front

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/activity"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"time"
)

type Server struct {
	logic *logic.Logic

	money *activity.LuckyMoney
}

func New(l *logic.Logic, money *activity.LuckyMoney) *Server {
	return &Server{
		logic: l,
		money: money,
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

	e.Use(cors.New(c), http.AuthenticationHandler)

	e.POST("/push/room", s.pushRoom)
	e.POST("/give-lucky-money", s.giveLuckyMoney)
}
