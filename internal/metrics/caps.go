package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// MaxSeriesPerFamily caps entries per family; excess bumps
// pme_series_dropped_total{family}.
const MaxSeriesPerFamily = 256

type familyCap struct {
	family    string
	remaining int
	dropped   *prometheus.CounterVec
}

func newFamilyCap(family string, dropped *prometheus.CounterVec) *familyCap {
	return &familyCap{family: family, remaining: MaxSeriesPerFamily, dropped: dropped}
}

func (c *familyCap) allow() bool {
	if c.remaining <= 0 {
		c.dropped.WithLabelValues(c.family).Inc()
		return false
	}
	c.remaining--
	return true
}
