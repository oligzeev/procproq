package database

import (
	"context"
	"database/sql"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
)

const (
	getOrders       = `SELECT order_id, process_id, body FROM pp_order`
	createOrder     = `INSERT INTO pp_order (order_id, process_id, body) VALUES ($1, $2, $3)`
	getOrderById    = `SELECT order_id, process_id, body FROM pp_order WHERE order_id = $1`
	deleteOrderById = `DELETE FROM pp_order WHERE order_id = $1`
)

type Order struct {
	Id        string `db:"order_id"`
	ProcessId string `db:"process_id"`
	Body      Body   `db:"body"`
}

type OrderRepo interface {
	Create(ctx context.Context, obj *Order) error
	GetAll(ctx context.Context, result *[]Order) error
	GetById(ctx context.Context, id string, result *Order) error
	DeleteById(ctx context.Context, id string) error
}

type RDBOrderRepo struct {
	db          DB
	newUUIDFunc NewUUIDFunc
}

func NewRDBOrderRepo(db DB, newUUIDFunc NewUUIDFunc) OrderRepo {
	return &RDBOrderRepo{db: db, newUUIDFunc: newUUIDFunc}
}

func (s RDBOrderRepo) GetAll(ctx context.Context, result *[]Order) error {
	const op = "OrderRepo.GetAll"

	if err := s.db.SelectContext(ctx, result, getOrders); err != nil {
		return domain.E(op, err)
	}
	return nil
}

func (s RDBOrderRepo) Create(ctx context.Context, obj *Order) error {
	const op = "OrderRepo.Create"

	id, err := s.newUUIDFunc()
	if err != nil {
		return domain.E(op, "can't generate uuid", err)
	}
	obj.Id = id.String()

	if tx, ok := TransactionFromContext(ctx); ok {
		_, err = tx.ExecContext(ctx, createOrder, obj.Id, obj.ProcessId, Body(obj.Body))
	} else {
		_, err = s.db.ExecContext(ctx, createOrder, obj.Id, obj.ProcessId, Body(obj.Body))
	}
	if err != nil {
		return domain.E(op, fmt.Errorf("can't create order (%s)", obj.ProcessId), err)
	}
	return nil
}

func (s RDBOrderRepo) GetById(ctx context.Context, id string, result *Order) error {
	const op = "OrderRepo.GetById"

	if err := s.db.GetContext(ctx, result, getOrderById, id); err != nil {
		if err == sql.ErrNoRows {
			return domain.E(op, domain.ErrNotFound)
		}
		return domain.E(op, err)
	}
	return nil
}

// Delete order by Id
func (s RDBOrderRepo) DeleteById(ctx context.Context, id string) error {
	const op = "OrderRepo.DeleteById"

	result, err := s.db.ExecContext(ctx, deleteOrderById, id)
	if err != nil {
		return domain.E(op, err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return domain.E(op, domain.ErrNotFound)
	}
	return nil
}
