package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	api "gitlab.com/jetfueltw/cpw/alakazam/app/logic/api/http"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	web "gitlab.com/jetfueltw/cpw/micro/http"
	//_ "net/http/pprof"
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
	engine.Use(api.RecoverHandler) // use RecoverHandler's middleware

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

	handler(engine, srv) // http api router
	return web.NewServer(conf, engine)
}

// http api router of admin
func handler(e *gin.Engine, s *httpServer) {

	// 心跳
	e.GET("/healthz", healthz)
	// 封鎖
	// 全站
	e.POST("/blockade/:uid", api.ErrHandler(s.setBlockade))
	e.DELETE("/blockade/:uid", api.ErrHandler(s.removeBlockade))
	// 單一房間
	e.POST("/blockade/:uid/room/:id", api.ErrHandler(s.setBlockade))
	e.DELETE("/blockade/:uid/room/:id", api.ErrHandler(s.removeBlockade))

	// 禁言
	// 全站
	e.POST("/banned/:uid", api.ErrHandler(s.setBanned))
	e.DELETE("/banned/:uid", api.ErrHandler(s.removeBanned))
	// 單一房間
	e.POST("/banned/:uid/room/:id", api.ErrHandler(s.setBanned))
	e.DELETE("/banned/:uid/room/:id", api.ErrHandler(s.removeBanned))

	// 踢人
	e.DELETE("/kick/:uid", api.ErrHandler(s.kick))

	// 房間
	e.POST("/room", api.ErrHandler(s.CreateRoom))
	e.PUT("/room/:id", api.ErrHandler(s.UpdateRoom))
	e.GET("/room/:id", api.ErrHandler(s.GetRoom))
	e.DELETE("/room/:id", api.ErrHandler(s.DeleteRoom))

	// 房管
	e.POST("/room/manage", api.ErrHandler(s.AddManage))
	e.DELETE("/room/:id/manage/:uid", api.ErrHandler(s.DeleteManage))

	e.POST("/room/notice", api.ErrHandler(s.notice)) // todo
	e.GET("/online", api.ErrHandler(s.online)) // 所有房間在線人數

	// 訊息
	e.POST("/custom", api.ErrHandler(s.custom)) // todo 客制訊息內容
	// 一般/置頂/公告
	e.POST("/push", api.ErrHandler(s.push))
	e.DELETE("/push/:id", api.ErrHandler(s.deleteTopMessage))
	e.POST("/red-envelope", api.ErrHandler(s.giveRedEnvelope)) // 紅包
	e.POST("/bets", api.ErrHandler(s.bets)) // 下注
	e.POST("/betsWin", api.ErrHandler(s.betsWin)) // 中獎
	e.POST("/gift", api.ErrHandler(s.gift)) // 送禮
	e.POST("/reward", api.ErrHandler(s.reward)) // 打賞
	e.POST("/follow", api.ErrHandler(s.follow)) // 追隨主播

	// 敏感詞
	e.POST("/shield", api.ErrHandler(s.CreateShield))
	e.PUT("/shield", api.ErrHandler(s.UpdateShield))
	e.DELETE("/shield/:id", api.ErrHandler(s.DeleteShield))

	// 會員資料
	e.GET("/room/:id/user/:uid", api.ErrHandler(s.profile))
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

//used for aws ALB
func healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
