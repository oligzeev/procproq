package tracing

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/opentracing/opentracing-go"
)

type SpanReadMappingService struct {
	service domain.ReadMappingService
}

func NewSpanReadMappingService(service domain.ReadMappingService) *SpanReadMappingService {
	return &SpanReadMappingService{service: service}
}

func (s SpanReadMappingService) GetAll(ctx context.Context) ([]domain.ReadMapping, error) {
	const op = "ReadMappingService.GetAll"
	span, spanCtx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()
	return s.service.GetAll(spanCtx)
}

func (s SpanReadMappingService) Create(ctx context.Context, obj *domain.ReadMapping) (*domain.ReadMapping, error) {
	const op = "ReadMappingService.Create"
	span, spanCtx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()
	return s.service.Create(spanCtx, obj)
}

func (s SpanReadMappingService) GetById(ctx context.Context, id string) (*domain.ReadMapping, error) {
	const op = "ReadMappingService.GetById"
	span, spanCtx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()
	return s.service.GetById(spanCtx, id)
}

func (s SpanReadMappingService) DeleteById(ctx context.Context, id string) error {
	const op = "ReadMappingService.DeleteById"
	span, spanCtx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()
	return s.service.DeleteById(spanCtx, id)
}
