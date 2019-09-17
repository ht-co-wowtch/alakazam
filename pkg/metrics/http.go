package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"net/http"
)

func RunHttp(addr string) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(addr, nil)
		if err != nil && err != http.ErrServerClosed {
			log.Error(err.Error())
		}
	}()
}
