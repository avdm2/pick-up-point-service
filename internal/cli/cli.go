package cli

import (
	"errors"
	"flag"
	"fmt"
	"homework-1/internal/models"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	dateLayout = "02-01-2006"
)

var (
	errEmptyArgs   = errors.New("empty args")
	errIncorrectId = errors.New("empty or non-positive order or customer id")
)

type Module interface {
	Add(order models.Order) error
	Return(id models.ID) error
	Receive(ordersId []models.ID) ([]models.Order, error)
	Orders(customerId models.ID, n int) ([]models.Order, error)
	Refund(customerId models.ID, orderId models.ID) error
	Refunds(page int, limit int) ([]models.Order, error)
}

type Deps struct {
	Module Module
}

type CLI struct {
	Deps
}

func NewCLI(d Deps) CLI {
	return CLI{
		Deps: d,
	}
}

func (c CLI) Run() error {
	args := os.Args[1:]
	if len(args) == 0 {
		return fmt.Errorf("cli.Run() error: %w", errEmptyArgs)
	}

	com := args[0]
	switch com {
	case help:
		return c.help()
	case addOrder:
		return c.addOrder(args[1:])
	case returnOrder:
		return c.returnOrder(args[1:])
	case receiveOrder:
		return c.receiveOrder(args[1:])
	case getOrders:
		return c.getOrders(args[1:])
	case createRefund:
		return c.createRefund(args[1:])
	case getRefunds:
		return c.getRefunds(args[1:])
	default:
		return c.unknownCommand()
	}
}

func commandList() []command {
	return []command{
		{
			name:        help,
			description: "Справка",
		},
		{
			name:        addOrder,
			description: "Добавить заказ",
		},
		{
			name:        returnOrder,
			description: "Удалить заказ",
		},
		{
			name:        receiveOrder,
			description: "Получить заказ",
		},
		{
			name:        getOrders,
			description: "Получить список заказов",
		},
		{
			name:        createRefund,
			description: "Создать запрос на возврат",
		},
		{
			name:        getRefunds,
			description: "Получить список возвратов",
		},
	}
}

func (c CLI) help() error {
	fmt.Println("Список доступных команд: ")

	commands := commandList()
	for _, com := range commands {
		fmt.Printf("%s: %s\n", com.name, com.description)
	}

	return nil
}

func (c CLI) unknownCommand() error {
	fmt.Println("Введенная команда не найдена. Проверьте количество аргументов или используйте другие команды " +
		"(для вывода списка команд воспользуйтесь командой \"help\")")

	return nil
}

func (c CLI) addOrder(args []string) error {
	var orderId, customerId models.ID
	var expirationTime string

	fs := flag.NewFlagSet(addOrder, flag.ContinueOnError)
	fs.Int64Var((*int64)(&orderId), "orderId", -1, "use --orderId=1")
	fs.Int64Var((*int64)(&customerId), "customerId", -1, "use --customerId=1")
	fs.StringVar(&expirationTime, "expirationTime", "01-01-1990", "use --expirationTime=01-01-2024")

	if errFs := fs.Parse(args); errFs != nil {
		return fmt.Errorf("cli.addOrder error: %w", errFs)
	}

	if orderId <= 0 || customerId <= 0 {
		return fmt.Errorf("cli.addOrder error: %w", errIncorrectId)
	}

	date, errDate := time.Parse(dateLayout, expirationTime)
	if errDate != nil {
		return fmt.Errorf("cli.addOrder error: %w", errDate)
	}

	order := models.NewOrder(orderId, customerId, date)

	if errAdd := c.Module.Add(*order); errAdd != nil {
		return fmt.Errorf("cli.addOrder error: %w", errAdd)
	}

	fmt.Printf("Заказ %d добавлен!\n", orderId)
	return nil
}

func (c CLI) returnOrder(args []string) error {
	var orderId models.ID

	fs := flag.NewFlagSet(returnOrder, flag.ContinueOnError)
	fs.Int64Var((*int64)(&orderId), "orderId", -1, "use --orderId=1")

	if errFs := fs.Parse(args); errFs != nil {
		return fmt.Errorf("cli.returnOrder error: %w", errFs)
	}

	if orderId <= 0 {
		return fmt.Errorf("cli.returnOrder error: %w", errIncorrectId)
	}

	if errReturn := c.Module.Return(orderId); errReturn != nil {
		return fmt.Errorf("cli.returnOrder error: %w", errReturn)
	}

	fmt.Printf("Заказ %d удален!\n", orderId)
	return nil
}

func (c CLI) receiveOrder(args []string) error {
	var ordersStr string

	fs := flag.NewFlagSet(receiveOrder, flag.ContinueOnError)
	fs.StringVar(&ordersStr, "orders", "0", "use --orders=1,2,3,4,5")

	if errFs := fs.Parse(args); errFs != nil {
		return fmt.Errorf("cli.receiveOrder error: %w", errFs)
	}

	orderIds, errParseId := parseIDs(ordersStr)
	if errParseId != nil {
		return fmt.Errorf("cli.receiveOrder error: %w", errParseId)
	}

	orders, errReceiving := c.Module.Receive(orderIds)
	if errReceiving != nil {
		return fmt.Errorf("cli.receiveOrder error: %w", errReceiving)
	}

	if len(orders) == 0 {
		fmt.Printf("Заказов c номерами [%s] нет!\n", ordersStr)
		return nil
	}

	fmt.Println("Список полученных заказов:")
	for _, order := range orders {
		fmt.Println(order)
	}

	return nil
}

func (c CLI) getOrders(args []string) error {
	var customerId models.ID
	var n int

	fs := flag.NewFlagSet(getOrders, flag.ContinueOnError)
	fs.Int64Var((*int64)(&customerId), "customerId", -1, "use --customerId=1")
	fs.IntVar(&n, "n", -1, "use --n=1")

	if errFs := fs.Parse(args); errFs != nil {
		return fmt.Errorf("cli.getOrders error: %w", errFs)
	}

	if customerId <= 0 {
		return fmt.Errorf("cli.getOrders error: %w", errIncorrectId)
	}

	orders, errGet := c.Module.Orders(customerId, n)
	if errGet != nil {
		return fmt.Errorf("cli.getOrders error: %w", errGet)
	}

	if len(orders) == 0 {
		fmt.Printf("Заказов у пользователя %d нет!\n", customerId)
		return nil
	}

	fmt.Printf("Список заказов пользователя %d:\n", customerId)
	for _, order := range orders {
		fmt.Println(order)
	}

	return nil
}

func (c CLI) createRefund(args []string) error {
	var orderId, customerId models.ID

	fs := flag.NewFlagSet(createRefund, flag.ContinueOnError)
	fs.Int64Var((*int64)(&orderId), "orderId", -1, "use --orderId=1")
	fs.Int64Var((*int64)(&customerId), "customerId", -1, "use --customerId=1")

	if errFs := fs.Parse(args); errFs != nil {
		return fmt.Errorf("cli.createRefund error: %w", errFs)
	}

	if orderId <= 0 || customerId <= 0 {
		return fmt.Errorf("cli.createRefund error: %w", errIncorrectId)
	}

	if errRefund := c.Module.Refund(customerId, orderId); errRefund != nil {
		return fmt.Errorf("cli.createRefund error: %w", errRefund)
	}

	fmt.Printf("Возврат на заказ %d создан!\n", orderId)
	return nil
}

func (c CLI) getRefunds(args []string) error {
	var page, limit int

	fs := flag.NewFlagSet(getRefunds, flag.ContinueOnError)
	fs.IntVar(&page, "page", 0, "use --page=1")
	fs.IntVar(&limit, "limit", 0, "use --limit=1")

	if errFs := fs.Parse(args); errFs != nil {
		return fmt.Errorf("cli.getRefunds error: %w", errFs)
	}

	orders, errGetRefunds := c.Module.Refunds(page, limit)
	if errGetRefunds != nil {
		return fmt.Errorf("cli.getRefunds error: %w", errGetRefunds)
	}

	if len(orders) == 0 {
		fmt.Println("Возвратов нет")
		return nil
	}

	fmt.Println("Список возвратов:")
	for _, order := range orders {
		fmt.Println(order)
	}

	return nil
}

func parseIDs(idsStr string) ([]models.ID, error) {
	if idsStr == "" {
		return nil, errIncorrectId
	}

	strIds := strings.Split(idsStr, ",")
	var ids []models.ID
	for _, strId := range strIds {
		id, errParse := strconv.ParseInt(strings.TrimSpace(strId), 10, 64)
		if errParse != nil {
			return nil, errParse
		}
		ids = append(ids, models.ID(id))
	}

	return ids, nil
}
