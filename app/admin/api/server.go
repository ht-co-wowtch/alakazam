package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	api "gitlab.com/jetfueltw/cpw/alakazam/app/logic/api/http"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	web "gitlab.com/jetfueltw/cpw/micro/http"
	"net/http"
)

type httpServer struct {
	member  *member.Member
	message *message.Producer
	room    *room.Room
	nidoran *client.Client
}

func NewServer(conf *web.Conf, member *member.Member, message *message.Producer, room *room.Room, nidoran *client.Client) *http.Server {
	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := web.NewHandler()
	engine.Use(api.RecoverHandler)

	srv := &httpServer{
		member:  member,
		message: message,
		room:    room,
		nidoran: nidoran,
	}

	handler(engine, srv)
	return web.NewServer(conf, engine)
}

func handler(e *gin.Engine, s *httpServer) {
	// 封鎖
	e.POST("/blockade/:uid", api.ErrHandler(s.setBlockade))
	e.DELETE("/blockade/:uid", api.ErrHandler(s.removeBlockade))

	// 禁言
	e.POST("/banned/:uid", api.ErrHandler(s.setBanned))
	e.DELETE("/banned/:uid", api.ErrHandler(s.removeBanned))

	// 設定房間
	e.POST("/room", api.ErrHandler(s.CreateRoom))
	e.PUT("/room/:id", api.ErrHandler(s.UpdateRoom))
	e.GET("/room/:id", api.ErrHandler(s.GetRoom))
	e.DELETE("/room/:id", api.ErrHandler(s.DeleteRoom))

	e.POST("/push", api.ErrHandler(s.push))
	e.DELETE("/push/:id", api.ErrHandler(s.deleteTopMessage))

	e.POST("/red-envelope", api.ErrHandler(s.giveRedEnvelope))
}
