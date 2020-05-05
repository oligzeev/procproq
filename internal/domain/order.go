package domain

import (
	"context"
)

func CloneOrder(from, to *Order) {
	to.Id = from.Id
	to.ProcessId = from.ProcessId
	to.Body = from.Body
}

type Order struct {
	Id        string `json:"id"`
	ProcessId string `json:"processId"`
	Body      Body   `json:"body"`
}

/* TBD Structure stored in jsonb as-is
type OrderItem struct {
	Id string
	Instance Instance
}

type Instance struct {
	Id string
	Specification Specification
}

type Specification struct {
	Id string
}*/

type OrderService interface {
	SubmitOrder(ctx context.Context, order *Order, processId string) error
	GetOrders(ctx context.Context, result *[]Order) error
	GetOrderById(ctx context.Context, id string, result *Order) error
	CompleteJob(ctx context.Context, taskId, orderId string) error
}
