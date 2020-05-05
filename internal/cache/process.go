package cache

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
)

type CachedProcessService struct {
	service domain.ProcessService
	cache   *lru.Cache
}

func NewCachedProcessRepo(cacheSize int, service domain.ProcessService) (*CachedProcessService, error) {
	const op = "CachedProcessService.Init"

	cache, err := lru.New(cacheSize)
	if err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't initialize lru cache (%d)", cacheSize), err)
	}
	return &CachedProcessService{service: service, cache: cache}, nil
}

func (s CachedProcessService) GetAll(ctx context.Context, result *[]domain.Process) error {
	// Don't use cache
	return s.service.GetAll(ctx, result)
}

func (s CachedProcessService) Create(ctx context.Context, obj *domain.Process) error {
	err := s.service.Create(ctx, obj)
	if err != nil {
		return err
	}
	s.cache.Add(obj.Id, obj)
	return nil
}

func (s CachedProcessService) GetById(ctx context.Context, id string, result *domain.Process) error {
	const op = "CachedProcessService.GetById"

	if cachedObj, exists := s.cache.Get(id); exists {
		if cachedProcess, ok := cachedObj.(*domain.Process); ok {
			// Propagate values from cache
			domain.CloneProcess(cachedProcess, result)
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

func (s CachedProcessService) DeleteById(ctx context.Context, id string) error {
	if err := s.service.DeleteById(ctx, id); err != nil {
		return err
	}
	s.cache.Remove(id)
	return nil
}
