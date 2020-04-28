package database

import (
	"context"
	"database/sql"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	getReadMappings       = `SELECT read_mapping_id, body FROM pp_read_mapping`
	createReadMapping     = `INSERT INTO pp_read_mapping (read_mapping_id, body) VALUES ($1, $2)`
	getReadMappingById    = `SELECT read_mapping_id, body FROM pp_read_mapping WHERE read_mapping_id = $1`
	deleteReadMappingById = `DELETE FROM pp_read_mapping WHERE read_mapping_id = $1`
)

type ReadMappingEntity struct {
	Id   string `db:"read_mapping_id"`
	Body Body   `db:"body"`
}

func toReadMapping(obj *ReadMappingEntity) *domain.ReadMapping {
	return &domain.ReadMapping{Id: obj.Id, Body: domain.Body(obj.Body)}
}

func toReadMappings(arr []ReadMappingEntity) []domain.ReadMapping {
	result := make([]domain.ReadMapping, len(arr))
	for i, obj := range arr {
		result[i].Id = obj.Id
		result[i].Body = domain.Body(obj.Body)
	}
	return result
}

// ReadMappingRepo via postgres database
type DbReadMappingRepo struct {
	db *sqlx.DB
}

func NewDbReadMappingRepo(db *sqlx.DB) *DbReadMappingRepo {
	return &DbReadMappingRepo{db: db}
}

// Get all read mappings
func (s DbReadMappingRepo) GetAll(ctx context.Context) ([]domain.ReadMapping, error) {
	var result []ReadMappingEntity
	if err := s.db.SelectContext(ctx, &result, getReadMappings); err != nil {
		return nil, fmt.Errorf("can't get read mappings: %v", err)
	}
	return toReadMappings(result), nil
}

// Create read mapping
func (s DbReadMappingRepo) Create(ctx context.Context, obj *domain.ReadMapping) (*domain.ReadMapping, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("can't generate uuid: %v", err)
	}
	obj.Id = id.String()
	if _, err := s.db.ExecContext(ctx, createReadMapping, obj.Id, Body(obj.Body)); err != nil {
		return nil, fmt.Errorf("can't create read mapping (%s): %v", obj.Id, err)
	}
	return obj, nil
}

// Get read mapping by Id
func (s DbReadMappingRepo) GetById(ctx context.Context, id string) (*domain.ReadMapping, error) {
	var result ReadMappingEntity
	if err := s.db.GetContext(ctx, &result, getReadMappingById, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("can't get read mapping (%s): %v", id, err)
	}
	return toReadMapping(&result), nil
}

// Delete read mapping by Id
func (s DbReadMappingRepo) DeleteById(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, deleteReadMappingById, id); err != nil {
		return fmt.Errorf("can't delete read mapping (%s): %v", id, err)
	}
	return nil
}
