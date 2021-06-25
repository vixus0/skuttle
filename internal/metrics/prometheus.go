package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/vixus0/skuttle/v2/internal/logging"
)

var (
	skuttleTerminationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skuttle_teminations_total",
			Help: "Total number of EC2 instance terminations",
		},
		[]string{"az", "region", "instance_type"},
	)
	skuttleTerminationSkipsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skuttle_temination_skips_total",
			Help: "Total number of EC2 instance terminations skipped",
		},
		[]string{"az", "region", "instance_type"},
	)
	skuttleTerminationErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skuttle_termination_errors_total",
			Help: "Total number of errors terminating EC2 instances",
		},
		[]string{"az", "region", "instance_type"},
	)
)

func RecordNodeTermination(az, region, itype string) {
	skuttleTerminationsTotal.With(prometheus.Labels{
		"az":            az,
		"region":        region,
		"instance_type": itype,
	}).Inc()
}

func RecordNodeTerminationError(az, region, itype string) {
	skuttleTerminationErrorsTotal.With(prometheus.Labels{
		"az":            az,
		"region":        region,
		"instance_type": itype,
	}).Inc()
}

func RecordNodeTerminationSkip(az, region, itype string) {
	skuttleTerminationSkipsTotal.With(prometheus.Labels{
		"az":            az,
		"region":        region,
		"instance_type": itype,
	}).Inc()
}

type PrometheusExporter interface {
	Run()
	Wait()
}

type prometheusExporter struct {
	host string
	port int

	logger *logging.Logger
	server *http.Server
	wg     sync.WaitGroup
}

func NewPrometheusExporter(log *logging.Logger, host string, port int) *prometheusExporter {
	return &prometheusExporter{
		host: host,
		port: port,

		logger: log,
	}
}

func (p *prometheusExporter) Run() {
	promHandler := promhttp.Handler()

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(rw http.ResponseWriter, req *http.Request) {
		p.logger.Info("serving /metrics")
		promHandler.ServeHTTP(rw, req)
		p.logger.Info("served /metrics")
	})

	p.server = &http.Server{
		Addr: fmt.Sprintf("%s:%d", p.host, p.port),

		Handler: mux,

		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
	}

	p.wg.Add(1)
	p.server.ListenAndServe()
	p.wg.Done()
}

func (p *prometheusExporter) Wait() {
	p.server.Shutdown(context.Background())
	p.wg.Wait()
}
