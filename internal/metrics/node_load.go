package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/probe"
)

var (
	descLoad1  = prometheus.NewDesc("node_load1", "1m load average.", nil, nil)
	descLoad5  = prometheus.NewDesc("node_load5", "5m load average.", nil, nil)
	descLoad15 = prometheus.NewDesc("node_load15", "15m load average.", nil, nil)
)

func (c *Collector) collectLoad(ch chan<- prometheus.Metric, s *probe.Snapshot) {
	if !s.Has(probe.CapLoadAvg) {
		return
	}
	gauge(ch, descLoad1, s.LoadAvg.Load1)
	gauge(ch, descLoad5, s.LoadAvg.Load5)
	gauge(ch, descLoad15, s.LoadAvg.Load15)
}
