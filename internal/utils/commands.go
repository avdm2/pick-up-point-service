package utils

const (
	helpCommand         = "help"
	addOrderCommand     = "add"
	returnOrderCommand  = "return"
	receiveOrderCommand = "receive"
	getOrdersCommand    = "orders"
	createRefundCommand = "refund"
	getRefundsCommand   = "refunds"
)

type command struct {
	name        string
	description string
}
