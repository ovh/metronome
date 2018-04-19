package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Serve start the metrics endpoint
func Serve() {
	http.Handle(viper.GetString("metrics.path"), promhttp.Handler())
	addr := viper.GetString("metrics.addr")
	go func() {
		log.Fatal(http.ListenAndServe(addr, nil))
	}()
}
