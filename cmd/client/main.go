package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"homework-1/pkg/api/proto/orders_grpc/v1/orders_grpc/v1"
	"log"
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

	ctx = metadata.AppendToOutgoingContext(ctx, "x-my-header", "123")

	if errAdd := addOrder(ctx, client); errAdd != nil {
		log.Fatal(errAdd)
		return
	}

	if errGet := getOrders(ctx, client); errGet != nil {
		log.Fatal(errGet)
		return
	}

	// Из-за особенности реализации данный метод всегда будет выкидывать ошибку (добавить заказ с датой истечения срока
	// действия в прошлом нельзя, при этом вернуть заказ, срок хранения которого не прошел так же нельзя)

	//if errReturn := returnOrder(ctx, client); errReturn != nil {
	//	log.Fatal(errReturn)
	//	return
	//}

	if errReceive := receiveOrders(ctx, client); errReceive != nil {
		log.Fatal(errReceive)
		return
	}

	if errCreate := createRefund(ctx, client); errCreate != nil {
		log.Fatal(errCreate)
		return
	}

	if errGetRefunds := getRefunds(ctx, client); errGetRefunds != nil {
		log.Fatal(errGetRefunds)
		return
	}

	log.Println("Client done")
}

func addOrder(ctx context.Context, client orders_grpc.OrdersServiceClient) error {
	_, errAdd := client.AddOrder(ctx, &orders_grpc.AddOrderRequest{
		OrderId:        111,
		CustomerId:     101,
		ExpirationTime: "10-10-2024",
		PackageType:    "box",
		Weight:         10,
		Cost:           10,
	})

	if errAdd != nil {
		st := status.Convert(errAdd)
		log.Printf("Ошибка добавления заказа: %v, %v", st.Code(), st.Message())
		return errAdd
	}

	log.Println("Заказ добавлен успешно")
	return nil
}

func getOrders(ctx context.Context, client orders_grpc.OrdersServiceClient) error {
	resp, errGet := client.GetOrders(ctx, &orders_grpc.GetOrdersRequest{
		CustomerId: 101,
		N:          10,
	})

	if errGet != nil {
		st := status.Convert(errGet)
		log.Printf("Error getting orders: %v, %v", st.Code(), st.Message())
		return errGet
	}

	for _, order := range resp.GetOrders() {
		log.Printf("Заказ: %v\n", order)
	}

	return nil
}

func returnOrder(ctx context.Context, client orders_grpc.OrdersServiceClient) error {
	_, errReturn := client.ReturnOrder(ctx, &orders_grpc.ReturnOrderRequest{
		OrderId: 1,
	})

	if errReturn != nil {
		st := status.Convert(errReturn)
		log.Printf("Ошибка возврата заказа: %v, %v", st.Code(), st.Message())
		return errReturn
	}

	log.Println("Заказ возвращен успешно")
	return nil
}

func receiveOrders(ctx context.Context, client orders_grpc.OrdersServiceClient) error {
	resp, errReceive := client.ReceiveOrders(ctx, &orders_grpc.ReceiveOrdersRequest{
		OrderIds: []int64{111},
	})

	if errReceive != nil {
		st := status.Convert(errReceive)
		log.Printf("Ошибка получения заказов: %v, %v", st.Code(), st.Message())
		return errReceive
	}

	for _, order := range resp.GetOrders() {
		log.Printf("Заказ: %v\n", order)
	}

	return nil
}

func createRefund(ctx context.Context, client orders_grpc.OrdersServiceClient) error {
	_, errRefund := client.CreateRefund(ctx, &orders_grpc.CreateRefundRequest{
		OrderId:    111,
		CustomerId: 101,
	})

	if errRefund != nil {
		st := status.Convert(errRefund)
		log.Printf("Ошибка создания возврата: %v, %v", st.Code(), st.Message())
		return errRefund
	}

	log.Println("Возврат создан успешно")
	return nil
}

func getRefunds(ctx context.Context, client orders_grpc.OrdersServiceClient) error {
	resp, errRefunds := client.GetRefunds(ctx, &orders_grpc.GetRefundsRequest{
		Page:  0,
		Limit: 10,
	})

	if errRefunds != nil {
		st := status.Convert(errRefunds)
		log.Printf("Ошибка получения возвратов: %v, %v", st.Code(), st.Message())
		return errRefunds
	}

	for _, refund := range resp.GetRefunds() {
		log.Printf("Возврат: %v\n", refund)
	}

	return nil
}
