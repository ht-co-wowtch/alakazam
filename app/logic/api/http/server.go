package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	web "gitlab.com/jetfueltw/cpw/micro/http"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
	"net/http/httputil"
	"runtime"
	"strings"
	"time"
)

type httpServer struct {
	member  *member.Member
	message *message.Producer
	room    *room.Room
	client  *client.Client
}

func NewServer(conf *web.Conf, member *member.Member, message *message.Producer, room *room.Room, client *client.Client) *http.Server {
	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := web.NewHandler()
	engine.Use(RecoverHandler)

	srv := httpServer{
		member:  member,
		message: message,
		room:    room,
		client:  client,
	}

	c := cors.Config{
		AllowCredentials: true,
		AllowOrigins:     conf.Cors.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT"},
		AllowHeaders:     conf.Cors.Headers,
		MaxAge:           time.Minute * 5,
	}
	engine.Use(cors.New(c), AuthenticationHandler)
	handler(engine, srv)
	return web.NewServer(conf, engine)
}

func handler(e *gin.Engine, s httpServer) {
	e.POST("/push/room", ErrHandler(s.pushRoom))
	e.POST("/red-envelope", ErrHandler(s.giveRedEnvelope))
	e.PUT("/red-envelope", ErrHandler(s.takeRedEnvelope))
	e.GET("/red-envelope/:id", ErrHandler(s.getRedEnvelopeDetail))
	e.GET("/red-envelope-consume/:id", ErrHandler(s.getRedEnvelope))
	e.GET("/message/:room", ErrHandler(s.getMessage))
}

func AuthenticationHandler(c *gin.Context) {
	authorization := c.GetHeader("Authorization")
	token := strings.Split(authorization, " ")

	if len(token) != 2 || token[0] != "Bearer" || token[1] == "" {
		e := errdefs.Err(errors.ErrAuthorization)
		c.AbortWithStatusJSON(e.Status, e)
		return
	}

	c.Set("token", token[1])
	c.Next()
}

var errInternalServer = errdefs.New(0, 0, "应用程序错误")

// try catch log
func RecoverHandler(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			httprequest, _ := httputil.DumpRequest(c.Request, false)

			c.AbortWithStatusJSON(http.StatusInternalServerError, errInternalServer)

			log.DPanic("[Recovery]",
				zap.Time("time", time.Now()),
				zap.String("panic", string(httprequest)),
				zap.Any("error", err),
				zap.String("stack", string(buf)),
			)
		}
	}()

	c.Next()
}

type handlerFunc func(*gin.Context) error

func ErrHandler(f handlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := f(c); err != nil {
			e := errdefs.Err(err)
			if e.Status == http.StatusInternalServerError {
				log.Error(
					"api error",
					zap.String("path", c.Request.URL.Path),
					zap.String("rawQuery", c.Request.URL.RawQuery),
					zap.String("method", c.Request.Method),
					zap.Error(e.Err),
				)
			}
			c.JSON(e.Status, e)
		}
	}
}
