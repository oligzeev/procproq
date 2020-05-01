package tracing

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/opentracing/opentracing-go"
)

type SpanOrderService struct {
	service domain.OrderService
}

func NewSpanOrderService(service domain.OrderService) *SpanOrderService {
	return &SpanOrderService{service: service}
}

func (s SpanOrderService) SubmitOrder(ctx context.Context, order *domain.Order, processId string) (*domain.Order, error) {
	const op = "OrderService.SubmitOrder"
	span, spanCtx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()
	return s.service.SubmitOrder(spanCtx, order, processId)
}

func (s SpanOrderService) GetOrders(ctx context.Context) ([]domain.Order, error) {
	const op = "OrderService.GetOrders"
	span, spanCtx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()
	return s.service.GetOrders(spanCtx)
}

func (s SpanOrderService) GetOrderById(ctx context.Context, id string) (*domain.Order, error) {
	const op = "OrderService.GetOrderById"
	span, spanCtx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()
	return s.service.GetOrderById(spanCtx, id)
}

func (s SpanOrderService) CompleteJob(ctx context.Context, taskId, orderId string) error {
	const op = "OrderService.CompleteJob"
	span, spanCtx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()
	return s.service.CompleteJob(spanCtx, taskId, orderId)
}
