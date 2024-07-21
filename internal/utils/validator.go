package utils

import (
	"errors"
	"fmt"
	"homework-1/pkg/api/proto/orders_grpc/v1/orders_grpc/v1"
	"strconv"
	"strings"
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

func HandleCommand(command string) (interface{}, error) {

	arguments := strings.Split(command, " ")
	switch arguments[0] {
	case helpCommand:
		return nil, help()
	case addOrderCommand:
		req, err := addOrder(arguments[1:])
		if err != nil {
			return nil, fmt.Errorf("utils.HandleCommand error: %w\n", err)
		}
		return req, nil
	case returnOrderCommand:
		req, err := returnOrder(arguments[1:])
		if err != nil {
			return nil, fmt.Errorf("utils.HandleCommand error: %w\n", err)
		}
		return req, nil
	case receiveOrderCommand:
		req, err := receiveOrder(arguments[1:])
		if err != nil {
			return nil, fmt.Errorf("utils.HandleCommand error: %w\n", err)
		}
		return req, nil
	case getOrdersCommand:
		req, err := getOrders(arguments[1:])
		if err != nil {
			return nil, fmt.Errorf("utils.HandleCommand error: %w\n", err)
		}
		return req, nil
	case createRefundCommand:
		req, err := createRefund(arguments[1:])
		if err != nil {
			return nil, fmt.Errorf("utils.HandleCommand error: %w\n", err)
		}
		return req, nil
	case getRefundsCommand:
		req, err := getRefunds(arguments[1:])
		if err != nil {
			return nil, fmt.Errorf("utils.HandleCommand error: %w\n", err)
		}
		return req, nil
	default:
		return nil, unknownCommand()
	}
}

func help() error {
	fmt.Println("Список доступных команд: ")

	commands := commandList()
	for _, com := range commands {
		fmt.Printf("%s: %s\n", com.name, com.description)
	}

	return nil
}

func unknownCommand() error {
	fmt.Println("Введенная команда не найдена. Проверьте количество аргументов или используйте другие команды " +
		"(для вывода списка команд воспользуйтесь командой \"help\")")

	return nil
}

// addOrder --orderId=1 --customerId=1 --expirationTime=01-01-2024 --packageType=box --weight=1 --cost=1
func addOrder(args []string) (*orders_grpc.AddOrderRequest, error) {
	if len(args) != 6 {
		return nil, errIncorrectArgAmount
	}

	expirationTime := args[2]
	pack := args[3]

	orderIdInt, errParse := strconv.ParseInt(args[0], 10, 64)
	if errParse != nil {
		return nil, fmt.Errorf("cli.addOrder error: %w", errParse)
	}
	customerIdInt, errParse := strconv.ParseInt(args[1], 10, 64)
	if errParse != nil {
		return nil, fmt.Errorf("cli.addOrder error: %w", errParse)
	}
	if orderIdInt <= 0 || customerIdInt <= 0 {
		return nil, fmt.Errorf("cli.addOrder error: %w", errIncorrectId)
	}

	_, errDate := time.Parse(dateLayout, expirationTime)
	if errDate != nil {
		return nil, fmt.Errorf("cli.addOrder error: %w", errDate)
	}

	weightFloat, errParse := strconv.ParseFloat(args[4], 64)
	if errParse != nil {
		return nil, fmt.Errorf("cli.addOrder error: %w", errParse)
	}

	if weightFloat < 0 {
		return nil, fmt.Errorf("cli.addOrder error: %w", errNegativeWeight)
	}

	costInt, errParse := strconv.ParseFloat(args[5], 64)
	if errParse != nil {
		return nil, fmt.Errorf("cli.addOrder error: %w", errParse)
	}

	if costInt < 0 {
		return nil, fmt.Errorf("cli.addOrder error: %w", errNegativeCost)
	}

	return &orders_grpc.AddOrderRequest{
		OrderId:        orderIdInt,
		CustomerId:     customerIdInt,
		ExpirationTime: expirationTime,
		PackageType:    pack,
		Weight:         weightFloat,
		Cost:           costInt,
	}, nil
}

// returnOrder --orderId=1
func returnOrder(args []string) (*orders_grpc.ReturnOrderRequest, error) {
	if len(args) != 1 {
		return nil, errIncorrectArgAmount
	}

	orderIdInt, errParse := strconv.ParseInt(args[0], 10, 64)
	if errParse != nil {
		return nil, fmt.Errorf("cli.returnOrder error: %w", errParse)
	}
	if orderIdInt <= 0 {
		return nil, fmt.Errorf("cli.returnOrder error: %w", errIncorrectId)
	}

	return &orders_grpc.ReturnOrderRequest{
		OrderId: orderIdInt,
	}, nil
}

// receiveOrder --orders1=1,2,3,4,5
func receiveOrder(args []string) (*orders_grpc.ReceiveOrdersRequest, error) {
	if len(args) != 1 {
		return nil, errIncorrectArgAmount
	}

	orderIds, errParseId := parseIDs(args[0])
	if errParseId != nil {
		return nil, fmt.Errorf("cli.receiveOrder error: %w", errParseId)
	}

	return &orders_grpc.ReceiveOrdersRequest{
		OrderIds: orderIds,
	}, nil
}

// getOrders --customerId=1 --n=1
func getOrders(args []string) (*orders_grpc.GetOrdersRequest, error) {
	if len(args) != 2 {
		return nil, errIncorrectArgAmount
	}

	customerIdInt, errParse := strconv.ParseInt(args[0], 10, 64)
	if errParse != nil {
		return nil, fmt.Errorf("cli.getOrders error: %w", errParse)
	}
	if customerIdInt <= 0 {
		return nil, fmt.Errorf("cli.getOrders error: %w", errIncorrectId)
	}

	n, errParse := strconv.Atoi(args[1])
	if errParse != nil {
		return nil, fmt.Errorf("cli.getOrders error: %w", errParse)
	}

	return &orders_grpc.GetOrdersRequest{
		CustomerId: customerIdInt,
		N:          int32(n),
	}, nil
}

// createRefund --orderId=1 --customerId=1
func createRefund(args []string) (*orders_grpc.CreateRefundRequest, error) {
	if len(args) != 2 {
		return nil, errIncorrectArgAmount
	}

	orderIdInt, errParse := strconv.ParseInt(args[0], 10, 64)
	if errParse != nil {
		return nil, fmt.Errorf("cli.createRefund error: %w", errParse)
	}

	customerIdInt, errParse := strconv.ParseInt(args[1], 10, 64)
	if errParse != nil {
		return nil, fmt.Errorf("cli.createRefund error: %w", errParse)
	}

	if orderIdInt <= 0 || customerIdInt <= 0 {
		return nil, fmt.Errorf("cli.createRefund error: %w", errIncorrectId)
	}

	return &orders_grpc.CreateRefundRequest{
		OrderId:    orderIdInt,
		CustomerId: customerIdInt,
	}, nil
}

// getRefunds --page=1 --limit=1
func getRefunds(args []string) (*orders_grpc.GetRefundsRequest, error) {
	if len(args) != 2 {
		return nil, errIncorrectArgAmount
	}

	page, errParse := strconv.Atoi(args[0])
	if errParse != nil {
		return nil, fmt.Errorf("cli.getRefunds error: %w", errParse)
	}

	limit, errParse := strconv.Atoi(args[1])
	if errParse != nil {
		return nil, fmt.Errorf("cli.getRefunds error: %w", errParse)
	}

	return &orders_grpc.GetRefundsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	}, nil
}

func parseIDs(idsStr string) ([]int64, error) {
	if idsStr == "" {
		return nil, errIncorrectId
	}

	strIds := strings.Split(idsStr, ",")
	var ids []int64
	for _, strId := range strIds {
		id, errParse := strconv.ParseInt(strings.TrimSpace(strId), 10, 64)
		if errParse != nil {
			return nil, errParse
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func commandList() []command {
	return []command{
		{
			name:        helpCommand,
			description: "Справка",
		},
		{
			name:        addOrderCommand,
			description: "Добавить заказ",
		},
		{
			name:        returnOrderCommand,
			description: "Удалить заказ",
		},
		{
			name:        receiveOrderCommand,
			description: "Получить заказ",
		},
		{
			name:        getOrdersCommand,
			description: "Получить список заказов",
		},
		{
			name:        createRefundCommand,
			description: "Создать запрос на возврат",
		},
		{
			name:        getRefundsCommand,
			description: "Получить список возвратов",
		},
	}
}
