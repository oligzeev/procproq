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

func toReadMapping(from *database.ReadMapping, to *domain.ReadMapping) {
	to.Id = from.Id
	to.Body = domain.Body(from.Body)
}

func fromReadMapping(from *domain.ReadMapping, to *database.ReadMapping) {
	to.Id = from.Id
	to.Body = database.Body(from.Body)
}

func toReadMappings(arr []database.ReadMapping) []domain.ReadMapping {
	result := make([]domain.ReadMapping, len(arr))
	for i, obj := range arr {
		toReadMapping(&obj, &result[i])
	}
	return result
}

type ReadMappingService struct {
	repo *database.ReadMappingRepo
}

func NewReadMappingService(readMappingRepo *database.ReadMappingRepo) *ReadMappingService {
	return &ReadMappingService{repo: readMappingRepo}
}

func (s ReadMappingService) GetAll(ctx context.Context, result *[]domain.ReadMapping) error {
	const op = "ReadMappingService.GetAll"

	var repoResult []database.ReadMapping
	err := s.repo.GetAll(ctx, &repoResult)
	if err != nil {
		return domain.E(op, err)
	}

	// Propagate result
	*result = toReadMappings(repoResult)
	return nil
}

func (s ReadMappingService) Create(ctx context.Context, result *domain.ReadMapping) error {
	const op = "ReadMappingService.Create"

	var dbResult database.ReadMapping
	fromReadMapping(result, &dbResult)
	err := s.repo.Create(ctx, &dbResult)
	if err != nil {
		return domain.E(op, err)
	}

	// Propagate generated id
	result.Id = dbResult.Id

	// Prepare jsonpath evaluators
	result.PreparedBody, err = prepareBody(result.Body)
	if err != nil {
		return domain.E(op, err)
	}
	return nil
}

func (s ReadMappingService) GetById(ctx context.Context, id string, result *domain.ReadMapping) error {
	const op = "ReadMappingService.GetById"

	var mapping database.ReadMapping
	err := s.repo.GetById(ctx, id, &mapping)
	if err != nil {
		return domain.E(op, err)
	}

	// Propagate result
	toReadMapping(&mapping, result)

	// Prepare jsonpath evaluators
	result.PreparedBody, err = prepareBody(result.Body)
	if err != nil {
		return domain.E(op, err)
	}
	return nil
}

func prepareBody(body domain.Body) (domain.PreparedBody, error) {
	const op = "ReadMappingService.PrepareBody"

	result := make(map[string]gval.Evaluable)
	for key, value := range body {
		strValue := value.(string)
		tasksPath, err := jsonpathLanguage.NewEvaluable(strValue)
		if err != nil {
			return nil, domain.E(op, fmt.Sprintf("can't create evaluator (%s)", value), err)
		}
		result[key] = tasksPath
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
