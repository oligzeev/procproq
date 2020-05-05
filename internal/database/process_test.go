package database

import (
	"database/sql"
	"errors"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessRepo_GetAll_Success(t *testing.T) {
	assert := assert.New(t)

	var processes []Process
	var tasks []Task
	var relations []TaskRelation

	mockDB := new(MockDB)
	mockDB.On("SelectContext", testCtx, &processes, getProcesses, []interface{}(nil)).Return(nil)
	mockDB.On("SelectContext", testCtx, &tasks, getTasks, []interface{}(nil)).Return(nil)
	mockDB.On("SelectContext", testCtx, &relations, getTaskRelations, []interface{}(nil)).Return(nil)

	repo := RDBProcessRepo{db: mockDB}
	err := repo.GetAll(testCtx, &processes)
	assert.Nil(err)
}

func TestProcessRepo_GetById_Success(t *testing.T) {
	const id = "1"
	assert := assert.New(t)

	mockDB := new(MockDB)
	process := Process{}
	mockDB.On("GetContext", testCtx, &process, getProcessById, []interface{}{id}).Return(nil)
	mockDB.On("SelectContext", testCtx, &process.Tasks, getTasksByProcessId, []interface{}{id}).Return(nil)
	mockDB.On("SelectContext", testCtx, &process.TaskRelations, getTaskRelationsByProcessId, []interface{}{id}).Return(nil)

	repo := RDBProcessRepo{db: mockDB}
	err := repo.GetById(testCtx, id, &process)
	assert.Nil(err)
}

func TestProcessRepo_GetById_NotFound(t *testing.T) {
	const (
		id = "1"
		op = "ProcessRepo.GetById"
	)
	assert := assert.New(t)

	mockDB := new(MockDB)
	process := Process{}
	mockDB.On("GetContext", testCtx, &process, getProcessById, []interface{}{id}).Return(sql.ErrNoRows)

	repo := RDBProcessRepo{db: mockDB}
	err := repo.GetById(testCtx, id, &process)
	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(domain.ErrNotFound, domainErr.Code)
}

func TestProcessRepo_GetById_ProcessError(t *testing.T) {
	const (
		id         = "1"
		op         = "ProcessRepo.GetById"
		mockErrMsg = "mock error"
	)
	assert := assert.New(t)
	mockErr := errors.New(mockErrMsg)

	mockDB := new(MockDB)
	process := Process{}
	mockDB.On("GetContext", testCtx, &process, getProcessById, []interface{}{id}).Return(mockErr)

	repo := RDBProcessRepo{db: mockDB}
	err := repo.GetById(testCtx, id, &process)
	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(fmt.Sprintf("can't select process (%s)", id), domainErr.Msg)
	assert.NotNil(domainErr.Err)
	assert.Equal(mockErrMsg, domainErr.Err.Error())
}

func TestProcessRepo_GetById_TaskError(t *testing.T) {
	const (
		id         = "1"
		op         = "ProcessRepo.GetById"
		mockErrMsg = "mock error"
	)
	assert := assert.New(t)
	mockErr := errors.New(mockErrMsg)

	mockDB := new(MockDB)
	process := Process{}
	mockDB.On("GetContext", testCtx, &process, getProcessById, []interface{}{id}).Return(nil)
	mockDB.On("SelectContext", testCtx, &process.Tasks, getTasksByProcessId, []interface{}{id}).Return(mockErr)

	repo := RDBProcessRepo{db: mockDB}
	err := repo.GetById(testCtx, id, &process)
	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(fmt.Sprintf("can't select tasks (%s)", id), domainErr.Msg)
	assert.NotNil(domainErr.Err)
	assert.Equal(mockErrMsg, domainErr.Err.Error())
}

func TestProcessRepo_GetById_RelationError(t *testing.T) {
	const (
		id         = "1"
		op         = "ProcessRepo.GetById"
		mockErrMsg = "mock error"
	)
	assert := assert.New(t)
	mockErr := errors.New(mockErrMsg)

	mockDB := new(MockDB)
	process := Process{}
	mockDB.On("GetContext", testCtx, &process, getProcessById, []interface{}{id}).Return(nil)
	mockDB.On("SelectContext", testCtx, &process.Tasks, getTasksByProcessId, []interface{}{id}).Return(nil)
	mockDB.On("SelectContext", testCtx, &process.TaskRelations, getTaskRelationsByProcessId, []interface{}{id}).Return(mockErr)

	repo := RDBProcessRepo{db: mockDB}
	err := repo.GetById(testCtx, id, &process)
	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(fmt.Sprintf("can't select task relations (%s)", id), domainErr.Msg)
	assert.NotNil(domainErr.Err)
	assert.Equal(mockErrMsg, domainErr.Err.Error())
}

func TestProcessRepo_DeleteById_Success(t *testing.T) {
	const id = "1"
	assert := assert.New(t)

	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(1, nil)

	mockDB := new(MockDB)
	txCtx := WithTransaction(testCtx, mockDB)
	mockDB.On("ExecContext", txCtx, deleteProcessById, []interface{}{id}).Return(mockResult, nil)
	mockDB.On("ExecContext", txCtx, deleteTasksByProcessId, []interface{}{id}).Return(nil, nil)
	mockDB.On("ExecContext", txCtx, deleteTaskRelationsByProcessId, []interface{}{id}).Return(nil, nil)

	repo := RDBProcessRepo{db: mockDB}
	err := repo.DeleteById(txCtx, id)
	assert.Nil(err)
}

func TestProcessRepo_DeleteById_NoTx(t *testing.T) {
	const (
		op = "ProcessRepo.DeleteById"
		id = "1"
	)
	assert := assert.New(t)

	mockDB := new(MockDB)
	repo := RDBProcessRepo{db: mockDB}
	err := repo.DeleteById(testCtx, id)

	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal("there's no active transaction", domainErr.Msg)
}

func TestProcessRepo_DeleteById_NotFound(t *testing.T) {
	const (
		op = "ProcessRepo.DeleteById"
		id = "1"
	)
	assert := assert.New(t)

	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(0, nil)

	mockDB := new(MockDB)
	txCtx := WithTransaction(testCtx, mockDB)
	mockDB.On("ExecContext", txCtx, deleteProcessById, []interface{}{id}).Return(mockResult, nil)

	repo := RDBProcessRepo{db: mockDB}
	err := repo.DeleteById(txCtx, id)

	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(domain.ErrNotFound, domainErr.Code)
}

func TestProcessRepo_DeleteById_ProcessError(t *testing.T) {
	const (
		op         = "ProcessRepo.DeleteById"
		id         = "1"
		mockErrMsg = "mock error"
	)
	assert := assert.New(t)

	mockErr := errors.New(mockErrMsg)

	mockDB := new(MockDB)
	txCtx := WithTransaction(testCtx, mockDB)
	mockDB.On("ExecContext", txCtx, deleteProcessById, []interface{}{id}).Return(nil, mockErr)

	repo := RDBProcessRepo{db: mockDB}
	err := repo.DeleteById(txCtx, id)

	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(fmt.Sprintf("can't delete process (%s)", id), domainErr.Msg)
	assert.NotNil(domainErr.Err)
	assert.Equal(mockErrMsg, domainErr.Err.Error())
}

func TestProcessRepo_DeleteById_TaskError(t *testing.T) {
	const (
		op         = "ProcessRepo.DeleteById"
		id         = "1"
		mockErrMsg = "mock error"
	)
	assert := assert.New(t)
	mockErr := errors.New(mockErrMsg)

	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(1, nil)

	mockDB := new(MockDB)
	txCtx := WithTransaction(testCtx, mockDB)
	mockDB.On("ExecContext", txCtx, deleteProcessById, []interface{}{id}).Return(mockResult, nil)
	mockDB.On("ExecContext", txCtx, deleteTasksByProcessId, []interface{}{id}).Return(nil, mockErr)

	repo := RDBProcessRepo{db: mockDB}
	err := repo.DeleteById(txCtx, id)

	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(fmt.Sprintf("can't delete tasks (%s)", id), domainErr.Msg)
	assert.NotNil(domainErr.Err)
	assert.Equal(mockErrMsg, domainErr.Err.Error())
}

func TestProcessRepo_DeleteById_RelationError(t *testing.T) {
	const (
		op         = "ProcessRepo.DeleteById"
		id         = "1"
		mockErrMsg = "mock error"
	)
	assert := assert.New(t)
	mockErr := errors.New(mockErrMsg)

	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(1, nil)

	mockDB := new(MockDB)
	txCtx := WithTransaction(testCtx, mockDB)
	mockDB.On("ExecContext", txCtx, deleteProcessById, []interface{}{id}).Return(mockResult, nil)
	mockDB.On("ExecContext", txCtx, deleteTasksByProcessId, []interface{}{id}).Return(mockResult, nil)
	mockDB.On("ExecContext", txCtx, deleteTaskRelationsByProcessId, []interface{}{id}).Return(nil, mockErr)

	repo := RDBProcessRepo{db: mockDB}
	err := repo.DeleteById(txCtx, id)

	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(fmt.Sprintf("can't delete task relations (%s)", id), domainErr.Msg)
	assert.NotNil(domainErr.Err)
	assert.Equal(mockErrMsg, domainErr.Err.Error())
}
