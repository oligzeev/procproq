package service

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
)

func toProcess(from *database.Process, to *domain.Process) {
	to.Id = from.Id
	to.Name = from.Name
	to.Tasks = toTasks(from.Tasks)
	to.TaskRelations = toTaskRelations(from.TaskRelations)
}

func fromProcess(from *domain.Process, to *database.Process) {
	to.Id = from.Id
	to.Name = from.Name
	to.Tasks = fromTasks(from.Id, from.Tasks)
	to.TaskRelations = fromTaskRelations(from.Id, from.TaskRelations)
}

func toProcesses(arr []database.Process) []domain.Process {
	result := make([]domain.Process, len(arr))
	for i, obj := range arr {
		toProcess(&obj, &result[i])
	}
	return result
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
	repo       database.ProcessRepo
	execTxFunc domain.ExecTxFunc
}

func NewProcessService(processRepo database.ProcessRepo, execTxFunc domain.ExecTxFunc) *ProcessService {
	return &ProcessService{repo: processRepo, execTxFunc: execTxFunc}
}

func (s ProcessService) GetAll(ctx context.Context, result *[]domain.Process) error {
	const op = "ProcessService.GetAll"

	var repoResult []database.Process
	if err := s.repo.GetAll(ctx, &repoResult); err != nil {
		return domain.E(op, err)
	}

	// Propagate result
	*result = toProcesses(repoResult)
	return nil
}

func (s ProcessService) Create(ctx context.Context, result *domain.Process) error {
	const op = "ProcessService.Create"

	var repoResult database.Process
	fromProcess(result, &repoResult)
	err := s.execTxFunc(ctx, func(txCtx context.Context) error {
		return s.repo.Create(txCtx, &repoResult)
	})
	if err != nil {
		return domain.E(op, err)
	}

	// Propagate generated id
	result.Id = repoResult.Id
	return nil
}

func (s ProcessService) GetById(ctx context.Context, id string, result *domain.Process) error {
	const op = "ProcessService.GetById"

	var repoResult database.Process
	if err := s.repo.GetById(ctx, id, &repoResult); err != nil {
		return domain.E(op, err)
	}

	// Propagate result
	toProcess(&repoResult, result)
	return nil
}

func (s ProcessService) DeleteById(ctx context.Context, id string) error {
	const op = "ProcessService.DeleteById"

	err := s.execTxFunc(ctx, func(txCtx context.Context) error {
		return s.repo.DeleteById(txCtx, id)
	})
	if err != nil {
		return domain.E(op, err)
	}
	return nil
}
