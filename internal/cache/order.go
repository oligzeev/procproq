package cache

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
)

type CachedOrderRepo struct {
	repo  domain.OrderRepo
	cache *lru.Cache
}

func NewCachedOrderRepo(cacheSize int, repo domain.OrderRepo) (*CachedOrderRepo, error) {
	var cache *lru.Cache
	var err error
	if cache, err = lru.New(cacheSize); err != nil {
		return nil, fmt.Errorf("can't create order lru cache (%d): %v", cacheSize, err)
	}
	return &CachedOrderRepo{repo: repo, cache: cache}, nil
}

func (s CachedOrderRepo) GetAll(ctx context.Context) ([]domain.Order, error) {
	// Don't use cache
	return s.repo.GetAll(ctx)
}

func (s CachedOrderRepo) Create(ctx context.Context, obj *domain.Order) (*domain.Order, error) {
	result, err := s.repo.Create(ctx, obj)
	if err != nil {
		return nil, err
	}
	s.cache.Add(obj.Id, obj)
	return result, nil
}

func (s CachedOrderRepo) GetById(ctx context.Context, id string) (*domain.Order, error) {
	if cachedObj, exists := s.cache.Get(id); exists {
		if obj, ok := cachedObj.(*domain.Order); ok {
			log.Tracef("Return cached read mapping (%s)", id)
			return obj, nil
		}
		return nil, fmt.Errorf("can't convert %T to order", cachedObj)
	} else {
		var obj *domain.Order
		var err error
		if obj, err = s.repo.GetById(ctx, id); err == nil {
			if obj != nil && !exists {
				s.cache.Add(obj.Id, obj)
			}
		}
		return obj, err
	}
}

func (s CachedOrderRepo) DeleteById(ctx context.Context, id string) error {
	if err := s.repo.DeleteById(ctx, id); err != nil {
		return err
	}
	s.cache.Remove(id)
	return nil
}
