package tracing

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/opentracing/opentracing-go"
)

type SpanOrderRepo struct {
	repo domain.OrderRepo
}

func NewSpanOrderRepo(repo domain.OrderRepo) *SpanOrderRepo {
	return &SpanOrderRepo{repo: repo}
}

func (s SpanOrderRepo) GetAll(ctx context.Context) ([]domain.Order, error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "OrderRepo.GetAll")
	defer span.Finish()
	return s.repo.GetAll(spanCtx)
}

func (s SpanOrderRepo) Create(ctx context.Context, obj *domain.Order) (*domain.Order, error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "OrderRepo.Create")
	defer span.Finish()
	return s.repo.Create(spanCtx, obj)
}

func (s SpanOrderRepo) GetById(ctx context.Context, id string) (*domain.Order, error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "OrderRepo.GetById")
	defer span.Finish()
	return s.repo.GetById(spanCtx, id)
}

func (s SpanOrderRepo) DeleteById(ctx context.Context, id string) error {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "OrderRepo.DeleteById")
	defer span.Finish()
	return s.repo.DeleteById(spanCtx, id)
}
