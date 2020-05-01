package database

import (
	"context"
	"database/sql"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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

type OrderRepo struct {
	db *sqlx.DB
}

func NewOrderRepo(db *sqlx.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (s OrderRepo) GetAll(ctx context.Context) ([]Order, error) {
	const op = "OrderRepo.GetAll"

	var order []Order
	if err := s.db.SelectContext(ctx, &order, getOrders); err != nil {
		return nil, domain.E(op, err)
	}
	return order, nil
}

func (s OrderRepo) Create(ctx context.Context, obj *Order) (*Order, error) {
	const op = "OrderRepo.Create"

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, domain.E(op, "can't generate uuid", err)
	}
	obj.Id = id.String()

	if tx, ok := TransactionFromContext(ctx); ok {
		_, err = tx.ExecContext(ctx, createOrder, obj.Id, obj.ProcessId, Body(obj.Body))
	} else {
		_, err = s.db.ExecContext(ctx, createOrder, obj.Id, obj.ProcessId, Body(obj.Body))
	}
	if err != nil {
		return nil, domain.E(op, fmt.Errorf("can't create order (%s)", obj.ProcessId), err)
	}
	return obj, nil
}

func (s OrderRepo) GetById(ctx context.Context, id string) (*Order, error) {
	const op = "OrderRepo.GetById"

	var result Order
	if err := s.db.GetContext(ctx, &result, getOrderById, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.E(op, domain.ErrNotFound)
		}
		return nil, domain.E(op, err)
	}
	return &result, nil
}

// Delete order by Id
func (s OrderRepo) DeleteById(ctx context.Context, id string) error {
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
