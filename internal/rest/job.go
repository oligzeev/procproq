package rest

import (
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	if err := h.orderService.CompleteJob(c.Request.Context(), obj.TaskId, obj.OrderId); err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
	}
}
