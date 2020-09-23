package queue

import (
	"github.com/prometheus/client_golang/prometheus"
)

var queueTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "txmsg_queue_total",
		Help: "Number of total",
	},
	[]string{"name"},
)

var queueSize = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "txmsg_queue_size",
		Help: "Size of queue",
	},
	[]string{"name"},
)

var queueDroppedTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "txmsg_queue_dropped_total",
		Help: "Number of dropped",
	},
	[]string{"name"},
)

func init() {
	prometheus.MustRegister(queueTotal)
	prometheus.MustRegister(queueSize)
	prometheus.MustRegister(queueDroppedTotal)
}