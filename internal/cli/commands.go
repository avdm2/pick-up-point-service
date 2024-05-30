package cli

const (
	help         = "help"
	addOrder     = "add"
	deleteOrder  = "delete"
	receiveOrder = "receive"
	getOrders    = "orders"
	createRefund = "refund"
	getRefunds   = "refunds"
)

type command struct {
	name        string
	description string
}
