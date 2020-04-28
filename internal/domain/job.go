package domain

import (
	"context"
)

const (
	HttpTaskCategory int = iota
)

type JobStartMessage struct {
	TaskId  string `json:"taskId"`
	OrderId string `json:"orderId"`
	Body    Body   `json:"body"`
}

type JobCompleteMessage struct {
	TaskId  string `json:"taskId"`
	OrderId string `json:"orderId"`
}

type Job struct {
	TaskId        string `json:"taskId" db:"task_id"`
	Category      int    `json:"category" db:"category"`
	Action        string `json:"action" db:"action"`
	OrderId       string `json:"orderId" db:"order_id"`
	ReadMappingId string `json:"readMappingId" db:"read_mapping_id"`
	Trace         string `json:"trace" db:"trace"`
}

type JobRepo interface {
	CreateJobs(ctx context.Context, orderId string, process *Process) error
	GetReadyJobs(ctx context.Context, jobLimit int) ([]Job, error)
	CompleteJob(ctx context.Context, taskId, orderId string) error
}
