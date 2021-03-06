package app

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// promExport models an exported field
type promExport struct {
	name, help string
	typ        prometheus.ValueType
	value      func(m *HostMetrics, t *time.Time) float64

	// desc is dynamically build upon first use in Desc().
	desc *prometheus.Desc
}

// promExporter exports Prometheus metrics.
type promExporter []*promExport

func init() {
	prom := &promExporter{
		&promExport{
			name: "last_success",
			help: "duration since last success in seconds",
			typ:  prometheus.CounterValue,
			value: func(m *HostMetrics, t *time.Time) float64 {
				since := float64(-1)
				if m.SucceededAt != nil {
					since = float64(t.Sub(*m.SucceededAt) / time.Second)
				}
				return since
			},
		},
		&promExport{
			name: "last_duration",
			help: "duration of last successful run in seconds",
			typ:  prometheus.GaugeValue,
			value: func(m *HostMetrics, t *time.Time) float64 {
				dur := float64(-1)
				if m.SucceededAt != nil {
					dur = float64(m.SuccessDuration) / float64(time.Second)
				}
				return dur
			},
		},
		&promExport{
			name: "space_used",
			help: "total space used for backups in bytes",
			typ:  prometheus.GaugeValue,
			value: func(m *HostMetrics, t *time.Time) float64 {
				return float64(m.SpaceUsedTotal())
			},
		},
		&promExport{
			name: "space_used_by_snapshots",
			help: "space used by snapshots in bytes",
			typ:  prometheus.GaugeValue,
			value: func(m *HostMetrics, t *time.Time) float64 {
				return float64(m.SpaceUsedBySnapshots)
			},
		},
		&promExport{
			name: "space_used_by_dataset",
			help: "space used by dataset in bytes",
			typ:  prometheus.GaugeValue,
			value: func(m *HostMetrics, t *time.Time) float64 {
				return float64(m.SpaceUsedByDataset)
			},
		},
		&promExport{
			name: "space_used_by_children",
			help: "space used by children in bytes",
			typ:  prometheus.GaugeValue,
			value: func(m *HostMetrics, t *time.Time) float64 {
				return float64(m.SpaceUsedByChildren)
			},
		},
		&promExport{
			name: "space_used_by_reserved",
			help: "space reserved in bytes",
			typ:  prometheus.GaugeValue,
			value: func(m *HostMetrics, t *time.Time) float64 {
				return float64(m.SpaceUsedByRefReservation)
			},
		},
		&promExport{
			name: "compression",
			help: "compression ratio",
			typ:  prometheus.GaugeValue,
			value: func(m *HostMetrics, t *time.Time) float64 {
				return m.CompressionFactor
			},
		},
	}
	prometheus.MustRegister(prom)
}

var hostLabels = []string{"host"}

func (f *promExport) Desc() *prometheus.Desc {
	if f.desc == nil {
		name := prometheus.BuildFQName("zackup", "", f.name)
		f.desc = prometheus.NewDesc(name, f.help, hostLabels, nil)
	}
	return f.desc
}

// Describe implements the prometheus.Collector interface
func (e promExporter) Describe(c chan<- *prometheus.Desc) {
	for _, f := range e {
		c <- f.Desc()
	}
}

// Collect implements the prometheus.Collector interface
func (e promExporter) Collect(c chan<- prometheus.Metric) {
	now := time.Now()

	for _, m := range state.export() {
		for _, f := range e {
			val := f.value(&m, &now)
			c <- prometheus.MustNewConstMetric(f.desc, f.typ, val, m.Host)
		}
	}
}
