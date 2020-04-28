package rest

import (
	"encoding/json"
	"example.com/oligzeev/pp-gin/internal/config"
	"example.com/oligzeev/pp-gin/internal/domain"
	"example.com/oligzeev/pp-gin/internal/tracing"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

type StubRestHandler struct {
	client             *retryablehttp.Client
	completeJobTimeout time.Duration
	responseUrl        string
}

func NewStubRestHandler(cfg config.StubConfig) *StubRestHandler {
	client := retryablehttp.NewClient()
	client.RetryMax = cfg.SendJobRetriesMax
	return &StubRestHandler{
		client:             client,
		completeJobTimeout: time.Duration(cfg.SendJobTimeoutSec),
		responseUrl:        cfg.ResponseUrl,
	}
}

func (h StubRestHandler) Register(router *gin.Engine) {
	group := router.Group("/stub")
	group.POST("/start", h.start)
}

func (h StubRestHandler) start(c *gin.Context) {
	ctx := c.Request.Context()
	span := opentracing.SpanFromContext(ctx)

	var startJob domain.JobStartMessage
	if err := c.BindJSON(&startJob); err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	go func() {
		asyncSpan, asyncCtx := tracing.FollowNewSpanFromContext(span, "Send complete job message")
		defer asyncSpan.Finish()

		taskId := startJob.TaskId
		orderId := startJob.OrderId

		msgBytes, err := json.Marshal(&domain.JobCompleteMessage{TaskId: taskId, OrderId: orderId})
		if err != nil {
			log.Errorf("can't marshal complete job message (%s, %s): %v", taskId, orderId, err)
			return
		}
		//response, err := h.sendCompleteJobMessage(asyncCtx, msgBytes)
		response, err := Send(asyncCtx, h.client, h.responseUrl, http.MethodPost, msgBytes)
		if err != nil {
			log.Errorf("can't send complete job message (%s, %s): %v", taskId, orderId, err)
			return
		}
		if response.ContentLength == 0 {
			log.Debugf("complete job (%s, %s): %v", taskId, orderId, response.Status)
		} else {
			resBytes, err := ioutil.ReadAll(response.Body)
			if resBytes != nil {
				log.Errorf("can't read complete job message (%s, %s): %v", taskId, orderId, err)
				return
			}
			log.Debugf("complete job (%s, %s): %v, %s", taskId, orderId, response.Status, string(resBytes))
		}
	}()
	c.Status(http.StatusOK)
}
