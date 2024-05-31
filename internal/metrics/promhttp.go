package metrics

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-metrics"
	"github.com/hashicorp/go-metrics/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/promauto" // Side effects needed for promauto metrics to work
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	mon      *prometheus.PrometheusSink
	counters map[string]int64
}

var globalMetrics Metrics

func init() {
	mon, err := prometheus.NewPrometheusSink()
	if err != nil {
		log.Fatalf("cannot initialize prometheus sink: %s", err)
	}
	_, err = metrics.NewGlobal(metrics.DefaultConfig("home_controller"), mon)
	if err != nil {
		log.Fatalf("cannot initialize global metrics instance: %s", err)
	}

	globalMetrics = Metrics{
		mon:      mon,
		counters: map[string]int64{},
	}
}

func InitMetricsServer(port int) {
	go func() {
		metricServer := http.NewServeMux()
		metricServer.Handle("/metrics", promhttp.Handler())
		log.Printf("Start listen metrics handler: %d", port)
		err := http.ListenAndServe(":"+strconv.Itoa(port), metricServer)
		if err != nil {
			log.Printf("failed listen metrics handler: %s", err)
		}
	}()
}

func RedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		c.Next()

		labels := []metrics.Label{{Name: "method", Value: c.Request.Method}, {Name: "route", Value: c.Request.RequestURI}, {Name: "status", Value: strconv.Itoa(c.Writer.Status())}}

		duration := time.Since(t)
		metrics.AddSampleWithLabels([]string{"request_duration"}, float32(duration.Milliseconds()), labels)
	}
}

func AddCounter(name string, delta int64) {
	globalMetrics.counters[name] += delta
	newVal := globalMetrics.counters[name]
	metrics.AddSample([]string{name}, float32(newVal))
}
