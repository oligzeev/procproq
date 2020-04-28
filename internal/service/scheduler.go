package service

import (
	"context"
	"encoding/json"
	"example.com/oligzeev/pp-gin/internal/config"
	"example.com/oligzeev/pp-gin/internal/domain"
	"example.com/oligzeev/pp-gin/internal/rest"
	"example.com/oligzeev/pp-gin/internal/tracing"
	"github.com/PaesslerAG/gval"
	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const ErrorPrefix = "error during job scheduling"

type JobScheduler struct {
	jobService      domain.JobRepo
	orderRepo       domain.OrderRepo
	readMappingRepo domain.ReadMappingRepo
	client          *retryablehttp.Client
	period          time.Duration
	sendJobTimeout  time.Duration
	jobLimit        int
}

func NewJobScheduler(cfg config.SchedulerConfig, jobService domain.JobRepo, orderRepo domain.OrderRepo,
	readMappingRepo domain.ReadMappingRepo) *JobScheduler {

	client := retryablehttp.NewClient()
	client.RetryMax = cfg.SendJobRetriesMax
	return &JobScheduler{
		jobService:      jobService,
		orderRepo:       orderRepo,
		readMappingRepo: readMappingRepo,
		client:          client,
		period:          time.Duration(cfg.PeriodSec),
		sendJobTimeout:  time.Duration(cfg.SendJobTimeoutSec),
		jobLimit:        cfg.JobLimit,
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
	log.Traceln("start jobs scheduler")

	// ctx, cancel := context.WithTimeout(context.Background(), rest.RepoTimeoutSec*time.Second)
	// defer cancel()

	jobs, err := s.jobService.GetReadyJobs(context.Background(), s.jobLimit)
	if err != nil {
		log.Errorf(ErrorPrefix+", can't get ready jobs: %v", err)
		return
	}

	log.Tracef("start jobs execution: %v", len(jobs))
	for _, job := range jobs {
		span, spanCtx, err := tracing.StartContextFromSpanStr(context.Background(), "Scheduler start job", job.Trace)
		if err != nil {
			spanCtx = context.Background()
			log.Errorf("can't extract span context from job: %v", err)
		}
		orderId := job.OrderId
		taskId := job.TaskId
		mappingId := job.ReadMappingId

		msgBytes, err := s.buildStartJobMessage(spanCtx, job, orderId, mappingId)
		if err != nil {
			log.Errorf(ErrorPrefix+", can't build start job message (%s, %s): %v", taskId, orderId, err)
			continue
		}
		if job.Category == domain.HttpTaskCategory {
			if _, err := rest.Send(spanCtx, s.client, job.Action, http.MethodPost, msgBytes); err != nil {
				log.Errorf(ErrorPrefix+" (%s, %s, %s): %v", job.Action, taskId, orderId, err)
				continue
			}

			/*if err = s.sendStartJobMessage(spanCtx, job.Action, msgBytes); err != nil {
				log.Errorf(ErrorPrefix + " (%s, %s, %s): %v", job.Action, taskId, orderId, err)
				continue
			}*/
			log.Debugf("start job completed (%s, %s)", taskId, orderId)
		} else {
			log.Debugf("start job skipped (%s, %s)", taskId, orderId)
		}
		span.Finish()
	}
	log.Tracef("complete jobs execution: %v", len(jobs))
}

func (s JobScheduler) buildStartJobMessage(ctx context.Context, job domain.Job, orderId, mappingId string) ([]byte, error) {
	order, err := s.orderRepo.GetById(ctx, orderId)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get order (%s)", orderId)
	}
	mapping, err := s.readMappingRepo.GetById(ctx, mappingId)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get read mapping (%s)", mappingId)
	}
	body, err := s.buildStartJobBody(mapping, order.Body)
	if err != nil {
		return nil, errors.Wrap(err, "can't build start job body")
	}
	msgBytes, err := json.Marshal(&domain.JobStartMessage{
		TaskId:  job.TaskId,
		OrderId: job.OrderId,
		Body:    body,
	})
	if err != nil {
		return nil, errors.Wrap(err, "can't marshal message")
	}
	return msgBytes, nil
}

func (s JobScheduler) buildStartJobBody(mapping *domain.ReadMapping, orderBody domain.Body) (domain.Body, error) {
	var result = make(domain.Body)
	for key, value := range mapping.Body {
		builder := gval.Full(jsonpath.PlaceholderExtension())
		strValue := value.(string)
		if tasksPath, err := builder.NewEvaluable(strValue); err != nil {
			return nil, errors.Wrapf(err, "can't create new evaluator (%s)", value)
		} else if value, err := tasksPath(context.Background(), map[string]interface{}(orderBody)); err != nil {
			return nil, errors.Wrapf(err, "can't evaluate value (%s)", strValue)
		} else {
			result[key] = value
		}
	}
	return result, nil
}
