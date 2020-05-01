package service

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
)

func toReadMapping(obj *database.ReadMapping) *domain.ReadMapping {
	return &domain.ReadMapping{Id: obj.Id, Body: domain.Body(obj.Body)}
}

func fromReadMapping(obj *domain.ReadMapping) *database.ReadMapping {
	return &database.ReadMapping{Id: obj.Id, Body: database.Body(obj.Body)}
}

func toReadMappings(arr []database.ReadMapping) []domain.ReadMapping {
	result := make([]domain.ReadMapping, len(arr))
	for i, obj := range arr {
		result[i].Id = obj.Id
		result[i].Body = domain.Body(obj.Body)
	}
	return result
}

type ReadMappingService struct {
	repo *database.ReadMappingRepo
}

func NewReadMappingService(readMappingRepo *database.ReadMappingRepo) *ReadMappingService {
	return &ReadMappingService{repo: readMappingRepo}
}

func (s ReadMappingService) GetAll(ctx context.Context) ([]domain.ReadMapping, error) {
	const op domain.ErrOp = "ReadMappingService.GetAll"

	result, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, domain.E(op, err)
	}
	return toReadMappings(result), nil
}

func (s ReadMappingService) Create(ctx context.Context, obj *domain.ReadMapping) (*domain.ReadMapping, error) {
	const op domain.ErrOp = "ReadMappingService.Create"

	result, err := s.repo.Create(ctx, fromReadMapping(obj))
	if err != nil {
		return nil, domain.E(op, err)
	}
	return toReadMapping(result), nil
}

func (s ReadMappingService) GetById(ctx context.Context, id string) (*domain.ReadMapping, error) {
	const op domain.ErrOp = "ReadMappingService.GetById"

	result, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, domain.E(op, err)
	}
	return toReadMapping(result), err
}

func (s ReadMappingService) DeleteById(ctx context.Context, id string) error {
	const op domain.ErrOp = "ReadMappingService.DeleteById"

	err := s.repo.DeleteById(ctx, id)
	if err != nil {
		return domain.E(op, err)
	}
	return nil
}
