// go test -coverprofile database_test.out -cover example.com/oligzeev/pp-gin/internal/database
// go tool cover -html database_test.out
package database

import (
	"context"
	"database/sql"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/stretchr/testify/mock"
	"testing"
)

var testCtx = context.WithValue(context.Background(), "mock", "test")

func toError(t *testing.T, op string, err error) *domain.Error {
	result, ok := err.(*domain.Error)
	if !ok {
		t.Fatalf("%s returns %T instead of domain.Error", op, err)
	}
	return result
}

type MockDB struct {
	mock.Mock
}

func (m MockDB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	arguments := m.Called(ctx, dest, query, args)
	return arguments.Error(0)
}

func (m MockDB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	arguments := m.Called(ctx, dest, query, args)
	return arguments.Error(0)
}

func (m MockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	arguments := m.Called(ctx, query, args)
	result := arguments.Get(0)
	if result != nil {
		return result.(sql.Result), arguments.Error(1)
	} else {
		return nil, arguments.Error(1)
	}
}

type MockResult struct {
	mock.Mock
}

func (m MockResult) LastInsertId() (int64, error) {
	arguments := m.Called()
	return int64(arguments.Int(0)), arguments.Error(1)
}

func (m MockResult) RowsAffected() (int64, error) {
	arguments := m.Called()
	return int64(arguments.Int(0)), arguments.Error(1)
}
