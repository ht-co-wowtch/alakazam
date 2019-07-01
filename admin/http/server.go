package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http"
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
	e.POST("/blockade", http.Handler(s.setBlockade))
	e.DELETE("/blockade", http.Handler(s.removeBlockade))

	// 禁言
	e.POST("/banned", http.Handler(s.setBanned))
	e.DELETE("/banned", http.Handler(s.removeBanned))

	// 設定房間
	e.POST("/room", http.Handler(s.CreateRoom))
	e.PUT("/room", http.Handler(s.UpdateRoom))
	e.GET("/room/:id", http.Handler(s.GetRoom))

	e.POST("/push/all", http.Handler(s.pushAll))

}
