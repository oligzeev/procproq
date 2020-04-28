package tracing

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/opentracing/opentracing-go"
)

type SpanReadMappingRepo struct {
	repo domain.ReadMappingRepo
}

func NewSpanReadMappingRepo(repo domain.ReadMappingRepo) *SpanReadMappingRepo {
	return &SpanReadMappingRepo{repo: repo}
}

func (s SpanReadMappingRepo) GetAll(ctx context.Context) ([]domain.ReadMapping, error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "ReadMappingRepo.GetAll")
	defer span.Finish()
	return s.repo.GetAll(spanCtx)
}

func (s SpanReadMappingRepo) Create(ctx context.Context, obj *domain.ReadMapping) (*domain.ReadMapping, error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "ReadMappingRepo.Create")
	defer span.Finish()
	return s.repo.Create(spanCtx, obj)
}

func (s SpanReadMappingRepo) GetById(ctx context.Context, id string) (*domain.ReadMapping, error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "ReadMappingRepo.GetById")
	defer span.Finish()
	return s.repo.GetById(spanCtx, id)
}

func (s SpanReadMappingRepo) DeleteById(ctx context.Context, id string) error {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "ReadMappingRepo.DeleteById")
	defer span.Finish()
	return s.repo.DeleteById(spanCtx, id)
}
