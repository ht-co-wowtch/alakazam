package http

import (
	"context"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	"net/http"
)

// Server is http server.
type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	server *http.Server
	logic  *logic.Logic
}

// New new a http server.
func New(c *conf.HTTPServer, l *logic.Logic) *Server {
	engine := gin.New()
	engine.Use(loggerHandler, recoverHandler)
	s := &Server{
		logic: l,
	}

	s.initRouter(engine)

	s.server = &http.Server{
		Addr:    c.Addr,
		Handler: engine,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func (s *Server) initRouter(engine *gin.Engine) {
	engine.POST("/push/room", s.pushRoom)
	engine.POST("/push/all", s.pushAll)
	engine.GET("/online/room", s.onlineRoom)
}

// Close close the server.
func (s *Server) Close() {
	if err := s.server.Shutdown(s.ctx); err != nil {
		log.Errorf("Server Shutdown:", err)
	} else {
		log.Infof("http server close: %s", s.server.Addr)
	}
	s.cancel()
}
