package database

import (
	"context"
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

type Job struct {
	TaskId        string `db:"task_id"`
	Category      int    `db:"category"`
	Action        string `db:"action"`
	OrderId       string `db:"order_id"`
	ReadMappingId string `db:"read_mapping_id"`
	Trace         string `db:"trace"`
}

type JobRepo interface {
	CreateJobs(ctx context.Context, orderId string, process *Process) error
	GetReadyJobs(ctx context.Context, jobLimit int, jobs *[]Job) error
	CompleteJob(ctx context.Context, taskId, orderId string) error
}

type RDBJobRepo struct {
	db *sqlx.DB
}

func NewRDBJobRepo(db *sqlx.DB) JobRepo {
	return &RDBJobRepo{db: db}
}

func (s RDBJobRepo) CreateJobs(ctx context.Context, orderId string, process *Process) error {
	const op = "JobRepo.CreateJobs"

	if tx, ok := TransactionFromContext(ctx); ok {
		jobTraceStr, err := tracing.SpanStrFromContext(ctx)
		if err != nil {
			return domain.E(op, "can't get span string", err)
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

				return domain.E(op, fmt.Sprintf("can't create job (%s)", task.Id), err)
			}
		}
		return nil
	}
	return domain.E(op, "there's no active transaction")
}

func (s RDBJobRepo) GetReadyJobs(ctx context.Context, jobLimit int, jobs *[]Job) error {
	const op = "JobRepo.GetReadyJobs"

	if err := s.db.SelectContext(ctx, jobs, getReadyJobs, jobLimit); err != nil {
		return domain.E(op, err)
	}
	return nil
}

func (s RDBJobRepo) CompleteJob(ctx context.Context, taskId, orderId string) error {
	const op = "JobRepo.CompleteJob"

	if tx, ok := TransactionFromContext(ctx); ok {
		result, err := tx.ExecContext(ctx, completeJob, taskId, orderId)
		if err != nil {
			return domain.E(op, fmt.Sprintf("can't complete job (%s, %s)", taskId, orderId), err)
		}
		if count, _ := result.RowsAffected(); count == 0 {
			return domain.E(op, domain.ErrNotFound)
		}
		if _, err := tx.ExecContext(ctx, completeRelatedJobs, taskId, orderId); err != nil {
			return domain.E(op, fmt.Sprintf("can't complete related jobs (%s, %s)", taskId, orderId), err)
		}
		return nil
	}
	return domain.E(op, "there's no active transaction")
}
