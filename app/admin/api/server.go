package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	api "gitlab.com/jetfueltw/cpw/alakazam/app/logic/api/http"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	web "gitlab.com/jetfueltw/cpw/micro/http"
	"net/http"
)

type httpServer struct {
	member       *member.Member
	message      *message.Producer
	shield       message.Filter
	delayMessage *message.DelayProducer
	room         room.Room
	nidoran      *client.Client
	noticeUrl    map[string]string
}

func NewServer(conf *web.Conf, member *member.Member, producer *message.Producer, room room.Room, nidoran *client.Client, shield message.Filter) *http.Server {
	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := web.NewHandler()
	engine.Use(api.RecoverHandler)

	delayMessage := message.NewDelayProducer(producer, nidoran)
	delayMessage.Start()

	srv := &httpServer{
		member:       member,
		message:      producer,
		shield:       shield,
		delayMessage: delayMessage,
		room:         room,
		nidoran:      nidoran,
		noticeUrl:    make(map[string]string),
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
	e.DELETE("/banned/:uid/room/:id", api.ErrHandler(s.removeBanned))

	// 踢人
	e.DELETE("/kick/:uid", api.ErrHandler(s.kick))

	// 房間
	e.POST("/room", api.ErrHandler(s.CreateRoom))
	e.PUT("/room/:id", api.ErrHandler(s.UpdateRoom))
	e.GET("/room/:id", api.ErrHandler(s.GetRoom))
	e.DELETE("/room/:id", api.ErrHandler(s.DeleteRoom))
	e.POST("/room/manage", api.ErrHandler(s.AddManage))
	e.DELETE("/room/:id/manage/:uid", api.ErrHandler(s.DeleteManage))
	e.POST("/room/notice", api.ErrHandler(s.notice))
	e.GET("/online", api.ErrHandler(s.online))

	// 訊息
	e.POST("/custom", api.ErrHandler(s.custom))
	e.POST("/push", api.ErrHandler(s.push))
	e.DELETE("/push/:id", api.ErrHandler(s.deleteTopMessage))
	e.POST("/red-envelope", api.ErrHandler(s.giveRedEnvelope))
	e.POST("/bets", api.ErrHandler(s.bets))
	e.POST("/betsWin", api.ErrHandler(s.betsWin))
	e.POST("/gift", api.ErrHandler(s.gift))
	e.POST("/reward", api.ErrHandler(s.reward))
	e.POST("/follow", api.ErrHandler(s.follow))

	// 敏感詞
	e.POST("/shield", api.ErrHandler(s.CreateShield))
	e.PUT("/shield", api.ErrHandler(s.UpdateShield))
	e.DELETE("/shield/:id", api.ErrHandler(s.DeleteShield))

	// 會員資料
	e.PUT("/profile/:token/renew", api.ErrHandler(s.renew))
}

func (s *httpServer) Close() error {
	if err := s.message.Close(); err != nil {
		return fmt.Errorf("message producer close error(%v)", err)
	}
	if err := s.delayMessage.Close(); err != nil {
		return fmt.Errorf("delay message producer close error(%v)", err)
	}
	return nil
}
