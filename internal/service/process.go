package service

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/jmoiron/sqlx"
)

type ProcessService struct {
	db              *sqlx.DB
	processRepo     domain.ProcessRepo
	readMappingRepo domain.ReadMappingRepo
}

func NewProcessService(db *sqlx.DB, processRepo domain.ProcessRepo, readMappingRepo domain.ReadMappingRepo) *ProcessService {
	return &ProcessService{db: db, processRepo: processRepo, readMappingRepo: readMappingRepo}
}

func (s ProcessService) GetProcesses(ctx context.Context) ([]domain.Process, error) {
	return s.processRepo.GetAll(ctx)
}

func (s ProcessService) CreateProcess(ctx context.Context, process *domain.Process) (*domain.Process, error) {
	result, err := database.ExecTx(ctx, s.db, func(txCtx context.Context) (interface{}, error) {
		return s.processRepo.Create(txCtx, process)
	})
	if result == nil {
		return nil, err
	}
	return result.(*domain.Process), err
}

func (s ProcessService) GetProcessById(ctx context.Context, id string) (*domain.Process, error) {
	return s.processRepo.GetById(ctx, id)
}

func (s ProcessService) DeleteProcessById(ctx context.Context, id string) error {
	_, err := database.ExecTx(ctx, s.db, func(txCtx context.Context) (interface{}, error) {
		return nil, s.processRepo.DeleteById(txCtx, id)
	})
	return err
}

func (s ProcessService) GetReadMappings(ctx context.Context) ([]domain.ReadMapping, error) {
	return s.readMappingRepo.GetAll(ctx)
}

func (s ProcessService) CreateReadMapping(ctx context.Context, mapping *domain.ReadMapping) (*domain.ReadMapping, error) {
	return s.readMappingRepo.Create(ctx, mapping)
}

func (s ProcessService) GetReadMappingById(ctx context.Context, id string) (*domain.ReadMapping, error) {
	return s.readMappingRepo.GetById(ctx, id)
}

func (s ProcessService) DeleteReadMappingById(ctx context.Context, id string) error {
	return s.readMappingRepo.DeleteById(ctx, id)
}
