package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/probe"
)

var (
	descNetRxBytes = prometheus.NewDesc(
		"node_network_receive_bytes_total", "Received bytes.",
		[]string{"device"}, nil)
	descNetRxPackets = prometheus.NewDesc(
		"node_network_receive_packets_total", "Received packets.",
		[]string{"device"}, nil)
	descNetRxErrs = prometheus.NewDesc(
		"node_network_receive_errs_total", "Receive errors.",
		[]string{"device"}, nil)
	descNetRxDrop = prometheus.NewDesc(
		"node_network_receive_drop_total", "Received dropped packets.",
		[]string{"device"}, nil)
	descNetTxBytes = prometheus.NewDesc(
		"node_network_transmit_bytes_total", "Transmitted bytes.",
		[]string{"device"}, nil)
	descNetTxPackets = prometheus.NewDesc(
		"node_network_transmit_packets_total", "Transmitted packets.",
		[]string{"device"}, nil)
	descNetTxErrs = prometheus.NewDesc(
		"node_network_transmit_errs_total", "Transmit errors.",
		[]string{"device"}, nil)
	descNetTxDrop = prometheus.NewDesc(
		"node_network_transmit_drop_total", "Transmit dropped packets.",
		[]string{"device"}, nil)
)

func (c *Collector) collectNetDev(ch chan<- prometheus.Metric, s *probe.Snapshot) {
	if !s.Has(probe.CapNetDev) {
		return
	}
	cap := newFamilyCap("node_netdev", c.dropped)
	for _, f := range s.NetDev.Ifaces {
		if c.filters.NetDeviceExcluded(f.Name) {
			continue
		}
		if !cap.allow() {
			return
		}
		dev := f.Name
		counter(ch, descNetRxBytes, float64(f.RxBytes), dev)
		counter(ch, descNetRxPackets, float64(f.RxPackets), dev)
		counter(ch, descNetRxErrs, float64(f.RxErrs), dev)
		counter(ch, descNetRxDrop, float64(f.RxDrop), dev)
		counter(ch, descNetTxBytes, float64(f.TxBytes), dev)
		counter(ch, descNetTxPackets, float64(f.TxPackets), dev)
		counter(ch, descNetTxErrs, float64(f.TxErrs), dev)
		counter(ch, descNetTxDrop, float64(f.TxDrop), dev)
	}
}
