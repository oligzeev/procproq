package tracing

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/opentracing/opentracing-go"
)

type SpanProcessService struct {
	service domain.ProcessService
}

func NewSpanProcessService(service domain.ProcessService) *SpanProcessService {
	return &SpanProcessService{service: service}
}

func (s SpanProcessService) GetAll(ctx context.Context) ([]domain.Process, error) {
	const op = "ProcessService.GetAll"
	span, spanCtx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()
	return s.service.GetAll(spanCtx)
}

func (s SpanProcessService) Create(ctx context.Context, obj *domain.Process) (*domain.Process, error) {
	const op = "ProcessService.Create"
	span, spanCtx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()
	return s.service.Create(spanCtx, obj)
}

func (s SpanProcessService) GetById(ctx context.Context, id string) (*domain.Process, error) {
	const op = "ProcessService.GetById"
	span, spanCtx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()
	return s.service.GetById(spanCtx, id)
}

func (s SpanProcessService) DeleteById(ctx context.Context, id string) error {
	const op = "ProcessService.DeleteById"
	span, spanCtx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()
	return s.service.DeleteById(spanCtx, id)
}
