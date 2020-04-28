package rest

import (
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ProcessRestHandler struct {
	processService domain.ProcessService
}

func NewProcessRestHandler(processService domain.ProcessService) *ProcessRestHandler {
	return &ProcessRestHandler{processService: processService}
}

func (h ProcessRestHandler) Register(router *gin.Engine) {
	group := router.Group("/process")
	group.GET("/:"+ParamId, h.getProcessById)
	group.GET("/", h.getProcesses)
	group.DELETE("/:"+ParamId, h.deleteProcessById)
	group.POST("/", h.createProcess)
}

// GetProcessById godoc
// @Summary Get Process by Id
// @Description Method to get Process by id
// @Tags Process
// @Accept json
// @Produce json
// @Param id path string true "Process Id"
// @Success 200 {object} domain.Process
// @Failure 500 {object} domain.Error
// @Router /process/{id} [get]
func (h ProcessRestHandler) getProcessById(c *gin.Context) {
	id := c.Param(ParamId)
	result, err := h.processService.GetProcessById(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	if result != nil {
		c.JSON(http.StatusOK, result)
	} else {
		c.Status(http.StatusNotFound)
	}
}

// GetProcesses godoc
// @Summary Get Processes
// @Description Method to get all processes
// @Tags Process
// @Accept json
// @Produce json
// @Success 200 {array} domain.Process
// @Failure 500 {object} domain.Error
// @Router /process [get]
func (h ProcessRestHandler) getProcesses(c *gin.Context) {
	results, err := h.processService.GetProcesses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

// DeleteProcessById godoc
// @Summary Delete process by Id
// @Description Method to delete process by id
// @Tags Process
// @Accept json
// @Produce json
// @Param id path string true "Process Id"
// @Success 200
// @Failure 500 {object} domain.Error
// @Router /process/{id} [delete]
func (h ProcessRestHandler) deleteProcessById(c *gin.Context) {
	id := c.Param(ParamId)
	err := h.processService.DeleteProcessById(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
	}
}

// CreateProcess godoc
// @Summary Create Process
// @Description Method to create process
// @Tags Process
// @Accept json
// @Produce json
// @Param process body domain.Process true "Process (without id)"
// @Success 200 {object} domain.Process
// @Failure 500 {object} domain.Error
// @Router /process [post]
func (h ProcessRestHandler) createProcess(c *gin.Context) {
	var obj domain.Process
	if err := c.BindJSON(&obj); err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	result, err := h.processService.CreateProcess(c.Request.Context(), &obj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	c.JSON(http.StatusOK, result)
}
