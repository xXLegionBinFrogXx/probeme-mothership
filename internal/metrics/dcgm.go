package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/probe"
)

// Scrape-compatible with dcgm-exporter (dashboard 12239); never export
// FB/VRAM fields.
var (
	descGPUTemp = prometheus.NewDesc(
		"DCGM_FI_DEV_GPU_TEMP", "GPU temperature (in C).",
		gpuLabels, nil)
	descGPUPower = prometheus.NewDesc(
		"DCGM_FI_DEV_POWER_USAGE", "GPU power usage (in W).",
		gpuLabels, nil)
	descGPUSMClock = prometheus.NewDesc(
		"DCGM_FI_DEV_SM_CLOCK", "SM clock frequency (in MHz).",
		gpuLabels, nil)
	descGPUUtil = prometheus.NewDesc(
		"DCGM_FI_DEV_GPU_UTIL", "GPU utilization (in %).",
		gpuLabels, nil)
	descGPUPstate = prometheus.NewDesc(
		"DCGM_FI_DEV_PSTATE", "PState of the GPU (0 to 15, 0 is maximum performance).",
		gpuLabels, nil)
)

var gpuLabels = []string{"gpu", "UUID", "device", "modelName"}

func (c *Collector) collectGPU(ch chan<- prometheus.Metric, s *probe.Snapshot) {
	if !s.Has(probe.CapGPU) {
		return
	}
	cap := newFamilyCap("dcgm", c.dropped)
	for i, g := range s.GPU.GPUs {
		if !cap.allow() {
			return
		}
		idx := strconv.Itoa(i)
		labels := []string{idx, g.UUID, "nvidia" + idx, g.Name}
		gauge(ch, descGPUTemp, float64(g.TempC), labels...)
		// mW → W
		gauge(ch, descGPUPower, float64(g.PowerMW)/1000, labels...)
		gauge(ch, descGPUSMClock, float64(g.SMClockMHz), labels...)
		gauge(ch, descGPUUtil, float64(g.UtilPct), labels...)
		gauge(ch, descGPUPstate, float64(g.Pstate), labels...)
	}
}
