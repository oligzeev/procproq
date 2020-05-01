package service

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/jmoiron/sqlx"
)

func toOrders(arr []database.Order) []domain.Order {
	result := make([]domain.Order, len(arr))
	for i, obj := range arr {
		result[i].Id = obj.Id
		result[i].ProcessId = obj.ProcessId
		result[i].Body = domain.Body(obj.Body)
	}
	return result
}

func toOrder(obj *database.Order) *domain.Order {
	return &domain.Order{Id: obj.Id, ProcessId: obj.ProcessId, Body: domain.Body(obj.Body)}
}

func fromOrder(obj *domain.Order) *database.Order {
	return &database.Order{Id: obj.Id, ProcessId: obj.ProcessId, Body: database.Body(obj.Body)}
}

type OrderService struct {
	db             *sqlx.DB
	processService domain.ProcessService
	orderRepo      *database.OrderRepo
	jobRepo        *database.JobRepo
}

func NewOrderService(db *sqlx.DB, processService domain.ProcessService, orderRepo *database.OrderRepo,
	jobRepo *database.JobRepo) *OrderService {

	return &OrderService{db: db, processService: processService, orderRepo: orderRepo, jobRepo: jobRepo}
}

func (s OrderService) SubmitOrder(ctx context.Context, order *domain.Order, processId string) (*domain.Order, error) {
	const op = "OrderService.SubmitOrder"

	process, err := s.processService.GetById(ctx, processId)
	if err != nil {
		return nil, domain.E(op, err)
	}
	order.ProcessId = processId
	result, err := database.ExecTx(ctx, s.db, func(txCtx context.Context) (interface{}, error) {
		resultOrder, err := s.orderRepo.Create(txCtx, fromOrder(order))
		if err != nil {
			return nil, err
		}
		// TBD remove redundant operation 'fromProcess'
		err = s.jobRepo.CreateJobs(txCtx, resultOrder.Id, fromProcess(process))
		if err != nil {
			return nil, err
		}
		return resultOrder, nil
	})
	if err != nil {
		return nil, domain.E(op, err)
	}
	return toOrder(result.(*database.Order)), nil
}

func (s OrderService) GetOrders(ctx context.Context) ([]domain.Order, error) {
	const op = "OrderService.GetOrders"

	result, err := s.orderRepo.GetAll(ctx)
	if err != nil {
		return nil, domain.E(op, err)
	}
	return toOrders(result), nil
}

func (s OrderService) GetOrderById(ctx context.Context, id string) (*domain.Order, error) {
	const op = "OrderService.GetOrderById"

	// TBD Have to enrich current order state
	result, err := s.orderRepo.GetById(ctx, id)
	if err != nil {
		return nil, domain.E(op, err)
	}
	return toOrder(result), nil
}

func (s OrderService) CompleteJob(ctx context.Context, taskId, orderId string) error {
	const op = "OrderService.CompleteJob"

	_, err := database.ExecTx(ctx, s.db, func(txCtx context.Context) (interface{}, error) {
		return nil, s.jobRepo.CompleteJob(txCtx, taskId, orderId)
	})
	if err != nil {
		return domain.E(op, err)
	}
	return nil
}
