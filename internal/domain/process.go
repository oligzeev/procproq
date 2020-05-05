package domain

import (
	"context"
)

func CloneProcess(from, to *Process) {
	to.Id = from.Id
	to.Name = from.Name
	to.Tasks = from.Tasks
	to.TaskRelations = from.TaskRelations
}

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
	GetAll(ctx context.Context, result *[]Process) error
	Create(ctx context.Context, obj *Process) error
	GetById(ctx context.Context, id string, result *Process) error
	DeleteById(ctx context.Context, id string) error
}
