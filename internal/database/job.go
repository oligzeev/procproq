package database

import (
	"context"
	"errors"
	"example.com/oligzeev/pp-gin/internal/domain"
	"example.com/oligzeev/pp-gin/internal/tracing"
	"fmt"
	"github.com/jmoiron/sqlx"
)

const (
	createJobs = `INSERT INTO pp_job
(process_id, task_id, category, action, order_id, read_mapping_id, started, completed, ready_num, ready_req, trace)
VALUES ($1, $2, $3, $4, $5, $6, FALSE, FALSE, 0, $7, $8)`
	getReadyJobs = `UPDATE pp_job SET started = TRUE
WHERE (task_id, order_id) IN (
  SELECT task_id, order_id FROM pp_job WHERE ready_num >= ready_req AND started = FALSE LIMIT $1
) RETURNING task_id, category, action, order_id, read_mapping_id, trace`
	completeJob         = `UPDATE pp_job SET completed = TRUE WHERE completed = FALSE AND task_id = $1 and order_id = $2`
	completeRelatedJobs = `UPDATE pp_job t
SET ready_num = t.ready_num + 1
WHERE t.completed = FALSE AND t.task_id IN (
  SELECT r.child_id FROM pp_task_rel r WHERE r.parent_id = $1
) AND order_id = $2`
)

type DbJobRepo struct {
	db *sqlx.DB
}

func NewDbJobRepo(db *sqlx.DB) *DbJobRepo {
	return &DbJobRepo{db: db}
}

func (s DbJobRepo) CreateJobs(ctx context.Context, orderId string, process *domain.Process) error {
	if tx, ok := TransactionFromContext(ctx); ok {
		jobTraceStr, err := tracing.SpanStrFromContext(ctx)
		if err != nil {
			return err
		}
		for _, task := range process.Tasks {
			// Calculate count of relations
			readyRequired := 0
			for _, relation := range process.TaskRelations {
				if relation.ChildId == task.Id {
					readyRequired = readyRequired + 1
				}
			}
			if _, err := tx.ExecContext(ctx, createJobs, process.Id, task.Id, task.Category, task.Action, orderId,
				task.ReadMappingId, readyRequired, jobTraceStr); err != nil {
				return fmt.Errorf("can't create job (%s): %v", task.Id, err)
			}
		}
		return nil
	}
	return errors.New("can't create jobs without transaction")
}

func (s DbJobRepo) GetReadyJobs(ctx context.Context, jobLimit int) ([]domain.Job, error) {
	var jobs []domain.Job
	if err := s.db.SelectContext(ctx, &jobs, getReadyJobs, jobLimit); err != nil {
		return nil, fmt.Errorf("can't get ready jobs: %v", err)
	}
	return jobs, nil
}

func (s DbJobRepo) CompleteJob(ctx context.Context, taskId, orderId string) error {
	if tx, ok := TransactionFromContext(ctx); ok {
		if _, err := tx.ExecContext(ctx, completeJob, taskId, orderId); err != nil {
			return fmt.Errorf("can't complete job (%s, %s): %v", taskId, orderId, err)
		}
		if _, err := tx.ExecContext(ctx, completeRelatedJobs, taskId, orderId); err != nil {
			return fmt.Errorf("can't complete related jobs (%s, %s): %v", taskId, orderId, err)
		}
		return nil
	}
	return errors.New("can't complete job without transaction")
}
