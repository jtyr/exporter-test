package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

const (
	appName   = "exporter-test"
	meterName = "myMeter"
)

var (
	commonLabels = []attribute.KeyValue{
		attribute.String("app", appName)}
	reqCounter metric.Float64Counter
	logger     log.Logger
	build      string
)

type myExporter struct {
	Exporter *prometheus.Exporter
}

// initLogger creates new logger used throughout the application.
func initLogger() {
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(
		logger,
		"t", log.DefaultTimestampUTC,
		"app", appName,
		"build", build)
}

// initMeter creates new Prometheus exporter.
func initMeter() *prometheus.Exporter {
	exporter, err := prometheus.InstallNewPipeline(prometheus.Config{})
	if err != nil {
		level.Error(logger).Log(
			"msg", "failed to initialize Prometheus exporter",
			"err", err)
		os.Exit(1)
	}

	meter := global.Meter(meterName)

	ctx := context.Background()

	// Init the metrics
	reqCounter = metric.Must(meter).NewFloat64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of requests"))
	reqCounter.Add(ctx, float64(0), commonLabels...)

	// Start collecting runtime metrics
	if err = runtime.Start(); err != nil {
		level.Error(logger).Log(
			"msg", "failed to initialize runtime metrics collection",
			"err", err)
		os.Exit(1)
	}

	// Start collecting host metrics
	if err = host.Start(); err != nil {
		level.Error(logger).Log(
			"msg", "failed to initialize host metrics collection",
			"err", err)
		os.Exit(1)
	}

	return exporter
}

func (e *myExporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	level.Info(logger).Log(
		"remoteAddr", r.RemoteAddr)

	ctx := r.Context()
	reqCounter.Add(ctx, float64(1), commonLabels...)

	e.Exporter.ServeHTTP(w, r)
}

func MyServeHTTP(w http.ResponseWriter, r *http.Request) {
	level.Info(logger).Log(
		"msg", "Hello from the root endpoint")

	ctx := r.Context()
	reqCounter.Add(ctx, float64(1), commonLabels...)

	fmt.Fprintf(w, "Hello world\n")
}

func healtcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok\n")
}

func main() {
	// Init logger
	initLogger()

	// Init meter
	meter := initMeter()
	var myExp myExporter
	myExp.Exporter = meter

	// Setup HTTP server
	listen := os.Getenv("SERVER_LISTEN")

	if listen == "" {
		listen = "127.0.0.1:8080"
	}

	level.Info(logger).Log(
		"msg", fmt.Sprintf("Listening on %s", listen))

	rootHandler := otelhttp.NewHandler(http.HandlerFunc(MyServeHTTP), "root")
	metricsHandler := otelhttp.NewHandler(http.HandlerFunc(myExp.ServeHTTP), "metrics")

	http.Handle("/", rootHandler)
	http.Handle("/metrics", metricsHandler)
	http.HandleFunc("/healthcheck", healtcheckHandler)

	err := http.ListenAndServe(listen, nil)
	if err != nil {
		level.Error(logger).Log(
			"msg", "cannot create HTTP server",
			"err", err)
		os.Exit(1)
	}
}
