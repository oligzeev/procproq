package domain

import "context"

type Process struct {
	Id            string         `json:"id" db:"process_id"`
	Name          string         `json:"name" db:"name"`
	Tasks         []Task         `json:"tasks"`
	TaskRelations []TaskRelation `json:"taskRelations"`
}

func (p *Process) AddTask(task *Task) {
	p.Tasks = append(p.Tasks, *task)
}

func (p *Process) AddTaskRelation(taskRelation *TaskRelation) {
	p.TaskRelations = append(p.TaskRelations, *taskRelation)
}

type Task struct {
	ProcessId     string `json:"-" db:"process_id" ` // TBD Remove ProcessId
	Id            string `json:"id" db:"task_id"`
	Name          string `json:"name" db:"name"`
	Category      int    `json:"category" db:"category"` // TBD Return string value instead of integer
	Action        string `json:"action" db:"action"`
	ReadMappingId string `json:"readMappingId" db:"read_mapping_id"`
}

type TaskRelation struct {
	ProcessId string `json:"-" db:"process_id"` // TBD remove ProcessId
	ParentId  string `json:"parentId" db:"parent_id"`
	ChildId   string `json:"childId" db:"child_id"`
}

type ProcessRepo interface {
	GetAll(ctx context.Context) ([]Process, error)
	Create(ctx context.Context, process *Process) (*Process, error)
	GetById(ctx context.Context, id string) (*Process, error)
	DeleteById(ctx context.Context, id string) error
}

type ProcessService interface {
	GetProcesses(ctx context.Context) ([]Process, error)
	CreateProcess(ctx context.Context, process *Process) (*Process, error)
	GetProcessById(ctx context.Context, id string) (*Process, error)
	DeleteProcessById(ctx context.Context, id string) error
	GetReadMappings(ctx context.Context) ([]ReadMapping, error)
	CreateReadMapping(ctx context.Context, mapping *ReadMapping) (*ReadMapping, error)
	GetReadMappingById(ctx context.Context, id string) (*ReadMapping, error)
	DeleteReadMappingById(ctx context.Context, id string) error
}
