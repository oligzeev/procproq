package tracing

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/opentracing/opentracing-go"
)

type SpanProcessRepo struct {
	repo domain.ProcessRepo
}

func NewSpanProcessRepo(repo domain.ProcessRepo) *SpanProcessRepo {
	return &SpanProcessRepo{repo: repo}
}

func (s SpanProcessRepo) GetAll(ctx context.Context) ([]domain.Process, error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "ProcessRepo.GetAll")
	defer span.Finish()
	return s.repo.GetAll(spanCtx)
}

func (s SpanProcessRepo) Create(ctx context.Context, obj *domain.Process) (*domain.Process, error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "ProcessRepo.Create")
	defer span.Finish()
	return s.repo.Create(spanCtx, obj)
}

func (s SpanProcessRepo) GetById(ctx context.Context, id string) (*domain.Process, error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "ProcessRepo.GetById")
	defer span.Finish()
	return s.repo.GetById(spanCtx, id)
}

func (s SpanProcessRepo) DeleteById(ctx context.Context, id string) error {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "ProcessRepo.DeleteById")
	defer span.Finish()
	return s.repo.DeleteById(spanCtx, id)
}
