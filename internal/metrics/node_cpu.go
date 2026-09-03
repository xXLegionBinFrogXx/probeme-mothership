package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/probe"
)

var descCPUSeconds = prometheus.NewDesc(
	"node_cpu_seconds_total",
	"Seconds the CPUs spent in each mode.",
	[]string{"cpu", "mode"}, nil,
)

func (c *Collector) collectCPU(ch chan<- prometheus.Metric, s *probe.Snapshot) {
	if !s.Has(probe.CapCPU) {
		return
	}
	cap := newFamilyCap("node_cpu", c.dropped)

	// ticks → seconds; clk_tck comes from the provider, fall back to 100.
	clk := float64(s.CPU.ClkTck)
	if clk == 0 {
		clk = 100
	}

	for i, core := range s.CPU.Cores {
		if !cap.allow() {
			return
		}
		cpu := strconv.Itoa(i)
		counter(ch, descCPUSeconds, float64(core.User)/clk, cpu, "user")
		counter(ch, descCPUSeconds, float64(core.Nice)/clk, cpu, "nice")
		counter(ch, descCPUSeconds, float64(core.System)/clk, cpu, "system")
		counter(ch, descCPUSeconds, float64(core.Idle)/clk, cpu, "idle")
		counter(ch, descCPUSeconds, float64(core.Iowait)/clk, cpu, "iowait")
		counter(ch, descCPUSeconds, float64(core.IRQ)/clk, cpu, "irq")
		counter(ch, descCPUSeconds, float64(core.SoftIRQ)/clk, cpu, "softirq")
		counter(ch, descCPUSeconds, float64(core.Steal)/clk, cpu, "steal")
	}
}
