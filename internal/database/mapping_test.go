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
	mockDB.On("SelectContext", testCtx, &readMappings, GetReadMappings,
		[]interface{}(nil)).Return(nil)

	repo := ReadMappingRepo{db: mockDB}
	result, err := repo.GetAll(testCtx)
	assert.Nil(err)
	assert.Equal(readMappings, result)
}

func TestReadMappingRepo_GetAll_Error(t *testing.T) {
	const op = "ReadMappingRepo.GetAll"

	assert := assert.New(t)

	var readMappings []ReadMapping
	mockErr := errors.New("mock error")

	mockDB := new(MockDB)
	mockDB.On("SelectContext", testCtx, &readMappings, GetReadMappings, []interface{}(nil)).Return(mockErr)

	repo := ReadMappingRepo{db: mockDB}
	result, err := repo.GetAll(testCtx)
	assert.Nil(result)

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
	mockDB.On("ExecContext", testCtx, CreateReadMapping,
		[]interface{}{strUUID, readMapping.Body}).Return(nil, nil)

	repo := ReadMappingRepo{db: mockDB, newUUIDFunc: func() (uuid.UUID, error) {
		return mockUUID, nil
	}}
	result, err := repo.Create(testCtx, readMapping)
	assert.Nil(err)
	assert.NotNil(result)
	assert.Equal(strUUID, result.Id)
	assert.Equal(mockBody, result.Body)
}

func TestReadMappingRepo_Create_UUID(t *testing.T) {
	const op = "ReadMappingRepo.Create"

	assert := assert.New(t)

	mockBody := Body{"key1": "val1"}
	readMapping := &ReadMapping{Body: mockBody}

	mockDB := new(MockDB)
	mockErr := errors.New("mock error")

	repo := ReadMappingRepo{db: mockDB, newUUIDFunc: func() (uuid.UUID, error) {
		var mockUUID uuid.UUID
		return mockUUID, mockErr
	}}
	result, err := repo.Create(testCtx, readMapping)
	assert.Nil(result)

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
	mockDB.On("ExecContext", testCtx, CreateReadMapping,
		[]interface{}{strUUID, readMapping.Body}).Return(nil, mockErr)

	repo := ReadMappingRepo{db: mockDB, newUUIDFunc: func() (uuid.UUID, error) {
		return mockUUID, nil
	}}
	result, err := repo.Create(testCtx, readMapping)
	assert.Nil(result)

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
	mockDB.On("GetContext", testCtx, readMapping, GetReadMappingById,
		[]interface{}{id}).Return(nil)

	repo := ReadMappingRepo{db: mockDB}
	result, err := repo.GetById(testCtx, id)
	assert.Nil(err)
	assert.NotNil(result)
	assert.Equal(readMapping, result)
}

func TestReadMappingRepo_GetById_NotFound(t *testing.T) {
	const (
		id = "1"
		op = "ReadMappingRepo.GetById"
	)
	assert := assert.New(t)

	readMapping := &ReadMapping{}

	mockDB := new(MockDB)
	mockDB.On("GetContext", testCtx, readMapping, GetReadMappingById,
		[]interface{}{id}).Return(sql.ErrNoRows)

	repo := ReadMappingRepo{db: mockDB}
	result, err := repo.GetById(testCtx, id)
	assert.NotNil(err)
	assert.Nil(result)
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
	mockDB.On("GetContext", testCtx, readMapping, GetReadMappingById, []interface{}{id}).Return(mockErr)

	repo := ReadMappingRepo{db: mockDB}
	result, err := repo.GetById(testCtx, id)
	assert.NotNil(err)
	assert.Nil(result)
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
	mockDB.On("ExecContext", testCtx, DeleteReadMappingById, []interface{}{id}).Return(mockResult, nil)

	repo := ReadMappingRepo{db: mockDB}
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
	mockDB.On("ExecContext", testCtx, DeleteReadMappingById, []interface{}{id}).Return(mockResult, nil)

	repo := ReadMappingRepo{db: mockDB}
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
	mockDB.On("ExecContext", testCtx, DeleteReadMappingById,
		[]interface{}{id}).Return(nil, mockErr)

	repo := ReadMappingRepo{db: mockDB}
	err := repo.DeleteById(testCtx, id)
	assert.NotNil(err)
	domainErr := toError(t, op, err)
	assert.Equal(op, string(domainErr.Op))
	assert.Equal(mockErr, domainErr.Err)
}
