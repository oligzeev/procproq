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

type Process struct {
	Id            string `db:"process_id"`
	Name          string `db:"name"`
	Tasks         []Task
	TaskRelations []TaskRelation
}

func (p *Process) AddTask(task *Task) {
	p.Tasks = append(p.Tasks, *task)
}

func (p *Process) AddTaskRelation(taskRelation *TaskRelation) {
	p.TaskRelations = append(p.TaskRelations, *taskRelation)
}

type Task struct {
	ProcessId     string `db:"process_id"`
	Id            string `db:"task_id"`
	Name          string `db:"name"`
	Category      int    `db:"category"`
	Action        string `db:"action"`
	ReadMappingId string `db:"read_mapping_id"`
}

type TaskRelation struct {
	ProcessId string `db:"process_id"`
	ParentId  string `db:"parent_id"`
	ChildId   string `db:"child_id"`
}

// TBD It could be improved by storing jsonb or batch execution
// ProcessRepo via postgres database
type ProcessRepo struct {
	db *sqlx.DB
}

func NewProcessRepo(db *sqlx.DB) *ProcessRepo {
	return &ProcessRepo{db: db}
}

func (s ProcessRepo) GetAll(ctx context.Context) ([]Process, error) {
	const op = "ProcessRepo.GetAll"

	var processes []Process
	if err := s.db.SelectContext(ctx, &processes, getProcesses); err != nil {
		return nil, domain.E(op, "can't select processes", err)
	}
	if processes == nil {
		return make([]Process, 0), nil
	}
	var tasks []Task
	if err := s.db.SelectContext(ctx, &tasks, getTasks); err != nil {
		return nil, domain.E(op, "can't select tasks", err)
	}
	var relations []TaskRelation
	if err := s.db.SelectContext(ctx, &relations, getTaskRelations); err != nil {
		return nil, domain.E(op, "can't select task relations", err)
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

func (s ProcessRepo) Create(ctx context.Context, process *Process) (*Process, error) {
	const op = "ProcessRepo.Create"

	if tx, ok := TransactionFromContext(ctx); ok {
		id, err := uuid.NewUUID()
		if err != nil {
			return nil, domain.E(op, "can't generate uuid", err)
		}
		process.Id = id.String()

		if _, err := tx.Exec(createProcess, process.Id, process.Name); err != nil {
			return nil, domain.E(op, "can't insert process", err)
		}
		for _, task := range process.Tasks {
			if _, err := tx.Exec(createTask, process.Id, task.Id, task.Name, task.Category, task.Action,
				task.ReadMappingId); err != nil {

				return nil, domain.E(op, fmt.Sprintf("can't insert task (%s, %s)", process.Id, task.Id), err)
			}
		}
		for _, rel := range process.TaskRelations {
			if _, err := tx.Exec(createTaskRelation, process.Id, rel.ParentId, rel.ChildId); err != nil {
				return nil, domain.E(op, fmt.Sprintf("can't insert task relation (%s, %s, %s)", process.Id,
					rel.ParentId, rel.ChildId), err)
			}
		}
		return process, nil
	}
	return nil, domain.E(op, "there's no active transaction")
}

func (s ProcessRepo) GetById(ctx context.Context, id string) (*Process, error) {
	const op = "ProcessRepo.GetById"

	var process Process
	if err := s.db.GetContext(ctx, &process, getProcessById, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.E(op, domain.ErrNotFound)
		}
		return nil, domain.E(op, fmt.Sprintf("can't select process (%s)", id), err)
	}
	if err := s.db.SelectContext(ctx, &process.Tasks, getTasksByProcessId, id); err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't select tasks (%s)", id), err)
	}
	if err := s.db.SelectContext(ctx, &process.TaskRelations, getTaskRelationsByProcessId, id); err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't select task relations (%s)", id), err)
	}
	return &process, nil
}

func (s ProcessRepo) DeleteById(ctx context.Context, id string) error {
	const op = "ProcessRepo.DeleteById"

	if tx, ok := TransactionFromContext(ctx); ok {
		result, err := tx.ExecContext(ctx, deleteProcessById, id)
		if err != nil {
			return domain.E(op, fmt.Sprintf("can't delete process (%s)", id), err)
		}
		if count, _ := result.RowsAffected(); count == 0 {
			return domain.E(op, domain.ErrNotFound)
		}
		if _, err := tx.ExecContext(ctx, deleteTasksByProcessId, id); err != nil {
			return domain.E(op, fmt.Sprintf("can't delete tasks (%s)", id), err)
		}
		if _, err := tx.ExecContext(ctx, deleteTaskRelationsByProcessId, id); err != nil {
			return domain.E(op, fmt.Sprintf("can't delete task relations (%s)", id), err)
		}
		return nil
	}
	return domain.E(op, "there's no active transaction")
}
