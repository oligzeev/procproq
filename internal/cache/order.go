package cache

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
)

type CachedOrderService struct {
	service domain.OrderService
	cache   *lru.Cache
}

func NewCachedOrderService(cacheSize int, service domain.OrderService) (*CachedOrderService, error) {
	const op = "CachedReadMappingService.Init"

	cache, err := lru.New(cacheSize)
	if err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't initialize lru cache (%d)", cacheSize), err)
	}
	return &CachedOrderService{service: service, cache: cache}, nil
}

func (s CachedOrderService) SubmitOrder(ctx context.Context, order *domain.Order, processId string) (*domain.Order, error) {
	/*result, err := s.service.SubmitOrder(ctx, order, processId)
	if err != nil {
		return nil, err
	}
	s.cache.Add(order.Id, order)
	return result, nil*/
	return s.service.SubmitOrder(ctx, order, processId)
}

func (s CachedOrderService) GetOrders(ctx context.Context) ([]domain.Order, error) {
	return s.service.GetOrders(ctx)
}

func (s CachedOrderService) GetOrderById(ctx context.Context, id string) (*domain.Order, error) {
	const op = "CachedReadMappingService.GetOrderById"

	if cachedObj, exists := s.cache.Get(id); exists {
		if obj, ok := cachedObj.(*domain.Order); ok {
			return obj, nil
		}
		return nil, domain.E(op, fmt.Sprintf("incorrect type of cached object (%T)", cachedObj))
	} else {
		obj, err := s.service.GetOrderById(ctx, id)
		if err == nil && obj != nil {
			s.cache.Add(obj.Id, obj)
		}
		return obj, err
	}
}

func (s CachedOrderService) CompleteJob(ctx context.Context, taskId, orderId string) error {
	return s.service.CompleteJob(ctx, taskId, orderId)
}
