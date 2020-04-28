package cache

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
)

type CachedReadMappingRepo struct {
	repo  domain.ReadMappingRepo
	cache *lru.Cache
}

func NewCachedReadMappingRepo(cacheSize int, repo domain.ReadMappingRepo) (*CachedReadMappingRepo, error) {
	var cache *lru.Cache
	var err error
	if cache, err = lru.New(cacheSize); err != nil {
		return nil, fmt.Errorf("can't create read mapping lru cache (%d): %v", cacheSize, err)
	}
	return &CachedReadMappingRepo{repo: repo, cache: cache}, nil
}

func (s CachedReadMappingRepo) GetAll(ctx context.Context) ([]domain.ReadMapping, error) {
	// Don't use cache
	return s.repo.GetAll(ctx)
}

func (s CachedReadMappingRepo) Create(ctx context.Context, obj *domain.ReadMapping) (*domain.ReadMapping, error) {
	result, err := s.repo.Create(ctx, obj)
	if err != nil {
		return nil, err
	}
	s.cache.Add(obj.Id, obj)
	return result, nil
}

func (s CachedReadMappingRepo) GetById(ctx context.Context, id string) (*domain.ReadMapping, error) {
	if cachedObj, exists := s.cache.Get(id); exists {
		if obj, ok := cachedObj.(*domain.ReadMapping); ok {
			log.Tracef("Return cached read mapping (%s)", id)
			return obj, nil
		}
		return nil, fmt.Errorf("can't convert %T to order", cachedObj)
	} else {
		var obj *domain.ReadMapping
		var err error
		if obj, err = s.repo.GetById(ctx, id); err == nil {
			if obj != nil && !exists {
				s.cache.Add(obj.Id, obj)
			}
		}
		return obj, err
	}
}

func (s CachedReadMappingRepo) DeleteById(ctx context.Context, id string) error {
	if err := s.repo.DeleteById(ctx, id); err != nil {
		return err
	}
	s.cache.Remove(id)
	return nil
}
