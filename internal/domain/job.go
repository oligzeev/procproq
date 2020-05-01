package domain

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
