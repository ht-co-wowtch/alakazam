package http

import (
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
	"net/http/httputil"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// http request log
//func loggerHandler(c *gin.Context) {
//	// Start timer
//	start := time.Now()
//	path := c.Request.URL.Path
//	raw := c.Request.URL.RawQuery
//	method := c.Request.Method
//
//	// Process request
//	c.Next()
//
//	// Stop timer
//	end := time.Now()
//	latency := end.Sub(start)
//	statusCode := c.Writer.Status()
//	ecode := c.GetInt(contextErrCode)
//	clientIP := c.ClientIP()
//	if raw != "" {
//		path = path + "?" + raw
//	}
//	log.Infof("METHOD:%s | PATH:%s | CODE:%d | IP:%s | TIME:%d | ECODE:%d", method, path, statusCode, clientIP, latency/time.Millisecond, ecode)
//}

var errInternalServer = errdefs.New(0, 0, "应用程序错误")

// try catch log
func recoverHandler(c *gin.Context) {
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
