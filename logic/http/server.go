package http

import (
	"context"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	server "gitlab.com/jetfueltw/cpw/micro/http"
	"go.uber.org/zap"
	"net/http"
)

type LogicHttpServer interface {
	InitRoute(e *gin.Engine)
}

// Server is http server.
type Server struct {
	ctx context.Context

	cancel context.CancelFunc

	// http server 結構
	server *http.Server

	// 各個不同的http server (後台 or 前台)
	logic LogicHttpServer
}

// New new a http server.
func New(c *server.Conf, srv LogicHttpServer) *Server {
	engine := gin.New()
	engine.Use(recoverHandler)
	s := &Server{
		server: server.NewServer(c, engine),
		logic:  srv,
	}

	s.logic.InitRoute(engine)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

type HandlerFunc func(*gin.Context) error

func Handler(f HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := f(c); err != nil {
			errHandler(c, err)
		}
	}
}

func errHandler(c *gin.Context, err error) {
	e := errdefs.Err(err)
	if e.Status == http.StatusInternalServerError {
		log.Error(e, zap.Error(e.Err))
	}
	c.JSON(e.Status, e)
}

// Close close the server.
func (s *Server) Close() {
	if err := s.server.Shutdown(s.ctx); err != nil {
		log.Errorf("Server Shutdown: error(%v)", err)
	} else {
		log.Infof("http server close: %s", s.server.Addr)
	}
	s.cancel()
}
