package service

import (
	"context"
	"encoding/json"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
	"example.com/oligzeev/pp-gin/internal/rest"
	"example.com/oligzeev/pp-gin/internal/tracing"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type JobScheduler struct {
	jobRepo            database.JobRepo
	orderService       domain.OrderService
	readMappingService domain.ReadMappingService
	client             *retryablehttp.Client
	period             time.Duration
	sendJobTimeout     time.Duration
	jobLimit           int
}

func NewJobScheduler(cfg domain.SchedulerConfig, jobService database.JobRepo, orderService domain.OrderService,
	readMappingRepo domain.ReadMappingService) *JobScheduler {

	client := retryablehttp.NewClient()
	client.RetryMax = cfg.SendJobRetriesMax
	return &JobScheduler{
		jobRepo:            jobService,
		orderService:       orderService,
		readMappingService: readMappingRepo,
		client:             client,
		period:             time.Duration(cfg.PeriodSec),
		sendJobTimeout:     time.Duration(cfg.SendJobTimeoutSec),
		jobLimit:           cfg.JobLimit,
	}
}

func (s JobScheduler) Start() {
	go func() {
		for {
			s.schedule()
			time.Sleep(s.period * time.Second)
		}
	}()
}

func (s JobScheduler) schedule() {
	const op = "JobScheduler.Schedule"

	log.Trace(op + ": schedule")
	var jobs []database.Job
	if err := s.jobRepo.GetReadyJobs(context.Background(), s.jobLimit, &jobs); err != nil {
		log.Error(domain.E(op, "can't get ready jobs", err))
		return
	}
	log.Tracef(op+": jobs execution: %v", len(jobs))
	for _, job := range jobs {
		span, spanCtx, err := tracing.StartContextFromSpanStr(context.Background(), "Scheduler start job", job.Trace)
		if err != nil {
			spanCtx = context.Background()
			log.Warn(domain.E(op, "can't extract span context from job, skip span context", err))
		}
		orderId := job.OrderId
		taskId := job.TaskId
		mappingId := job.ReadMappingId

		msgBytes, err := s.buildStartJobMessage(spanCtx, job, orderId, mappingId)
		if err != nil {
			log.Error(domain.E(op, fmt.Sprintf("can't build start job message (%s, %s)", taskId, orderId), err))
			continue
		}
		if job.Category == domain.HttpTaskCategory {
			if _, err := rest.Send(spanCtx, s.client, job.Action, http.MethodPost, msgBytes); err != nil {
				log.Error(domain.E(op, fmt.Sprintf("can't send start job message (%s, %s, %s)", job.Action,
					taskId, orderId), err))
				continue
			}
			log.Tracef(op+": start job completed (%s, %s)", taskId, orderId)
		} else {
			log.Tracef(op+": start job skipped (%s, %s)", taskId, orderId)
		}
		span.Finish()
	}
	log.Tracef(op+": complete: %v", len(jobs))
}

func (s JobScheduler) buildStartJobMessage(ctx context.Context, job database.Job, orderId, mappingId string) ([]byte, error) {
	const op = "JobScheduler.BuildStartJobMessage"

	var order domain.Order
	if err := s.orderService.GetOrderById(ctx, orderId, &order); err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't get order (%s)", orderId), err)
	}
	var mapping domain.ReadMapping
	if err := s.readMappingService.GetById(ctx, mappingId, &mapping); err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't get read mapping (%s)", mappingId), err)
	}
	body, err := buildStartJobBody(ctx, &mapping, order.Body)
	if err != nil {
		return nil, domain.E(op, "can't build job body", err)
	}
	msgBytes, err := json.Marshal(&domain.JobStartMessage{
		TaskId:  job.TaskId,
		OrderId: job.OrderId,
		Body:    body,
	})
	if err != nil {
		return nil, domain.E(op, "can't marshal message", err)
	}
	return msgBytes, nil
}

func buildStartJobBody(ctx context.Context, mapping *domain.ReadMapping, orderBody domain.Body) (domain.Body, error) {
	const op = "JobScheduler.BuildStartJobBody"

	var result = make(domain.Body)
	for key, tasksPath := range mapping.PreparedBody {
		value, err := tasksPath(ctx, map[string]interface{}(orderBody))
		if err != nil {
			return nil, domain.E(op, fmt.Sprintf("can't evaluate value (%s)", value), err)
		}
		result[key] = value
	}
	return result, nil
}
