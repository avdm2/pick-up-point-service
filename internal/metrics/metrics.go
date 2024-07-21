package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	ordersLabel = "orders"
)

var (
	addedOrders = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "added_orders",
		Help: "total number of added orders",
	}, []string{
		ordersLabel,
	})

	receivedOrders = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "received_orders",
		Help: "total number of received orders",
	}, []string{
		ordersLabel,
	})

	refundedOrders = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "refunded_orders",
		Help: "total number of refunded orders",
	}, []string{
		ordersLabel,
	})
)

func IncAddedOrders(cnt int) {
	addedOrders.With(prometheus.Labels{
		ordersLabel: "added",
	}).Add(float64(cnt))
}

func IncReceivedOrders(cnt int) {
	receivedOrders.With(prometheus.Labels{
		ordersLabel: "received",
	}).Add(float64(cnt))
}

func IncRefundedOrders(cnt int) {
	refundedOrders.With(prometheus.Labels{
		ordersLabel: "refunded",
	}).Add(float64(cnt))
}
