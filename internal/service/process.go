package service

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/jmoiron/sqlx"
)

func toProcesses(arr []database.Process) []domain.Process {
	result := make([]domain.Process, len(arr))
	for i, obj := range arr {
		result[i].Id = obj.Id
		result[i].Name = obj.Name
		result[i].Tasks = toTasks(obj.Tasks)
		result[i].TaskRelations = toTaskRelations(obj.TaskRelations)
	}
	return result
}

func toProcess(obj *database.Process) *domain.Process {
	return &domain.Process{
		Id:            obj.Id,
		Name:          obj.Name,
		Tasks:         toTasks(obj.Tasks),
		TaskRelations: toTaskRelations(obj.TaskRelations),
	}
}

func fromProcess(obj *domain.Process) *database.Process {
	return &database.Process{
		Id:            obj.Id,
		Name:          obj.Name,
		Tasks:         fromTasks(obj.Id, obj.Tasks),
		TaskRelations: fromTaskRelations(obj.Id, obj.TaskRelations),
	}
}

func toTasks(arr []database.Task) []domain.Task {
	result := make([]domain.Task, len(arr))
	for i, obj := range arr {
		result[i].Id = obj.Id
		result[i].Name = obj.Name
		result[i].Category = obj.Category
		result[i].Action = obj.Action
		result[i].ReadMappingId = obj.ReadMappingId
	}
	return result
}

func fromTasks(processId string, arr []domain.Task) []database.Task {
	result := make([]database.Task, len(arr))
	for i, obj := range arr {
		result[i].ProcessId = processId
		result[i].Id = obj.Id
		result[i].Name = obj.Name
		result[i].Category = obj.Category
		result[i].Action = obj.Action
		result[i].ReadMappingId = obj.ReadMappingId
	}
	return result
}

func toTaskRelations(arr []database.TaskRelation) []domain.TaskRelation {
	result := make([]domain.TaskRelation, len(arr))
	for i, obj := range arr {
		result[i].ParentId = obj.ParentId
		result[i].ChildId = obj.ChildId
	}
	return result
}

func fromTaskRelations(processId string, arr []domain.TaskRelation) []database.TaskRelation {
	result := make([]database.TaskRelation, len(arr))
	for i, obj := range arr {
		result[i].ProcessId = processId
		result[i].ParentId = obj.ParentId
		result[i].ChildId = obj.ChildId
	}
	return result
}

type ProcessService struct {
	db   *sqlx.DB
	repo *database.ProcessRepo
}

func NewProcessService(db *sqlx.DB, processRepo *database.ProcessRepo) *ProcessService {
	return &ProcessService{db: db, repo: processRepo}
}

func (s ProcessService) GetAll(ctx context.Context) ([]domain.Process, error) {
	const op = "ProcessService.GetAll"

	result, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, domain.E(op, err)
	}
	return toProcesses(result), nil
}

func (s ProcessService) Create(ctx context.Context, process *domain.Process) (*domain.Process, error) {
	const op = "ProcessService.Create"

	result, err := database.ExecTx(ctx, s.db, func(txCtx context.Context) (interface{}, error) {
		return s.repo.Create(txCtx, fromProcess(process))
	})
	if err != nil {
		return nil, domain.E(op, err)
	}
	return toProcess(result.(*database.Process)), nil
}

func (s ProcessService) GetById(ctx context.Context, id string) (*domain.Process, error) {
	const op = "ProcessService.GetById"

	result, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, domain.E(op, err)
	}
	return toProcess(result), err
}

func (s ProcessService) DeleteById(ctx context.Context, id string) error {
	const op = "ProcessService.DeleteById"

	_, err := database.ExecTx(ctx, s.db, func(txCtx context.Context) (interface{}, error) {
		return nil, s.repo.DeleteById(txCtx, id)
	})
	if err != nil {
		return domain.E(op, err)
	}
	return nil
}
