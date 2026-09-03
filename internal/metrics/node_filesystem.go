package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/probe"
)

var (
	descFSSize = prometheus.NewDesc(
		"node_filesystem_size_bytes", "Filesystem size in bytes.",
		[]string{"device", "mountpoint", "fstype"}, nil)
	descFSFree = prometheus.NewDesc(
		"node_filesystem_free_bytes", "Filesystem free space in bytes.",
		[]string{"device", "mountpoint", "fstype"}, nil)
	descFSAvail = prometheus.NewDesc(
		"node_filesystem_avail_bytes", "Filesystem space available to non-root users in bytes.",
		[]string{"device", "mountpoint", "fstype"}, nil)
	descFSFiles = prometheus.NewDesc(
		"node_filesystem_files", "Filesystem total file nodes.",
		[]string{"device", "mountpoint", "fstype"}, nil)
	descFSFilesFree = prometheus.NewDesc(
		"node_filesystem_files_free", "Filesystem total free file nodes.",
		[]string{"device", "mountpoint", "fstype"}, nil)
	descFSReadonly = prometheus.NewDesc(
		"node_filesystem_readonly", "Filesystem read-only status: 1 yes, 0 no.",
		[]string{"device", "mountpoint", "fstype"}, nil)
)

func (c *Collector) collectFilesystem(ch chan<- prometheus.Metric, s *probe.Snapshot) {
	if !s.Has(probe.CapFilesystem) {
		return
	}
	cap := newFamilyCap("node_filesystem", c.dropped)
	for _, m := range s.Filesystem.Mounts {
		if m.Flags&probe.MountSkipped != 0 {
			continue
		}
		if c.filters.MountExcluded(m.Mountpoint) || c.filters.FSTypeExcluded(m.FSType) {
			continue
		}
		if !cap.allow() {
			return
		}
		labels := []string{m.Device, m.Mountpoint, m.FSType}
		gauge(ch, descFSSize, float64(m.SizeBytes), labels...)
		gauge(ch, descFSFree, float64(m.FreeBytes), labels...)
		gauge(ch, descFSAvail, float64(m.AvailBytes), labels...)
		gauge(ch, descFSFiles, float64(m.Files), labels...)
		gauge(ch, descFSFilesFree, float64(m.FilesFree), labels...)
		ro := 0.0
		if m.Flags&probe.MountRO != 0 {
			ro = 1
		}
		gauge(ch, descFSReadonly, ro, labels...)
	}
}
