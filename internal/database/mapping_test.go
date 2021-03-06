package database

import (
	"database/sql"
	"errors"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadMappingRepo_GetAll_Success(t *testing.T) {
	assert := assert.New(t)

	var readMappings []ReadMapping

	mockDB := new(MockDB)
	mockDB.On("SelectContext", testCtx, &readMappings, getReadMappings,
		[]interface{}(nil)).Return(nil)

	repo := RDBReadMappingRepo{db: mockDB}
	err := repo.GetAll(testCtx, &readMappings)
	assert.Nil(err)
}

func TestReadMappingRepo_GetAll_Error(t *testing.T) {
	const op = "ReadMappingRepo.GetAll"

	assert := assert.New(t)

	var readMappings []ReadMapping
	mockErr := errors.New("mock error")

	mockDB := new(MockDB)
	mockDB.On("SelectContext", testCtx, &readMappings, getReadMappings, []interface{}(nil)).Return(mockErr)

	repo := RDBReadMappingRepo{db: mockDB}
	err := repo.GetAll(testCtx, &readMappings)

	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(mockErr, domainErr.Err)
}

func TestReadMappingRepo_Create_Success(t *testing.T) {
	assert := assert.New(t)

	mockBody := Body{"key1": "val1"}
	readMapping := &ReadMapping{Body: mockBody}
	mockUUID, _ := uuid.NewUUID()
	strUUID := mockUUID.String()

	mockDB := new(MockDB)
	mockDB.On("ExecContext", testCtx, createReadMapping,
		[]interface{}{strUUID, readMapping.Body}).Return(nil, nil)

	repo := RDBReadMappingRepo{db: mockDB, newUUIDFunc: func() (uuid.UUID, error) {
		return mockUUID, nil
	}}
	err := repo.Create(testCtx, readMapping)
	assert.Nil(err)
	assert.Equal(strUUID, readMapping.Id)
	assert.Equal(mockBody, readMapping.Body)
}

func TestReadMappingRepo_Create_UUID(t *testing.T) {
	const op = "ReadMappingRepo.Create"

	assert := assert.New(t)

	mockBody := Body{"key1": "val1"}
	readMapping := &ReadMapping{Body: mockBody}

	mockDB := new(MockDB)
	mockErr := errors.New("mock error")

	repo := RDBReadMappingRepo{db: mockDB, newUUIDFunc: func() (uuid.UUID, error) {
		var mockUUID uuid.UUID
		return mockUUID, mockErr
	}}
	err := repo.Create(testCtx, readMapping)

	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal("can't generate uuid", domainErr.Msg)
	assert.Equal(mockErr, domainErr.Err)
}

func TestReadMappingRepo_Create_Error(t *testing.T) {
	const op = "ReadMappingRepo.Create"

	assert := assert.New(t)

	mockBody := Body{"key1": "val1"}
	readMapping := &ReadMapping{Body: mockBody}
	mockUUID, _ := uuid.NewUUID()
	strUUID := mockUUID.String()

	mockDB := new(MockDB)
	mockErr := errors.New("mock error")
	mockDB.On("ExecContext", testCtx, createReadMapping,
		[]interface{}{strUUID, readMapping.Body}).Return(nil, mockErr)

	repo := RDBReadMappingRepo{db: mockDB, newUUIDFunc: func() (uuid.UUID, error) {
		return mockUUID, nil
	}}
	err := repo.Create(testCtx, readMapping)

	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(mockErr, domainErr.Err)
}

func TestReadMappingRepo_GetById_Success(t *testing.T) {
	const id = "1"
	assert := assert.New(t)

	readMapping := &ReadMapping{}

	mockDB := new(MockDB)
	mockDB.On("GetContext", testCtx, readMapping, getReadMappingById,
		[]interface{}{id}).Return(nil)

	repo := RDBReadMappingRepo{db: mockDB}
	err := repo.GetById(testCtx, id, readMapping)
	assert.Nil(err)
}

func TestReadMappingRepo_GetById_NotFound(t *testing.T) {
	const (
		id = "1"
		op = "ReadMappingRepo.GetById"
	)
	assert := assert.New(t)

	readMapping := &ReadMapping{}

	mockDB := new(MockDB)
	mockDB.On("GetContext", testCtx, readMapping, getReadMappingById,
		[]interface{}{id}).Return(sql.ErrNoRows)

	repo := RDBReadMappingRepo{db: mockDB}
	err := repo.GetById(testCtx, id, readMapping)

	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(domain.ErrNotFound, domainErr.Code)
}

func TestReadMappingRepo_GetById_Error(t *testing.T) {
	const (
		id = "1"
		op = "ReadMappingRepo.GetById"
	)
	assert := assert.New(t)

	mockErr := errors.New("mock error")

	readMapping := &ReadMapping{}

	mockDB := new(MockDB)
	mockDB.On("GetContext", testCtx, readMapping, getReadMappingById, []interface{}{id}).Return(mockErr)

	repo := RDBReadMappingRepo{db: mockDB}
	err := repo.GetById(testCtx, id, readMapping)

	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(mockErr, domainErr.Err)
}

func TestReadMappingRepo_DeleteById_Success(t *testing.T) {
	const id = "1"
	assert := assert.New(t)

	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(1, nil)

	mockDB := new(MockDB)
	mockDB.On("ExecContext", testCtx, deleteReadMappingById, []interface{}{id}).Return(mockResult, nil)

	repo := RDBReadMappingRepo{db: mockDB}
	err := repo.DeleteById(testCtx, id)
	assert.Nil(err)
}

func TestReadMappingRepo_DeleteById_NotFound(t *testing.T) {
	const (
		id = "1"
		op = "ReadMappingRepo.DeleteById"
	)
	assert := assert.New(t)

	mockResult := new(MockResult)
	mockResult.On("RowsAffected").Return(0, nil)

	mockDB := new(MockDB)
	mockDB.On("ExecContext", testCtx, deleteReadMappingById, []interface{}{id}).Return(mockResult, nil)

	repo := RDBReadMappingRepo{db: mockDB}
	err := repo.DeleteById(testCtx, id)

	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(domain.ErrNotFound, domainErr.Code)
}

func TestReadMappingRepo_DeleteById_Error(t *testing.T) {
	const (
		id = "1"
		op = "ReadMappingRepo.DeleteById"
	)
	assert := assert.New(t)

	mockErr := errors.New("mock error")

	mockDB := new(MockDB)
	mockDB.On("ExecContext", testCtx, deleteReadMappingById,
		[]interface{}{id}).Return(nil, mockErr)

	repo := RDBReadMappingRepo{db: mockDB}
	err := repo.DeleteById(testCtx, id)

	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(mockErr, domainErr.Err)
}
