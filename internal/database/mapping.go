package database

import (
	"context"
	"database/sql"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/jmoiron/sqlx"
)

const (
	GetReadMappings       = `SELECT read_mapping_id, body FROM pp_read_mapping`
	CreateReadMapping     = `INSERT INTO pp_read_mapping (read_mapping_id, body) VALUES ($1, $2)`
	GetReadMappingById    = `SELECT read_mapping_id, body FROM pp_read_mapping WHERE read_mapping_id = $1`
	DeleteReadMappingById = `DELETE FROM pp_read_mapping WHERE read_mapping_id = $1`
)

type ReadMapping struct {
	Id   string `db:"read_mapping_id"`
	Body Body   `db:"body"`
}

type ReadMappingRepo struct {
	db          DB
	newUUIDFunc NewUUIDFunc
}

func NewReadMappingRepo(db *sqlx.DB, newUUIDFunc NewUUIDFunc) *ReadMappingRepo {
	return &ReadMappingRepo{db: db, newUUIDFunc: newUUIDFunc}
}

func (s ReadMappingRepo) GetAll(ctx context.Context, result *[]ReadMapping) error {
	const op = "ReadMappingRepo.GetAll"

	if err := s.db.SelectContext(ctx, result, GetReadMappings); err != nil {
		return domain.E(op, err)
	}
	return nil
}

func (s ReadMappingRepo) Create(ctx context.Context, result *ReadMapping) error {
	const op = "ReadMappingRepo.Create"

	id, err := s.newUUIDFunc()
	if err != nil {
		return domain.E(op, "can't generate uuid", err)
	}
	result.Id = id.String()

	if _, err := s.db.ExecContext(ctx, CreateReadMapping, result.Id, result.Body); err != nil {
		return domain.E(op, err)
	}
	return nil
}

func (s ReadMappingRepo) GetById(ctx context.Context, id string, result *ReadMapping) error {
	const op = "ReadMappingRepo.GetById"

	if err := s.db.GetContext(ctx, result, GetReadMappingById, id); err != nil {
		if err == sql.ErrNoRows {
			return domain.E(op, domain.ErrNotFound)
		}
		return domain.E(op, err)
	}
	return nil
}

func (s ReadMappingRepo) DeleteById(ctx context.Context, id string) error {
	const op = "ReadMappingRepo.DeleteById"

	result, err := s.db.ExecContext(ctx, DeleteReadMappingById, id)
	if err != nil {
		return domain.E(op, err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return domain.E(op, domain.ErrNotFound)
	}
	return nil
}
