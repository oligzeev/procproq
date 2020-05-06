package rest

import (
	"example.com/oligzeev/pp-gin/internal/domain"
	"example.com/oligzeev/pp-gin/internal/tracing"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type StubRestHandler struct {
	jobCompleteClient domain.JobCompleteClient
}

func NewStubRestHandler(jobCompleteClient domain.JobCompleteClient) *StubRestHandler {
	return &StubRestHandler{jobCompleteClient: jobCompleteClient}
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
		log.Error(err)
		c.JSON(http.StatusInternalServerError, E(err))
		return
	}
	go func() {
		asyncSpan, asyncCtx := tracing.FollowNewSpanFromContext(span, "Send complete job message")
		defer asyncSpan.Finish()

		err := h.jobCompleteClient.Complete(asyncCtx, &domain.JobCompleteMessage{TaskId: startJob.TaskId, OrderId: startJob.OrderId})
		if err != nil {
			log.Error(err)
		}
	}()
	c.Status(http.StatusOK)
}
