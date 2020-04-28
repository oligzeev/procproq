package service

import (
	"context"
	"errors"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/jmoiron/sqlx"
)

type OrderService struct {
	db          *sqlx.DB
	processRepo domain.ProcessRepo
	orderRepo   domain.OrderRepo
	jobRepo     domain.JobRepo
}

func NewOrderService(db *sqlx.DB, processRepo domain.ProcessRepo, orderRepo domain.OrderRepo, jobRepo domain.JobRepo) *OrderService {
	return &OrderService{db: db, processRepo: processRepo, orderRepo: orderRepo, jobRepo: jobRepo}
}

func (s OrderService) SubmitOrder(ctx context.Context, order *domain.Order, processId string) (*domain.Order, error) {
	process, err := s.processRepo.GetById(ctx, processId)
	if err != nil {
		return nil, err
	}
	if process == nil {
		return nil, errors.New("There's no process: " + processId)
	}
	order.ProcessId = processId
	result, err := database.ExecTx(ctx, s.db, func(txCtx context.Context) (interface{}, error) {
		resultOrder, err := s.orderRepo.Create(txCtx, order)
		if err != nil {
			return nil, err
		}
		err = s.jobRepo.CreateJobs(txCtx, order.Id, process)
		if err != nil {
			return nil, err
		}
		return resultOrder, nil
	})
	if result == nil {
		return nil, err
	}
	return result.(*domain.Order), err
}

func (s OrderService) GetOrders(ctx context.Context) ([]domain.Order, error) {
	return s.orderRepo.GetAll(ctx)
}

func (s OrderService) GetOrderById(ctx context.Context, id string) (*domain.Order, error) {
	// TBD Have to enrich current order state
	return s.orderRepo.GetById(ctx, id)
}

func (s OrderService) CompleteJob(ctx context.Context, taskId, orderId string) error {
	_, err := database.ExecTx(ctx, s.db, func(txCtx context.Context) (interface{}, error) {
		return nil, s.jobRepo.CompleteJob(txCtx, taskId, orderId)
	})
	return err
}
