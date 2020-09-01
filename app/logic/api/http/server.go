package http

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	web "gitlab.com/jetfueltw/cpw/micro/http"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"runtime"
	"strings"
	"time"
)

type httpServer struct {
	member  *member.Member
	history *message.History
	jwt     *member.Jwt
	msg     *msg
}

func NewServer(conf *conf.Config, me *member.Member, message *message.Producer, room room.Chat, client *client.Client, history *message.History) *http.Server {
	if conf.HTTPServer.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	srv := httpServer{
		member:  me,
		history: history,
		jwt:     member.NewJwt(conf.JwtSecret),
		msg: &msg{
			room:    room,
			client:  client,
			message: message,
			member:  me,
		},
	}

	c := cors.Config{
		AllowCredentials: true,
		AllowOrigins:     conf.HTTPServer.Cors.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT"},
		AllowHeaders:     conf.HTTPServer.Cors.Headers,
		MaxAge:           time.Minute * 5,
	}

	engine := web.NewHandler()
	engine.Use(RecoverHandler, cors.New(c), authenticationHandler)
	handler(engine, srv)
	server := web.NewServer(conf.HTTPServer, engine)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(err.Error())
		}
	}()
	return server
}

func handler(e *gin.Engine, s httpServer) {
	// 禁言
	e.POST("/banned/:uid/room/:id", s.authUid, ErrHandler(s.setBanned))
	e.DELETE("/banned/:uid/room/:id", s.authUid, ErrHandler(s.removeBanned))

	e.POST("/push/room", s.authUid, ErrHandler(s.pushRoom))
	e.POST("/push/key", s.authUid, ErrHandler(s.pushKey))
	e.POST("/red-envelope", s.authUid, ErrHandler(s.giveRedEnvelope))
	e.PUT("/red-envelope", s.authUid, ErrHandler(s.takeRedEnvelope))
	e.GET("/red-envelope/:id", ErrHandler(s.getRedEnvelopeDetail))
	e.GET("/message/:room", ErrHandler(s.getMessage))
}

func authenticationHandler(c *gin.Context) {
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

func (h *httpServer) authUid(c *gin.Context) {
	claims, err := h.jwt.Parse(c.GetString("token"))
	if err != nil {
		e := errdefs.Err(err)
		c.AbortWithStatusJSON(e.Status, e)
		return
	}

	uid, ok := claims["uid"]
	if !ok {
		log.Error("token not found uid")
		c.AbortWithStatusJSON(http.StatusForbidden, errors.ErrTokenUid)
		return
	}

	c.Set("uid", uid.(string))
}

func (s *httpServer) Close() error {
	if err := s.msg.message.Close(); err != nil {
		return fmt.Errorf("message producer close error(%v)", err)
	}
	return nil
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
			if e.Err != nil {
				var b []byte
				if c.Request.Method == "POST" || c.Request.Method == "PUT" {
					b, _ = ioutil.ReadAll(c.Request.Body)
				}

				log.Error(
					"api error",
					zap.Int("code", e.Code),
					zap.String("path", c.Request.URL.Path),
					zap.String("rawQuery", c.Request.URL.RawQuery),
					zap.String("method", c.Request.Method),
					zap.String("body", string(b)),
					zap.Error(e.Err),
				)
			}
			if e.Errors == nil {
				e.Errors = map[string]string{}
			}
			c.JSON(e.Status, e)
		}
	}
}
