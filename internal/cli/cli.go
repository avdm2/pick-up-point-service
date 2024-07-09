package cli

import (
	"bufio"
	"errors"
	"fmt"
	"homework-1/internal/models"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	dateLayout = "02-01-2006"
)

var (
	errIncorrectArgAmount = errors.New("incorrect amount of arguments")
	errIncorrectId        = errors.New("empty or non-positive order or customer id")
	errNegativeWeight     = errors.New("weight can not be negative")
	errNegativeCost       = errors.New("cost can not be negative")
)

type Module interface {
	AddOrder(orderId models.ID, customerId models.ID, expirationDate time.Time, pack models.PackageType, weight models.Kilo, cost models.Rub) error
	ReturnOrder(id models.ID) error
	ReceiveOrders(ordersId []models.ID) ([]models.Order, error)
	GetOrders(customerId models.ID, n int) ([]models.Order, error)
	RefundOrder(customerId models.ID, orderId models.ID) error
	GetRefunds(page int, limit int) ([]models.Order, error)
}

type Deps struct {
	Module Module
}

type CLI struct {
	Deps
	wg      sync.WaitGroup
	Tasks   chan string
	Workers int
}

func NewCLI(d Deps) *CLI {
	return &CLI{
		Deps:    d,
		Tasks:   make(chan string),
		Workers: 2,
	}
}

func (c *CLI) Run() error {

	for i := 0; i < c.Workers; i++ {
		c.wg.Add(1)
		fmt.Printf("[* w%d *] Запуск воркера\n", i)
		go c.worker(i)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("[>>] Введите команду:")
		cmd, errRead := reader.ReadString('\n')
		if errRead != nil {
			return fmt.Errorf("cli.Run error: %w", errRead)
		}

		cmd = strings.TrimRight(cmd, "\n")
		if cmd == "exit" {
			break
		}

		c.Tasks <- cmd
	}

	close(c.Tasks)
	c.wg.Wait()
	return nil
}

func (c *CLI) worker(id int) {
	defer c.wg.Done()
	for cmd := range c.Tasks {
		fmt.Printf("[* w%d *] Обработка команды [%s]\n", id, cmd)
		if errCmd := c.handleCommand(cmd); errCmd != nil {
			fmt.Printf("cli.Run error: %s\n", errCmd)
		}
		fmt.Printf("[* w%d *] Команда [%s] обработана\n", id, cmd)
	}
}

func (c *CLI) handleCommand(command string) error {

	// Имитация длительной работы
	time.Sleep(5 * time.Second)

	arguments := strings.Split(command, " ")
	switch arguments[0] {
	case help:
		return c.help()
	case addOrder:
		return c.addOrder(arguments[1:])
	case returnOrder:
		return c.returnOrder(arguments[1:])
	case receiveOrder:
		return c.receiveOrder(arguments[1:])
	case getOrders:
		return c.getOrders(arguments[1:])
	case createRefund:
		return c.createRefund(arguments[1:])
	case getRefunds:
		return c.getRefunds(arguments[1:])
	default:
		return c.unknownCommand()
	}
}

func (c *CLI) help() error {
	fmt.Println("Список доступных команд: ")

	commands := commandList()
	for _, com := range commands {
		fmt.Printf("%s: %s\n", com.name, com.description)
	}

	return nil
}

func (c *CLI) unknownCommand() error {
	fmt.Println("Введенная команда не найдена. Проверьте количество аргументов или используйте другие команды " +
		"(для вывода списка команд воспользуйтесь командой \"help\")")

	return nil
}

// addOrder --orderId=1 --customerId=1 --expirationTime=01-01-2024 --packageType=box --weight=1 --cost=1
func (c *CLI) addOrder(args []string) error {
	if len(args) != 6 {
		return errIncorrectArgAmount
	}

	expirationTime := args[2]
	pack := models.PackageType(args[3])

	orderIdInt, errParse := strconv.Atoi(args[0])
	if errParse != nil {
		return fmt.Errorf("cli.addOrder error: %w", errParse)
	}
	customerIdInt, errParse := strconv.Atoi(args[1])
	if errParse != nil {
		return fmt.Errorf("cli.addOrder error: %w", errParse)
	}

	orderId := models.ID(orderIdInt)
	customerId := models.ID(customerIdInt)
	if orderId <= 0 || customerId <= 0 {
		return fmt.Errorf("cli.addOrder error: %w", errIncorrectId)
	}

	date, errDate := time.Parse(dateLayout, expirationTime)
	if errDate != nil {
		return fmt.Errorf("cli.addOrder error: %w", errDate)
	}

	weightFloat, errParse := strconv.ParseFloat(args[4], 32)
	if errParse != nil {
		return fmt.Errorf("cli.addOrder error: %w", errParse)
	}

	weight := models.Kilo(weightFloat)
	if weight < 0 {
		return fmt.Errorf("cli.addOrder error: %w", errNegativeWeight)
	}

	costInt, errParse := strconv.Atoi(args[5])
	if errParse != nil {
		return fmt.Errorf("cli.addOrder error: %w", errParse)
	}

	cost := models.Rub(costInt)
	if cost < 0 {
		return fmt.Errorf("cli.addOrder error: %w", errNegativeCost)
	}

	if errAdd := c.Module.AddOrder(orderId, customerId, date, pack, weight, cost); errAdd != nil {
		return fmt.Errorf("cli.addOrder error: %w", errAdd)
	}

	fmt.Printf("Заказ %d добавлен!\n", orderId)
	return nil
}

// returnOrder --orderId=1
func (c *CLI) returnOrder(args []string) error {
	if len(args) != 1 {
		return errIncorrectArgAmount
	}

	orderIdInt, errParse := strconv.Atoi(args[0])
	if errParse != nil {
		return fmt.Errorf("cli.returnOrder error: %w", errParse)
	}

	orderId := models.ID(orderIdInt)

	if orderId <= 0 {
		return fmt.Errorf("cli.returnOrder error: %w", errIncorrectId)
	}

	if errReturn := c.Module.ReturnOrder(orderId); errReturn != nil {
		return fmt.Errorf("cli.returnOrder error: %w", errReturn)
	}

	fmt.Printf("Заказ %d возвращен!\n", orderId)
	return nil
}

// receiveOrder --orders=1,2,3,4,5
func (c *CLI) receiveOrder(args []string) error {
	if len(args) != 1 {
		return errIncorrectArgAmount
	}

	orderIds, errParseId := parseIDs(args[0])
	if errParseId != nil {
		return fmt.Errorf("cli.receiveOrder error: %w", errParseId)
	}

	orders, errReceiving := c.Module.ReceiveOrders(orderIds)
	if errReceiving != nil {
		return fmt.Errorf("cli.receiveOrder error: %w", errReceiving)
	}

	if len(orders) == 0 {
		fmt.Printf("Заказов c номерами [%s] нет!\n", args[0])
		return nil
	}

	fmt.Println("Список полученных заказов:")
	for _, order := range orders {
		fmt.Println(order)
	}

	return nil
}

// getOrders --customerId=1 --n=1
func (c *CLI) getOrders(args []string) error {
	if len(args) != 2 {
		return errIncorrectArgAmount
	}

	customerIdInt, errParse := strconv.Atoi(args[0])
	if errParse != nil {
		return fmt.Errorf("cli.getOrders error: %w", errParse)
	}

	n, errParse := strconv.Atoi(args[1])
	if errParse != nil {
		return fmt.Errorf("cli.getOrders error: %w", errParse)
	}

	customerId := models.ID(customerIdInt)

	if customerId <= 0 {
		return fmt.Errorf("cli.getOrders error: %w", errIncorrectId)
	}

	orders, errGet := c.Module.GetOrders(customerId, n)
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

// createRefund --orderId=1 --customerId=1
func (c *CLI) createRefund(args []string) error {
	if len(args) != 2 {
		return errIncorrectArgAmount
	}

	orderIdInt, errParse := strconv.Atoi(args[0])
	if errParse != nil {
		return fmt.Errorf("cli.createRefund error: %w", errParse)
	}

	customerIdInt, errParse := strconv.Atoi(args[1])
	if errParse != nil {
		return fmt.Errorf("cli.createRefund error: %w", errParse)
	}

	orderId := models.ID(orderIdInt)
	customerId := models.ID(customerIdInt)

	if orderId <= 0 || customerId <= 0 {
		return fmt.Errorf("cli.createRefund error: %w", errIncorrectId)
	}

	if errRefund := c.Module.RefundOrder(customerId, orderId); errRefund != nil {
		return fmt.Errorf("cli.createRefund error: %w", errRefund)
	}

	fmt.Printf("Возврат на заказ %d создан!\n", orderId)
	return nil
}

// getRefunds --page=1 --limit=1
func (c *CLI) getRefunds(args []string) error {
	if len(args) != 2 {
		return errIncorrectArgAmount
	}

	page, errParse := strconv.Atoi(args[0])
	if errParse != nil {
		return fmt.Errorf("cli.getRefunds error: %w", errParse)
	}

	limit, errParse := strconv.Atoi(args[1])
	if errParse != nil {
		return fmt.Errorf("cli.getRefunds error: %w", errParse)
	}

	orders, errGetRefunds := c.Module.GetRefunds(page, limit)
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
		{
			name:        exit,
			description: "Выход",
		},
	}
}
