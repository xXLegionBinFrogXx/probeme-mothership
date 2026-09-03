package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/probe"
)

var descThermalZoneTemp = prometheus.NewDesc(
	"node_thermal_zone_temp", "Thermal zone temperature in celsius.",
	[]string{"zone", "type"}, nil,
)

func (c *Collector) collectThermal(ch chan<- prometheus.Metric, s *probe.Snapshot) {
	if !s.Has(probe.CapThermal) {
		return
	}
	cap := newFamilyCap("node_thermal", c.dropped)
	for i, z := range s.Thermal.Zones {
		if !cap.allow() {
			return
		}
		// m°C → °C
		gauge(ch, descThermalZoneTemp, float64(z.TempMC)/1000, strconv.Itoa(i), z.Type)
	}
}
