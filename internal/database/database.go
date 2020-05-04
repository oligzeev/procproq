package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type NewUUIDFunc func() (uuid.UUID, error)

type DB interface {
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type Body domain.Body

func (b Body) Value() (driver.Value, error) {
	return json.Marshal(b)
}

func (b *Body) Scan(value interface{}) error {
	bodyBytes, ok := value.([]byte)
	if !ok {
		return errors.New("can't convert body to bytes")
	}
	return json.Unmarshal(bodyBytes, &b)
}

// For more usages of sqlx see https://jmoiron.github.io/sqlx/
func Connect(cfg domain.DbConfig) (*sqlx.DB, error) {
	const op = "Database.Connect"

	cs := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DbName)
	log.Debugf("Connect to database: %s", cs)

	var db *sqlx.DB
	var err error
	if db, err = sqlx.Connect("pgx", cs); err != nil {
		return nil, domain.E(op, "can't establish database connection", err)
	}
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	return db, nil
}

type txContextKey string

const txKey txContextKey = "transaction"

func WithTransaction(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func TransactionFromContext(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(*sqlx.Tx)
	return tx, ok
}

// Execute function in a database transaction
type txFunc func(txCtx context.Context) (interface{}, error)

func ExecTx(ctx context.Context, db *sqlx.DB, f txFunc) (interface{}, error) {
	const op = "Transaction.Exec"

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, domain.E(op, "can't begin transaction", err)
	}
	txCtx := WithTransaction(ctx, tx)
	result, err := f(txCtx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, domain.E(op, "can't rollback transaction", err)
		}
		return nil, domain.E(op, "transaction has rolled back", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, domain.E(op, "can't commit transaction", err)
	}
	return result, nil
}
