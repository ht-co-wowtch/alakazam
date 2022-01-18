package metrics

import (
	"net/http"
	"net/http/pprof"

	// "runtime/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/ht-co/micro/log"
	"go.uber.org/zap"
)

func RunHttp(addr string) {
	go func() {
		log.Infof("metrics server port [%s]", addr)
		gin.SetMode(gin.ReleaseMode)
		e := gin.New()

		e.GET("/metrics", gin.WrapH(promhttp.Handler()))
		e.GET("/healthz", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})
		pprofHandler(e)
		err := e.Run(addr)
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("metrics server", zap.Error(err))
		}
	}()
}

func pprofHandler(e *gin.Engine) {
	prefixRouter := e.Group("/debug/pprof")
	{
		prefixRouter.GET("/", gin.WrapF(pprof.Index))
		prefixRouter.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		prefixRouter.GET("/profile", gin.WrapF(pprof.Profile))
		prefixRouter.POST("/symbol", gin.WrapF(pprof.Symbol))
		prefixRouter.GET("/symbol", gin.WrapF(pprof.Symbol))
		prefixRouter.GET("/trace", gin.WrapF(pprof.Trace))
		prefixRouter.GET("/allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))
		prefixRouter.GET("/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
		prefixRouter.GET("/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
		prefixRouter.GET("/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
		prefixRouter.GET("/mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
		prefixRouter.GET("/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))
	}
}
