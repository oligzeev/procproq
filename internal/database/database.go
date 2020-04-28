package database

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"example.com/oligzeev/pp-gin/internal/config"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

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
func DbConnect(cfg config.DbConfig) (*sqlx.DB, error) {
	cs := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DbName)
	log.Debugf("Connect to database: %s\n", cs)

	var db *sqlx.DB
	var err error
	if db, err = sqlx.Connect("pgx", cs); err != nil {
		return nil, fmt.Errorf("can't establish database connection: %v", err)
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
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("can't begin transaction: %v", err)
	}
	txCtx := WithTransaction(ctx, tx)
	result, err := f(txCtx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, fmt.Errorf("can't rollback transaction: %v", err)
		}
		return nil, fmt.Errorf("transaction has rolled back: %v", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("can't commit transaction: %v", err)
	}
	return result, nil
}
