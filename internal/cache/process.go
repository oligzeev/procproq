package cache

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
)

type CachedProcessRepo struct {
	repo  domain.ProcessRepo
	cache *lru.Cache
}

func NewCachedProcessRepo(cacheSize int, repo domain.ProcessRepo) (*CachedProcessRepo, error) {
	var cache *lru.Cache
	var err error
	if cache, err = lru.New(cacheSize); err != nil {
		return nil, fmt.Errorf("can't create process lru cache (%d): %v", cacheSize, err)
	}
	return &CachedProcessRepo{repo: repo, cache: cache}, nil
}

func (s CachedProcessRepo) GetAll(ctx context.Context) ([]domain.Process, error) {
	// Don't use cache
	return s.repo.GetAll(ctx)
}

func (s CachedProcessRepo) Create(ctx context.Context, obj *domain.Process) (*domain.Process, error) {
	result, err := s.repo.Create(ctx, obj)
	if err != nil {
		return nil, err
	}
	s.cache.Add(obj.Id, obj)
	return result, nil
}

func (s CachedProcessRepo) GetById(ctx context.Context, id string) (*domain.Process, error) {
	if cachedObj, exists := s.cache.Get(id); exists {
		if obj, ok := cachedObj.(*domain.Process); ok {
			log.Tracef("Return cached process (%s)", id)
			return obj, nil
		}
		return nil, fmt.Errorf("can't convert %T to order", cachedObj)
	} else {
		var obj *domain.Process
		var err error
		if obj, err = s.repo.GetById(ctx, id); err == nil {
			if obj != nil && !exists {
				s.cache.Add(obj.Id, obj)
			}
		}
		return obj, err
	}
}

func (s CachedProcessRepo) DeleteById(ctx context.Context, id string) error {
	if err := s.repo.DeleteById(ctx, id); err != nil {
		return err
	}
	s.cache.Remove(id)
	return nil
}
