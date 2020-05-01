package domain

import "context"

type Process struct {
	Id            string         `json:"id"`
	Name          string         `json:"name"`
	Tasks         []Task         `json:"tasks"`
	TaskRelations []TaskRelation `json:"taskRelations"`
}

type Task struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Category      int    `json:"category"` // TBD Return string value instead of integer
	Action        string `json:"action"`
	ReadMappingId string `json:"readMappingId"`
}

type TaskRelation struct {
	ParentId string `json:"parentId"`
	ChildId  string `json:"childId"`
}

type ProcessService interface {
	GetAll(ctx context.Context) ([]Process, error)
	Create(ctx context.Context, process *Process) (*Process, error)
	GetById(ctx context.Context, id string) (*Process, error)
	DeleteById(ctx context.Context, id string) error
}
