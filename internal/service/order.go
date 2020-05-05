package service

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
)

func toOrder(from *database.Order, to *domain.Order) {
	to.Id = from.Id
	to.ProcessId = from.ProcessId
	to.Body = domain.Body(from.Body)
}

func fromOrder(from *domain.Order, to *database.Order) {
	to.Id = from.Id
	to.ProcessId = from.ProcessId
	to.Body = database.Body(from.Body)
}

func toOrders(arr []database.Order) []domain.Order {
	result := make([]domain.Order, len(arr))
	for i, obj := range arr {
		toOrder(&obj, &result[i])
	}
	return result
}

type OrderService struct {
	processService domain.ProcessService
	orderRepo      database.OrderRepo
	jobRepo        database.JobRepo
	execTxFunc     domain.ExecTxFunc
}

func NewOrderService(processService domain.ProcessService, orderRepo database.OrderRepo, jobRepo database.JobRepo,
	execTxFunc domain.ExecTxFunc) *OrderService {

	return &OrderService{
		processService: processService,
		orderRepo:      orderRepo,
		jobRepo:        jobRepo,
		execTxFunc:     execTxFunc,
	}
}

func (s OrderService) SubmitOrder(ctx context.Context, order *domain.Order, processId string) error {
	const op = "OrderService.SubmitOrder"

	var process domain.Process
	if err := s.processService.GetById(ctx, processId, &process); err != nil {
		return domain.E(op, err)
	}
	order.ProcessId = processId
	err := s.execTxFunc(ctx, func(txCtx context.Context) error {
		var repoOrder database.Order
		fromOrder(order, &repoOrder)
		if err := s.orderRepo.Create(txCtx, &repoOrder); err != nil {
			return err
		}

		// TBD remove redundant operation 'fromProcess'
		var repoProcess database.Process
		fromProcess(&process, &repoProcess)
		if err := s.jobRepo.CreateJobs(txCtx, repoOrder.Id, &repoProcess); err != nil {
			return err
		}

		// Propagate generated id
		order.Id = repoOrder.Id
		return nil
	})
	if err != nil {
		return domain.E(op, err)
	}
	return nil
}

func (s OrderService) GetOrders(ctx context.Context, result *[]domain.Order) error {
	const op = "OrderService.GetOrders"

	var repoResult []database.Order
	if err := s.orderRepo.GetAll(ctx, &repoResult); err != nil {
		return domain.E(op, err)
	}

	// Propagate result
	*result = toOrders(repoResult)
	return nil
}

func (s OrderService) GetOrderById(ctx context.Context, id string, result *domain.Order) error {
	const op = "OrderService.GetOrderById"

	// TBD Have to enrich current order state
	var repoResult database.Order
	if err := s.orderRepo.GetById(ctx, id, &repoResult); err != nil {
		return domain.E(op, err)
	}

	// Propagate result
	toOrder(&repoResult, result)
	return nil
}

func (s OrderService) CompleteJob(ctx context.Context, taskId, orderId string) error {
	const op = "OrderService.CompleteJob"

	err := s.execTxFunc(ctx, func(txCtx context.Context) error {
		return s.jobRepo.CompleteJob(txCtx, taskId, orderId)
	})
	if err != nil {
		return domain.E(op, err)
	}
	return nil
}
