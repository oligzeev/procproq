package domain

import (
	"context"
)

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

type OrderRepo interface {
	GetAll(ctx context.Context) ([]Order, error)
	Create(ctx context.Context, order *Order) (*Order, error)
	GetById(ctx context.Context, id string) (*Order, error)
	DeleteById(ctx context.Context, id string) error
}

type OrderService interface {
	SubmitOrder(ctx context.Context, order *Order, processId string) (*Order, error)
	GetOrders(ctx context.Context) ([]Order, error)
	GetOrderById(ctx context.Context, id string) (*Order, error)
	CompleteJob(ctx context.Context, taskId, orderId string) error
}
