package proxymetrics

import (
	"orla-alerts/solte.lab/src/api"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var sourceProxyMetrics *ProxyWorker

type ProxyWorker struct {
	dataNetFlow prometheus.Gauge
}

func (sp *ProxyWorker) Register(s *api.Server) (string, chi.Router) {
	routes := chi.NewRouter()
	routes.Use(s.SetRequestID)
	routes.Use(s.LogMetricsRequest)

	routes.Handle("/", promhttp.Handler())
	return "/metrics", routes
}

func NewSourcesProxy() *ProxyWorker {

	var (
		dataNetFlow = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "data_net_flow",
			Help: "Data Net Flow",
		})
	)
	sourceProxyMetrics = &ProxyWorker{dataNetFlow: dataNetFlow}
	return sourceProxyMetrics
}

func (sp *ProxyWorker) CountNetFlow(number int) {
	sp.dataNetFlow.Set(float64(number))
}
