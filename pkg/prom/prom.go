package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "ipld_eth_state_snapshot"

	connSubsystem  = "connections"
	statsSubsystem = "stats"
)

var (
	metrics            bool
	lastValidatedBlock prometheus.Gauge
)

func Init() {
	metrics = true

	lastValidatedBlock = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "last_validated_block",
		Help:      "Last validated block number",
	})
}

// RegisterDBCollector create metric collector for given connection
func RegisterDBCollector(name string, db DBStatsGetter) {
	if metrics {
		prometheus.Register(NewDBStatsCollector(name, db))
	}
}

// SetLastValidatedBlock sets the last validated block number
func SetLastValidatedBlock(blockNumber float64) {
	if metrics {
		lastValidatedBlock.Set(blockNumber)
	}
}
