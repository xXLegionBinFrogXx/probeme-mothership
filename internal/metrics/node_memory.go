package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/probe"
)

var (
	descMemTotal = prometheus.NewDesc(
		"node_memory_MemTotal_bytes", "Total memory in bytes.", nil, nil)
	descMemFree = prometheus.NewDesc(
		"node_memory_MemFree_bytes", "Free memory in bytes.", nil, nil)
	descMemAvailable = prometheus.NewDesc(
		"node_memory_MemAvailable_bytes", "Available memory in bytes.", nil, nil)
	descMemBuffers = prometheus.NewDesc(
		"node_memory_Buffers_bytes", "Buffer memory in bytes.", nil, nil)
	descMemCached = prometheus.NewDesc(
		"node_memory_Cached_bytes", "Page cache memory in bytes.", nil, nil)
	descMemSwapTotal = prometheus.NewDesc(
		"node_memory_SwapTotal_bytes", "Total swap in bytes.", nil, nil)
	descMemSwapFree = prometheus.NewDesc(
		"node_memory_SwapFree_bytes", "Free swap in bytes.", nil, nil)
)

func (c *Collector) collectMemory(ch chan<- prometheus.Metric, s *probe.Snapshot) {
	if !s.Has(probe.CapMemory) {
		return
	}
	m := s.Memory
	gauge(ch, descMemTotal, float64(m.Total))
	gauge(ch, descMemFree, float64(m.Free))
	gauge(ch, descMemAvailable, float64(m.Available))
	gauge(ch, descMemBuffers, float64(m.Buffers))
	gauge(ch, descMemCached, float64(m.Cached))
	gauge(ch, descMemSwapTotal, float64(m.SwapTotal))
	gauge(ch, descMemSwapFree, float64(m.SwapFree))
}
