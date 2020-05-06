package domain

import "context"

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

type JobCompleteClient interface {
	Complete(ctx context.Context, msg *JobCompleteMessage) error
}

type JobStartClient interface {
	Start(ctx context.Context, dest string, msg *JobStartMessage) error
}
