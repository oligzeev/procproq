package tracing

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/opentracing/opentracing-go"
)

type SpanJobRepo struct {
	repo domain.JobRepo
}

func NewSpanJobRepo(repo domain.JobRepo) *SpanJobRepo {
	return &SpanJobRepo{repo: repo}
}

func (s SpanJobRepo) CreateJobs(ctx context.Context, orderId string, process *domain.Process) error {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "JobRepo.CreateJobs")
	defer span.Finish()
	return s.repo.CreateJobs(spanCtx, orderId, process)
}

func (s SpanJobRepo) GetReadyJobs(ctx context.Context, jobLimit int) ([]domain.Job, error) {
	// span, spanCtx := opentracing.StartSpanFromContext(ctx, "JobRepo.GetReadyJobs")
	// defer span.Finish()
	// return s.repo.GetReadyJobs(spanCtx, jobLimit)
	return s.repo.GetReadyJobs(ctx, jobLimit)
}

func (s SpanJobRepo) CompleteJob(ctx context.Context, taskId, orderId string) error {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "JobRepo.CompleteJob")
	defer span.Finish()
	return s.repo.CompleteJob(spanCtx, taskId, orderId)
}
