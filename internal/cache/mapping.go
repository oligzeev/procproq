package cache

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
)

type CachedReadMappingService struct {
	service domain.ReadMappingService
	cache   *lru.Cache
}

func NewCachedReadMappingService(cacheSize int, service domain.ReadMappingService) (*CachedReadMappingService, error) {
	const op domain.ErrOp = "CachedReadMappingService.Init"

	cache, err := lru.New(cacheSize)
	if err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't initialize lru cache (%d)", cacheSize), err)
	}
	return &CachedReadMappingService{service: service, cache: cache}, nil
}

func (s CachedReadMappingService) GetAll(ctx context.Context) ([]domain.ReadMapping, error) {
	// Don't use cache
	return s.service.GetAll(ctx)
}

func (s CachedReadMappingService) Create(ctx context.Context, obj *domain.ReadMapping) (*domain.ReadMapping, error) {
	result, err := s.service.Create(ctx, obj)
	if err != nil {
		return nil, err
	}
	s.cache.Add(obj.Id, obj)
	return result, nil
}

func (s CachedReadMappingService) GetById(ctx context.Context, id string) (*domain.ReadMapping, error) {
	const op domain.ErrOp = "CachedReadMappingService.GetById"

	if cachedObj, exists := s.cache.Get(id); exists {
		if obj, ok := cachedObj.(*domain.ReadMapping); ok {
			return obj, nil
		}
		return nil, domain.E(op, fmt.Sprintf("incorrect type of cached object (%T)", cachedObj))
	} else {
		obj, err := s.service.GetById(ctx, id)
		if err == nil && obj != nil {
			s.cache.Add(obj.Id, obj)
		}
		return obj, err
	}
}

func (s CachedReadMappingService) DeleteById(ctx context.Context, id string) error {
	if err := s.service.DeleteById(ctx, id); err != nil {
		return err
	}
	s.cache.Remove(id)
	return nil
}
