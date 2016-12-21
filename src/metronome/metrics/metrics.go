package metrics

import (
	"net/http"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/spf13/viper"
)

var onceMetricsDefaults sync.Once

func setDefaults() {
	onceMetricsDefaults.Do(func() {
		viper.SetDefault("metrics.addr", ":9100")
		viper.SetDefault("metrics.path", "/metrics")
	})
}

// Serve start the metrics endpoint
func Serve() {
	setDefaults()

	http.Handle(viper.GetString("metrics.path"), promhttp.Handler())
	addr := viper.GetString("metrics.addr")
	go func() {
		log.Fatal(http.ListenAndServe(addr, nil))
	}()
}
