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

type OrderEntity struct {
	Id        string `db:"order_id"`
	ProcessId string `db:"process_id"`
	Body      Body   `db:"body"`
}

func toOrder(obj *OrderEntity) *domain.Order {
	return &domain.Order{Id: obj.Id, ProcessId: obj.ProcessId, Body: domain.Body(obj.Body)}
}

func toOrders(arr []OrderEntity) []domain.Order {
	result := make([]domain.Order, len(arr))
	for i, obj := range arr {
		result[i].Id = obj.Id
		result[i].ProcessId = obj.ProcessId
		result[i].Body = domain.Body(obj.Body)
	}
	return result
}

// OrderRepo via postgres database
type DbOrderRepo struct {
	db *sqlx.DB
}

func NewDbOrderRepo(db *sqlx.DB) *DbOrderRepo {
	return &DbOrderRepo{db: db}
}

// Get all orders
func (s DbOrderRepo) GetAll(ctx context.Context) ([]domain.Order, error) {
	var order []OrderEntity
	if err := s.db.SelectContext(ctx, &order, getOrders); err != nil {
		return nil, fmt.Errorf("can't get orders: %v", err)
	}
	return toOrders(order), nil
}

// Create order
func (s DbOrderRepo) Create(ctx context.Context, obj *domain.Order) (*domain.Order, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("can't generate uuid: %v", err)
	}
	obj.Id = id.String()

	if tx, ok := TransactionFromContext(ctx); ok {
		_, err = tx.ExecContext(ctx, createOrder, obj.Id, obj.ProcessId, Body(obj.Body))
	} else {
		_, err = s.db.ExecContext(ctx, createOrder, obj.Id, obj.ProcessId, Body(obj.Body))
	}
	if err != nil {
		return nil, fmt.Errorf("can't create order (%s, %s): %v", obj.Id, obj.ProcessId, err)
	}
	return obj, nil
}

// Get order by Id
func (s DbOrderRepo) GetById(ctx context.Context, id string) (*domain.Order, error) {
	var result OrderEntity
	if err := s.db.GetContext(ctx, &result, getOrderById, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("can't get order (%s): %v", id, err)
	}
	return toOrder(&result), nil
}

// Delete order by Id
func (s DbOrderRepo) DeleteById(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, deleteOrderById, id); err != nil {
		return fmt.Errorf("can't delete order (%s): %v", id, err)
	}
	return nil
}
