package service

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
	"example.com/oligzeev/pp-gin/internal/tracing"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

type JobScheduler struct {
	jobRepo            database.JobRepo
	orderService       domain.OrderService
	readMappingService domain.ReadMappingService
	period             time.Duration
	jobLimit           int
	startJobClient     domain.JobStartClient
}

func NewJobScheduler(
	cfg domain.SchedulerConfig,
	jobService database.JobRepo,
	orderService domain.OrderService,
	readMappingRepo domain.ReadMappingService,
	startJobClient domain.JobStartClient,
) *JobScheduler {
	return &JobScheduler{
		jobRepo:            jobService,
		orderService:       orderService,
		readMappingService: readMappingRepo,
		period:             cfg.PeriodSec,
		jobLimit:           cfg.JobLimit,
		startJobClient:     startJobClient,
	}
}

func (s JobScheduler) Start(groupCtx context.Context) error {
	const op = "JobScheduler.Start"

	log.Tracef("%s: starting", op)
	for {
		select {
		case <-groupCtx.Done():
			log.Tracef("%s: exit", op)
			return groupCtx.Err()
		default:
			s.schedule()
			time.Sleep(s.period * time.Second)
		}
	}
}

func (s JobScheduler) schedule() {
	const op = "JobScheduler.Schedule"

	log.Tracef("%s: get ready jobs", op)
	var jobs []database.Job
	if err := s.jobRepo.GetReadyJobs(context.Background(), s.jobLimit, &jobs); err != nil {
		log.Error(domain.E(op, "can't get ready jobs", err))
		return
	}

	log.Tracef("%s: jobs execution (%v)", op, len(jobs))
	for _, job := range jobs {
		if err := s.processJob(&job); err != nil {
			log.Error(err)
		}
	}
	log.Tracef("%s: finished (%v)", op, len(jobs))
}

func (s JobScheduler) processJob(job *database.Job) error {
	const op = "JobScheduler.ProcessJob"

	// Propagate span from job or use background
	span, spanCtx, err := tracing.StartContextFromSpanStr(context.Background(), "JobScheduler.ProcessJob", job.Trace)
	defer span.Finish()
	if err != nil {
		spanCtx = context.Background()
		log.Warn(domain.E(op, "can't extract span context from job, skip span context", err))
	}

	orderId := job.OrderId
	taskId := job.TaskId
	mappingId := job.ReadMappingId

	// Build start message body
	body, err := s.buildStartJobBody(spanCtx, orderId, mappingId)
	if err != nil {
		return domain.E(op, fmt.Sprintf("can't build start message (%s, %s)", taskId, orderId), err)
	}

	// Send start message
	if job.Category == domain.HttpTaskCategory {
		var startMsg = domain.JobStartMessage{TaskId: job.TaskId, OrderId: job.OrderId, Body: body}
		if err = s.startJobClient.Start(spanCtx, job.Action, &startMsg); err != nil {
			return domain.E(op, fmt.Sprintf("can't send start message (%s, %s)", taskId, orderId), err)
		}
		log.Tracef("%s: start completed (%s, %s)", op, taskId, orderId)
	} else {
		log.Tracef("%s: start skipped (%s, %s)", op, taskId, orderId)
	}
	return nil
}

func (s JobScheduler) buildStartJobBody(ctx context.Context, orderId, mappingId string) (domain.Body, error) {
	const op = "JobScheduler.BuildStartJobMessage"

	var order domain.Order
	if err := s.orderService.GetOrderById(ctx, orderId, &order); err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't get order (%s)", orderId), err)
	}
	var mapping domain.ReadMapping
	if err := s.readMappingService.GetById(ctx, mappingId, &mapping); err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't get read mapping (%s)", mappingId), err)
	}
	var result = make(domain.Body)
	for key, tasksPath := range mapping.PreparedBody {
		value, err := tasksPath(ctx, map[string]interface{}(order.Body))
		if err != nil {
			return nil, domain.E(op, fmt.Sprintf("can't evaluate value (%s)", value), err)
		}
		result[key] = value
	}
	return result, nil
}
