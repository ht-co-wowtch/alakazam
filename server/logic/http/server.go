package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
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

	s.initRouter()
	return s
}

func (s *Server) initRouter() {
	s.engine.POST("/push/room", s.pushRoom)
	s.engine.POST("/push/all", s.pushAll)
	s.engine.GET("/online/room", s.onlineRoom)
}

// Close close the server.
func (s *Server) Close() {
}
