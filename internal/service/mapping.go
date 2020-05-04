package service

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	"github.com/PaesslerAG/gval"
	"github.com/PaesslerAG/jsonpath"
)

var (
	jsonpathLanguage = gval.Full(jsonpath.PlaceholderExtension())
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
	const op = "ReadMappingService.GetAll"

	var result []database.ReadMapping
	err := s.repo.GetAll(ctx, &result)
	if err != nil {
		return nil, domain.E(op, err)
	}
	return toReadMappings(result), nil
}

func (s ReadMappingService) Create(ctx context.Context, obj *domain.ReadMapping) (*domain.ReadMapping, error) {
	const op = "ReadMappingService.Create"

	dbObj := fromReadMapping(obj)
	err := s.repo.Create(ctx, dbObj)
	if err != nil {
		return nil, domain.E(op, err)
	}
	return toReadMapping(dbObj), nil
}

func (s ReadMappingService) GetById(ctx context.Context, id string) (*domain.ReadMapping, error) {
	const op = "ReadMappingService.GetById"

	var mapping database.ReadMapping
	err := s.repo.GetById(ctx, id, &mapping)
	if err != nil {
		return nil, domain.E(op, err)
	}
	result := toReadMapping(&mapping)
	result.PreparedBody = make(map[string]gval.Evaluable)
	for key, value := range result.Body {
		strValue := value.(string)
		tasksPath, err := jsonpathLanguage.NewEvaluable(strValue)
		if err != nil {
			return nil, domain.E(op, fmt.Sprintf("can't create evaluator (%s)", value), err)
		}
		result.PreparedBody[key] = tasksPath
	}
	return result, nil
}

func (s ReadMappingService) DeleteById(ctx context.Context, id string) error {
	const op = "ReadMappingService.DeleteById"

	err := s.repo.DeleteById(ctx, id)
	if err != nil {
		return domain.E(op, err)
	}
	return nil
}
