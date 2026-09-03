package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/probe"
)

var descBootTime = prometheus.NewDesc(
	"node_boot_time_seconds",
	"Node boot time, in unixtime.", nil, nil,
)

func (c *Collector) collectUptime(ch chan<- prometheus.Metric, s *probe.Snapshot) {
	if !s.Has(probe.CapUptime) {
		return
	}
	gauge(ch, descBootTime, float64(s.Uptime.BootTimeUnixS))
}
