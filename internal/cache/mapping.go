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
	const op = "CachedReadMappingService.Init"

	cache, err := lru.New(cacheSize)
	if err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't initialize lru cache (%d)", cacheSize), err)
	}
	return &CachedReadMappingService{service: service, cache: cache}, nil
}

func (s CachedReadMappingService) GetAll(ctx context.Context, result *[]domain.ReadMapping) error {
	// Don't use cache
	return s.service.GetAll(ctx, result)
}

func (s CachedReadMappingService) Create(ctx context.Context, obj *domain.ReadMapping) error {
	err := s.service.Create(ctx, obj)
	if err != nil {
		return err
	}
	s.cache.Add(obj.Id, obj)
	return nil
}

func (s CachedReadMappingService) GetById(ctx context.Context, id string, result *domain.ReadMapping) error {
	const op = "CachedReadMappingService.GetById"

	if cachedObj, exists := s.cache.Get(id); exists {
		if cachedMapping, ok := cachedObj.(*domain.ReadMapping); ok {
			// Propagate values from cache
			domain.CloneReadMapping(cachedMapping, result)
			return nil
		}
		return domain.E(op, fmt.Sprintf("incorrect type of cached object (%T)", cachedObj))
	}
	if err := s.service.GetById(ctx, id, result); err != nil {
		return err
	}
	s.cache.Add(result.Id, result)
	return nil
}

func (s CachedReadMappingService) DeleteById(ctx context.Context, id string) error {
	if err := s.service.DeleteById(ctx, id); err != nil {
		return err
	}
	s.cache.Remove(id)
	return nil
}
