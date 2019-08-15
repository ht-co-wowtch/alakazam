package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/member"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/message"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/room"
)

type Server struct {
	member  *member.Member
	message *message.Producer
	room    *room.Room
	logic   *logic.Logic
}

func New(l *logic.Logic, member *member.Member, room *room.Room, message *message.Producer) *Server {
	return &Server{
		member:  member,
		room:    room,
		logic:   l,
		message: message,
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

	e.POST("/red-envelope", http.Handler(s.giveRedEnvelope))
}
