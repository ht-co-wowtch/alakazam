package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic/conf"
)

// Server is http server.
type Server struct {
	engine *gin.Engine
	logic  *logic.Logic
}

// New new a http server.
func New(c *conf.HTTPServer, l *logic.Logic) *Server {
	engine := gin.New()
	engine.Use(loggerHandler, recoverHandler)
	go func() {
		if err := engine.Run(c.Addr); err != nil {
			panic(err)
		}
	}()
	s := &Server{
		engine: engine,
		logic:  l,
	}

	if c.IsStage {
		s.initRouter()
	} else {
		s.initV1()
	}
	return s
}

func (s *Server) initRouter() {
	s.engine.POST("/push/keys", s.pushKeys)
	s.engine.POST("/push/mids", s.pushMids)
	s.engine.POST("/push/room", s.pushRoom)
	s.engine.POST("/push/all", s.pushAll)
	s.engine.GET("/online/room", s.onlineRoom)
}

func (s *Server) initV1() {
	v1 := s.engine.Group("/v1")
	v1.POST("/healthy", s.base)
	v1.GET("/healthy", s.base)
	/*
		action          from         to         do
		禁言 mute		 room_id     user_id    detail..
		封鎖 Banned     room_id     user_id
		紅包 red        room_id     user_id
		假人 faker      room_id     user_id
		v1.POST("/action/from/to/content")
	*/

	//禁言
	v1.POST("/jinyan/:from/:to/:content", s.jinyan)
	//封鎖
	v1.POST("/fengsuo/:from/:to/:content", s.fengsuo)
	//紅包
	v1.POST("/hongbao/:from/:to/:content", s.hongbao)
	//假人
	v1.POST("/faker/:from/:to/:content", s.faker)
}

// Close close the server.
func (s *Server) Close() {
}
