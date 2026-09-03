package probe

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/buildinfo"
	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/provider"
)

// SelfMetrics owns every pme_* series; metric name strings live here and
// in internal/metrics only.
type SelfMetrics struct {
	Skipped  prometheus.Counter
	Timeouts prometheus.Counter
	Duration *prometheus.GaugeVec
	Errors   *prometheus.CounterVec
	Dropped  *prometheus.CounterVec

	buildInfo *prometheus.Desc
	age       *prometheus.Desc
	pub       *Published
	abiMajor  string
}

func NewSelfMetrics(pub *Published) *SelfMetrics {
	return &SelfMetrics{
		Skipped: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pme_probe_skipped_total",
			Help: "Probe ticks skipped because the previous collect was still in flight.",
		}),
		Timeouts: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pme_probe_timeouts_total",
			Help: "Probe ticks that exceeded the collect timeout.",
		}),
		Duration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pme_provider_duration_seconds",
			Help: "Duration of the most recent collect_all call.",
		}, []string{"provider"}),
		Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "pme_provider_errors_total",
			Help: "Non-OK collect_all results.",
		}, []string{"provider"}),
		Dropped: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "pme_series_dropped_total",
			Help: "Series dropped because the per-family cap was exceeded.",
		}, []string{"family"}),
		buildInfo: prometheus.NewDesc("pme_build_info",
			"Build information; value is always 1.",
			[]string{"version", "commit", "goversion", "probeme_abi"}, nil),
		age: prometheus.NewDesc("pme_probe_age_seconds",
			"Age of the published snapshot at scrape time.", nil, nil),
		pub:      pub,
		abiMajor: strconv.Itoa(provider.ABIMajor),
	}
}

func (m *SelfMetrics) Collectors() []prometheus.Collector {
	return []prometheus.Collector{m.Skipped, m.Timeouts, m.Duration, m.Errors, m.Dropped, m}
}

func (m *SelfMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.buildInfo
	ch <- m.age
}

func (m *SelfMetrics) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(m.buildInfo, prometheus.GaugeValue, 1,
		buildinfo.Version, buildinfo.Commit, buildinfo.GoVersion(), m.abiMajor)
	if s := m.pub.Load(); s != nil {
		ch <- prometheus.MustNewConstMetric(m.age, prometheus.GaugeValue, time.Since(s.PublishedAt).Seconds())
	}
}
