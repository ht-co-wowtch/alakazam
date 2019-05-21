package admin

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
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
	// 封鎖
	e.POST("/blockade", s.setBlockade)
	e.DELETE("/blockade", s.removeBlockade)

	// 禁言
	e.POST("/banned", s.setBanned)
	e.DELETE("/banned", s.removeBanned)

	// 設定房間
	e.POST("/room", s.SetRoom)

	e.POST("/push/all", s.pushAll)
}
