package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/probe"
)

type Collector struct {
	pub     *probe.Published
	filters Filters
	dropped *prometheus.CounterVec
}

func NewCollector(pub *probe.Published, filters Filters, dropped *prometheus.CounterVec) *Collector {
	return &Collector{pub: pub, filters: filters, dropped: dropped}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- descCPUSeconds
	ch <- descMemTotal
	ch <- descMemFree
	ch <- descMemAvailable
	ch <- descMemBuffers
	ch <- descMemCached
	ch <- descMemSwapTotal
	ch <- descMemSwapFree
	ch <- descLoad1
	ch <- descLoad5
	ch <- descLoad15
	ch <- descBootTime
	ch <- descFSSize
	ch <- descFSFree
	ch <- descFSAvail
	ch <- descFSFiles
	ch <- descFSFilesFree
	ch <- descFSReadonly
	ch <- descDiskReadsCompleted
	ch <- descDiskReadBytes
	ch <- descDiskReadTime
	ch <- descDiskWritesCompleted
	ch <- descDiskWrittenBytes
	ch <- descDiskWriteTime
	ch <- descDiskIONow
	ch <- descDiskIOTime
	ch <- descNetRxBytes
	ch <- descNetRxPackets
	ch <- descNetRxErrs
	ch <- descNetRxDrop
	ch <- descNetTxBytes
	ch <- descNetTxPackets
	ch <- descNetTxErrs
	ch <- descNetTxDrop
	ch <- descThermalZoneTemp
	ch <- descGPUTemp
	ch <- descGPUPower
	ch <- descGPUSMClock
	ch <- descGPUUtil
	ch <- descGPUPstate
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	s := c.pub.Load()
	if s == nil {
		return
	}
	c.collectCPU(ch, s)
	c.collectMemory(ch, s)
	c.collectLoad(ch, s)
	c.collectUptime(ch, s)
	c.collectFilesystem(ch, s)
	c.collectDisk(ch, s)
	c.collectNetDev(ch, s)
	c.collectThermal(ch, s)
	c.collectGPU(ch, s)
}

func gauge(ch chan<- prometheus.Metric, d *prometheus.Desc, v float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(d, prometheus.GaugeValue, v, labels...)
}

func counter(ch chan<- prometheus.Metric, d *prometheus.Desc, v float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(d, prometheus.CounterValue, v, labels...)
}
