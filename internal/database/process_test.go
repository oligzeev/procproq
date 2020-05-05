package database

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessRepo_GetAll_Success(t *testing.T) {
	assert := assert.New(t)

	var processes []Process
	var tasks []Task
	var relations []TaskRelation

	mockDB := new(MockDB)
	mockDB.On("SelectContext", testCtx, &processes, getProcesses,
		[]interface{}(nil)).Return(nil)
	mockDB.On("SelectContext", testCtx, &tasks, getTasks,
		[]interface{}(nil)).Return(nil)
	mockDB.On("SelectContext", testCtx, &relations, getTaskRelations,
		[]interface{}(nil)).Return(nil)

	repo := RDBProcessRepo{db: mockDB}
	err := repo.GetAll(testCtx, &processes)
	assert.Nil(err)
}

func TestProcessRepo_DeleteById_Success(t *testing.T) {
	const id = "1"
	assert := assert.New(t)

	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(1, nil)

	mockDB := new(MockDB)
	txCtx := WithTransaction(testCtx, mockDB)

	mockDB.On("ExecContext", txCtx, deleteProcessById, []interface{}{id}).Return(mockResult, nil)
	mockDB.On("ExecContext", txCtx, deleteTasksByProcessId,
		[]interface{}{id}).Return(nil, nil)
	mockDB.On("ExecContext", txCtx, deleteTaskRelationsByProcessId,
		[]interface{}{id}).Return(nil, nil)

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
