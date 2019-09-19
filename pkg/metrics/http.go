package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net/http"
)

func RunHttp(addr string) {
	go func() {
		log.Infof("metrics server port [%s]", addr)
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/healthz", func(rsp http.ResponseWriter, req *http.Request) {
			rsp.WriteHeader(http.StatusOK)
		})
		err := http.ListenAndServe(addr, nil)
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("metrics server", zap.Error(err))
		}
	}()
}
