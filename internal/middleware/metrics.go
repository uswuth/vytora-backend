package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var (
	httpReqDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	httpReqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.Register(httpReqDur)
	prometheus.Register(httpReqTotal)
}

func MetricsMiddleware(c *fiber.Ctx) error {
	start := time.Now()
	path := c.Route().Path
	if path == "" {
		path = c.Path()
	}

	err := c.Next()

	status := strconv.Itoa(c.Response().StatusCode())
	duration := time.Since(start).Seconds()

	httpReqDur.WithLabelValues(c.Method(), path, status).Observe(duration)
	httpReqTotal.WithLabelValues(c.Method(), path, status).Inc()

	return err
}

func MetricsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		gatherer := prometheus.DefaultGatherer
		mfs, err := gatherer.Gather()
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		for _, mf := range mfs {
			if len(mf.GetMetric()) == 0 {
				continue
			}
			if err := encodeMetricFamily(c, mf); err != nil {
				return err
			}
		}
		return nil
	}
}

func encodeMetricFamily(c *fiber.Ctx, mf *dto.MetricFamily) error {
	name := mf.GetName()
	help := mf.GetHelp()
	mtype := mf.GetType()

	if help != "" {
		c.SendString("# HELP " + name + " " + help + "\n")
	}
	c.SendString("# TYPE " + name + " " + mtype.String() + "\n")

	for _, m := range mf.GetMetric() {
		labels := ""
		for _, lp := range m.GetLabel() {
			labels += lp.GetName() + "=\"" + lp.GetValue() + "\","
		}
		if len(labels) > 0 {
			labels = "{" + labels[:len(labels)-1] + "}"
		}
		val := ""
		switch mtype {
		case dto.MetricType_COUNTER:
			val = formatFloat(m.GetCounter().GetValue())
		case dto.MetricType_GAUGE:
			val = formatFloat(m.GetGauge().GetValue())
		case dto.MetricType_HISTOGRAM:
			h := m.GetHistogram()
			buckets := h.GetBucket()
			sampleCount := h.GetSampleCount()
			sampleSum := h.GetSampleSum()
			for _, b := range buckets {
				le := formatFloat(b.GetUpperBound())
				c.SendString(name + "_bucket" + labels + "{le=\"" + le + "\"} " + strconv.FormatUint(b.GetCumulativeCount(), 10) + "\n")
			}
			c.SendString(name + "_bucket" + labels + "{le=\"+Inf\"} " + strconv.FormatUint(sampleCount, 10) + "\n")
			c.SendString(name + "_sum" + labels + " " + formatFloat(sampleSum) + "\n")
			c.SendString(name + "_count" + labels + " " + strconv.FormatUint(sampleCount, 10) + "\n")
			return nil
		default:
			continue
		}
		c.SendString(name + labels + " " + val + "\n")
	}
	return nil
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}