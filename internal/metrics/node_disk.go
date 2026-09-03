package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/probe"
)

var (
	descDiskReadsCompleted = prometheus.NewDesc(
		"node_disk_reads_completed_total", "Total successful reads.",
		[]string{"device"}, nil)
	descDiskReadBytes = prometheus.NewDesc(
		"node_disk_read_bytes_total", "Total bytes read successfully.",
		[]string{"device"}, nil)
	descDiskReadTime = prometheus.NewDesc(
		"node_disk_read_time_seconds_total", "Total seconds spent on all reads.",
		[]string{"device"}, nil)
	descDiskWritesCompleted = prometheus.NewDesc(
		"node_disk_writes_completed_total", "Total successful writes.",
		[]string{"device"}, nil)
	descDiskWrittenBytes = prometheus.NewDesc(
		"node_disk_written_bytes_total", "Total bytes written successfully.",
		[]string{"device"}, nil)
	descDiskWriteTime = prometheus.NewDesc(
		"node_disk_write_time_seconds_total", "Total seconds spent on all writes.",
		[]string{"device"}, nil)
	descDiskIONow = prometheus.NewDesc(
		"node_disk_io_now", "I/Os currently in progress.",
		[]string{"device"}, nil)
	descDiskIOTime = prometheus.NewDesc(
		"node_disk_io_time_seconds_total", "Total seconds spent doing I/O.",
		[]string{"device"}, nil)
)

func (c *Collector) collectDisk(ch chan<- prometheus.Metric, s *probe.Snapshot) {
	if !s.Has(probe.CapDiskIO) {
		return
	}
	cap := newFamilyCap("node_disk", c.dropped)
	for _, d := range s.DiskIO.Disks {
		if c.filters.DiskExcluded(d.Name) {
			continue
		}
		if !cap.allow() {
			return
		}
		dev := d.Name
		counter(ch, descDiskReadsCompleted, float64(d.Reads), dev)
		counter(ch, descDiskReadBytes, float64(d.ReadBytes), dev)
		// ms → seconds
		counter(ch, descDiskReadTime, float64(d.ReadTimeMS)/1000, dev)
		counter(ch, descDiskWritesCompleted, float64(d.Writes), dev)
		counter(ch, descDiskWrittenBytes, float64(d.WriteBytes), dev)
		counter(ch, descDiskWriteTime, float64(d.WriteTimeMS)/1000, dev)
		gauge(ch, descDiskIONow, float64(d.IOInProgress), dev)
		counter(ch, descDiskIOTime, float64(d.IOTimeMS)/1000, dev)
	}
}
