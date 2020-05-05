package database

import (
	"context"
	"database/sql"
	"example.com/oligzeev/pp-gin/internal/domain"
)

const (
	getReadMappings       = `SELECT read_mapping_id, body FROM pp_read_mapping`
	createReadMapping     = `INSERT INTO pp_read_mapping (read_mapping_id, body) VALUES ($1, $2)`
	getReadMappingById    = `SELECT read_mapping_id, body FROM pp_read_mapping WHERE read_mapping_id = $1`
	deleteReadMappingById = `DELETE FROM pp_read_mapping WHERE read_mapping_id = $1`
)

type ReadMapping struct {
	Id   string `db:"read_mapping_id"`
	Body Body   `db:"body"`
}

type ReadMappingRepo interface {
	GetAll(ctx context.Context, result *[]ReadMapping) error
	Create(ctx context.Context, order *ReadMapping) error
	GetById(ctx context.Context, id string, result *ReadMapping) error
	DeleteById(ctx context.Context, id string) error
}

type RDBReadMappingRepo struct {
	db          DB
	newUUIDFunc NewUUIDFunc
}

func NewRDBReadMappingRepo(db DB, newUUIDFunc NewUUIDFunc) ReadMappingRepo {
	return &RDBReadMappingRepo{db: db, newUUIDFunc: newUUIDFunc}
}

func (s RDBReadMappingRepo) GetAll(ctx context.Context, result *[]ReadMapping) error {
	const op = "ReadMappingRepo.GetAll"

	if err := s.db.SelectContext(ctx, result, getReadMappings); err != nil {
		return domain.E(op, err)
	}
	return nil
}

func (s RDBReadMappingRepo) Create(ctx context.Context, result *ReadMapping) error {
	const op = "ReadMappingRepo.Create"

	id, err := s.newUUIDFunc()
	if err != nil {
		return domain.E(op, "can't generate uuid", err)
	}
	result.Id = id.String()

	if _, err := s.db.ExecContext(ctx, createReadMapping, result.Id, result.Body); err != nil {
		return domain.E(op, err)
	}
	return nil
}

func (s RDBReadMappingRepo) GetById(ctx context.Context, id string, result *ReadMapping) error {
	const op = "ReadMappingRepo.GetById"

	if err := s.db.GetContext(ctx, result, getReadMappingById, id); err != nil {
		if err == sql.ErrNoRows {
			return domain.E(op, domain.ErrNotFound)
		}
		return domain.E(op, err)
	}
	return nil
}

func (s RDBReadMappingRepo) DeleteById(ctx context.Context, id string) error {
	const op = "ReadMappingRepo.DeleteById"

	result, err := s.db.ExecContext(ctx, deleteReadMappingById, id)
	if err != nil {
		return domain.E(op, err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return domain.E(op, domain.ErrNotFound)
	}
	return nil
}
