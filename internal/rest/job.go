package rest

import (
	"context"
	"encoding/json"
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type JobRestHandler struct {
	orderService domain.OrderService
}

func NewJobRestHandler(orderService domain.OrderService) *JobRestHandler {
	return &JobRestHandler{orderService: orderService}
}

func (h JobRestHandler) Register(router *gin.Engine) {
	group := router.Group("/job")
	group.POST("/complete", h.completeJob)
}

// CompleteJob godoc
// @Summary Complete Job
// @Description Method to complete job
// @Tags Job
// @Accept json
// @Produce json
// @Param complete_job_message body domain.JobCompleteMessage true "Complete Job Message"
// @Success 200
// @Failure 500 {object} domain.Error
// @Router /job/complete [post]
func (h JobRestHandler) completeJob(c *gin.Context) {
	var obj domain.JobCompleteMessage
	if err := c.BindJSON(&obj); err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, E(err))
		return
	}
	if err := h.orderService.CompleteJob(c.Request.Context(), obj.TaskId, obj.OrderId); err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, E(err))
	}
}

type JobCompleteRestClient struct {
	baseUrl string
	client  *retryablehttp.Client
}

func NewJobCompleteRestClient(baseUrl string, client *retryablehttp.Client) domain.JobCompleteClient {
	return &JobCompleteRestClient{baseUrl: baseUrl, client: client}
}

func (c JobCompleteRestClient) Complete(ctx context.Context, msg *domain.JobCompleteMessage) error {
	const op = "JobCompleteRestClient.Complete"

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return domain.E(op, fmt.Sprintf("can't marshal request (%s, %s)", msg.TaskId, msg.OrderId), err)
	}

	_, err = Send(ctx, c.client, c.baseUrl+"/complete/", http.MethodPost, msgBytes)
	if err != nil {
		return domain.E(op, fmt.Sprintf("can't send request (%s, %s)", msg.TaskId, msg.OrderId), err)
	}

	// TDB Check status code

	return nil
}

type JobStartRestClient struct {
	client *retryablehttp.Client
}

func NewJobStartRestClient(client *retryablehttp.Client) domain.JobStartClient {
	return &JobStartRestClient{client: client}
}

func (c JobStartRestClient) Start(ctx context.Context, dest string, msg *domain.JobStartMessage) error {
	const op = "JobStartRestClient.Start"

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return domain.E(op, fmt.Sprintf("can't marshal request (%s, %s)", msg.TaskId, msg.OrderId), err)
	}

	_, err = Send(ctx, c.client, dest, http.MethodPost, msgBytes)
	if err != nil {
		return domain.E(op, fmt.Sprintf("can't send request (%s, %s)", msg.TaskId, msg.OrderId), err)
	}

	// TDB Check status code

	return nil
}
