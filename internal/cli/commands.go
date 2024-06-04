package cli

const (
	help         = "help"
	addOrder     = "add"
	returnOrder  = "return"
	receiveOrder = "receive"
	getOrders    = "orders"
	createRefund = "refund"
	getRefunds   = "refunds"
)

type command struct {
	name        string
	description string
}
