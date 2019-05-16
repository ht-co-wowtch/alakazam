package admin

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic"
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
	e.POST("/blockade", s.setBlockade)
	e.POST("/banned", s.setBanned)

	e.POST("/push/all", s.pushAll)
}
