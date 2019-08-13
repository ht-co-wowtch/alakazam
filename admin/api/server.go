package api

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
	e.POST("/blockade/:uid", http.Handler(s.setBlockade))
	e.DELETE("/blockade/:uid", http.Handler(s.removeBlockade))

	// 禁言
	e.POST("/banned/:uid", http.Handler(s.setBanned))
	e.DELETE("/banned/:uid", http.Handler(s.removeBanned))

	// 設定房間
	e.POST("/room", http.Handler(s.CreateRoom))
	e.PUT("/room/:id", http.Handler(s.UpdateRoom))
	e.GET("/room/:id", http.Handler(s.GetRoom))
	e.DELETE("/room/:id", http.Handler(s.DeleteRoom))

	e.POST("/push", http.Handler(s.push))
	e.DELETE("/push/:id", http.Handler(s.deleteTopMessage))
}
