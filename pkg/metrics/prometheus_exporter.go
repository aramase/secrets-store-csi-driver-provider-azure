package metrics

import (
	"fmt"
	"net/http"
	"time"

	cgprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"k8s.io/klog/v2"
)

const (
	readHeaderTimeout = 5 * time.Second
)

func initPrometheusExporter(port int) error {
	exporter, err := prometheus.New()
	if err != nil {
		return err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithView(
			metric.NewView(
				metric.Instrument{
					Kind: metric.InstrumentKindHistogram,
				},
				metric.Stream{
					Aggregation: aggregation.ExplicitBucketHistogram{
						// Use custom buckets to avoid the default buckets which are too small for our use case.
						// Start 100ms with last bucket being [~4m, +Inf)
						Boundaries: cgprometheus.ExponentialBucketsRange(0.1, 2, 11),
					},
				},
			),
		),
	)
	global.SetMeterProvider(meterProvider)

	http.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	go func() {
		server := &http.Server{
			Addr:              fmt.Sprintf(":%v", port),
			ReadHeaderTimeout: readHeaderTimeout,
		}
		klog.ErrorS(server.ListenAndServe(), "listen and server error")
	}()

	return err
}
