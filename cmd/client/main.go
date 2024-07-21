package main

import (
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"homework-1/internal/utils"
	"homework-1/pkg/api/proto/orders_grpc/v1/orders_grpc/v1"
	"log"
	"os"
	"strings"
)

const (
	target = "localhost:50051"
)

func main() {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	client := orders_grpc.NewOrdersServiceClient(conn)

	ctx := context.Background()

	runClient(ctx, client)
}

func runClient(ctx context.Context, client orders_grpc.OrdersServiceClient) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("[>>] Введите команду:")
		cmd, errRead := reader.ReadString('\n')
		if errRead != nil {
			fmt.Printf("client error: %s", errRead)
			break
		}

		cmd = strings.TrimRight(cmd, "\n")
		if cmd == "exit" {
			break
		}

		req, errHandle := utils.HandleCommand(cmd)
		if errHandle != nil {
			fmt.Printf("client error: %s", errHandle)
		}

		proceedCommand(ctx, req, client)
	}
}

func proceedCommand(ctx context.Context, req interface{}, client orders_grpc.OrdersServiceClient) {
	switch req.(type) {
	case *orders_grpc.AddOrderRequest:
		_, errAdd := client.AddOrder(ctx, req.(*orders_grpc.AddOrderRequest))
		if errAdd != nil {
			st := status.Convert(errAdd)
			log.Printf("Ошибка добавления заказа: %v, %v", st.Code(), st.Message())
		}
		log.Println("Заказ добавлен успешно")
	case *orders_grpc.ReturnOrderRequest:
		_, errReturn := client.ReturnOrder(ctx, req.(*orders_grpc.ReturnOrderRequest))
		if errReturn != nil {
			st := status.Convert(errReturn)
			log.Printf("Ошибка возврата заказа: %v, %v", st.Code(), st.Message())
		}
		log.Println("Заказ возвращен успешно")
	case *orders_grpc.ReceiveOrdersRequest:
		resp, errReceive := client.ReceiveOrders(ctx, req.(*orders_grpc.ReceiveOrdersRequest))
		if errReceive != nil {
			st := status.Convert(errReceive)
			log.Printf("Ошибка получения заказов: %v, %v", st.Code(), st.Message())
		}
		for _, order := range resp.GetOrders() {
			log.Printf("Заказ: %v\n", order)
		}

	case *orders_grpc.GetOrdersRequest:
		resp, errGet := client.GetOrders(ctx, req.(*orders_grpc.GetOrdersRequest))
		if errGet != nil {
			st := status.Convert(errGet)
			log.Printf("Ошибка получения заказов: %v, %v", st.Code(), st.Message())
		}
		for _, order := range resp.GetOrders() {
			log.Printf("Заказ: %v\n", order)
		}
	case *orders_grpc.CreateRefundRequest:
		_, errRefund := client.CreateRefund(ctx, req.(*orders_grpc.CreateRefundRequest))
		if errRefund != nil {
			st := status.Convert(errRefund)
			log.Printf("Ошибка создания возврата: %v, %v", st.Code(), st.Message())
		}
		log.Println("Возврат создан успешно")
	case *orders_grpc.GetRefundsRequest:
		resp, errGetRefunds := client.GetRefunds(ctx, req.(*orders_grpc.GetRefundsRequest))
		if errGetRefunds != nil {
			st := status.Convert(errGetRefunds)
			log.Printf("Ошибка получения возвратов: %v, %v", st.Code(), st.Message())
		}
		for _, refund := range resp.GetRefunds() {
			log.Printf("Возврат: %v\n", refund)
		}
	}
}
