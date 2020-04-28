package database

import (
	"context"
	"database/sql"
	"errors"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	createProcess                  = `INSERT INTO pp_process (process_id, name) VALUES ($1, $2)`
	getProcesses                   = `SELECT process_id, name FROM pp_process`
	getProcessById                 = `SELECT process_id, name FROM pp_process WHERE process_id = $1 LIMIT 1`
	deleteProcessById              = `DELETE FROM pp_process WHERE process_id = $1`
	createTask                     = `INSERT INTO pp_task (process_id, task_id, name, category, action, read_mapping_id) VALUES ($1, $2, $3, $4, $5, $6)`
	createTaskRelation             = `INSERT INTO pp_task_rel (process_id, parent_id, child_id) VALUES ($1, $2, $3)`
	getTasks                       = `SELECT process_id, task_id, name, category, action, read_mapping_id FROM pp_task`
	getTaskRelations               = `SELECT process_id, parent_id, child_id FROM pp_task_rel`
	getTasksByProcessId            = `SELECT task_id, name, category, action, read_mapping_id FROM pp_task WHERE process_id = $1`
	getTaskRelationsByProcessId    = `SELECT parent_id, child_id FROM pp_task_rel WHERE process_id = $1`
	deleteTasksByProcessId         = `DELETE FROM pp_task WHERE process_id = $1`
	deleteTaskRelationsByProcessId = `DELETE FROM pp_task_rel WHERE process_id = $1`
)

// TBD It could be improved by storing jsonb or batch execution
// ProcessRepo via postgres database
type DbProcessRepo struct {
	db *sqlx.DB
}

func NewDbProcessRepo(db *sqlx.DB) *DbProcessRepo {
	return &DbProcessRepo{db: db}
}

// Get all processes
func (s DbProcessRepo) GetAll(ctx context.Context) ([]domain.Process, error) {
	var processes []domain.Process
	if err := s.db.SelectContext(ctx, &processes, getProcesses); err != nil {
		return nil, fmt.Errorf("can't get processes: %v", err)
	}
	if processes == nil {
		return make([]domain.Process, 0), nil
	}
	var tasks []domain.Task
	if err := s.db.SelectContext(ctx, &tasks, getTasks); err != nil {
		return nil, fmt.Errorf("can't get tasks: %v", err)
	}
	var relations []domain.TaskRelation
	if err := s.db.SelectContext(ctx, &relations, getTaskRelations); err != nil {
		return nil, fmt.Errorf("can't get task relations: %v", err)
	}
	for i, process := range processes {
		for _, task := range tasks {
			if task.ProcessId == process.Id {
				processes[i].AddTask(&task)
			}
		}
		for _, relation := range relations {
			if relation.ProcessId == process.Id {
				processes[i].AddTaskRelation(&relation)
			}
		}
	}
	return processes, nil
}

// Create process
func (s DbProcessRepo) Create(ctx context.Context, process *domain.Process) (*domain.Process, error) {
	if tx, ok := TransactionFromContext(ctx); ok {
		id, err := uuid.NewUUID()
		if err != nil {
			return nil, fmt.Errorf("can't generate uuid: %v", err)
		}
		process.Id = id.String()

		if _, err := tx.Exec(createProcess, process.Id, process.Name); err != nil {
			return nil, fmt.Errorf("can't create process (%s): %v", process.Id, err)
		}
		for _, task := range process.Tasks {
			if _, err := tx.Exec(createTask, process.Id, task.Id, task.Name, task.Category, task.Action, task.ReadMappingId); err != nil {
				return nil, fmt.Errorf("can't create task (%s, %s): %v", process.Id, task.Id, err)
			}
		}
		for _, rel := range process.TaskRelations {
			if _, err := tx.Exec(createTaskRelation, process.Id, rel.ParentId, rel.ChildId); err != nil {
				return nil, fmt.Errorf("can't create task relation (%s, %s, %s): %v", process.Id, rel.ParentId,
					rel.ChildId, err)
			}
		}
		return process, nil
	}
	return nil, errors.New("can't create process without transaction")
}

// Get process by Id
func (s DbProcessRepo) GetById(ctx context.Context, id string) (*domain.Process, error) {
	var process domain.Process
	if err := s.db.GetContext(ctx, &process, getProcessById, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("can't get process (%s): %v", id, err)
	}
	if err := s.db.SelectContext(ctx, &process.Tasks, getTasksByProcessId, id); err != nil {
		return nil, fmt.Errorf("can't get tasks by process (%s): %v", id, err)
	}
	if err := s.db.SelectContext(ctx, &process.TaskRelations, getTaskRelationsByProcessId, id); err != nil {
		return nil, fmt.Errorf("can't get task relations by process (%s): %v", id, err)
	}
	return &process, nil
}

// Delete process by Id
func (s DbProcessRepo) DeleteById(ctx context.Context, id string) error {
	if tx, ok := TransactionFromContext(ctx); ok {
		if _, err := tx.ExecContext(ctx, deleteProcessById, id); err != nil {
			return fmt.Errorf("can't delete process (%s): %v", id, err)
		}
		if _, err := tx.ExecContext(ctx, deleteTasksByProcessId, id); err != nil {
			return fmt.Errorf("can't delete tasks by process (%s): %v", id, err)
		}
		if _, err := tx.ExecContext(ctx, deleteTaskRelationsByProcessId, id); err != nil {
			return fmt.Errorf("can't delete task relations by process (%s): %v", id, err)
		}
		return nil
	}
	return errors.New("can't delete process without transaction")
}
