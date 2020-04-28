package rest

import (
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

type MappingRestHandler struct {
	processService domain.ProcessService
}

func NewMappingRestHandler(processService domain.ProcessService) *MappingRestHandler {
	return &MappingRestHandler{processService: processService}
}

func (h MappingRestHandler) Register(router *gin.Engine) {
	group := router.Group("/mapping")
	group.GET("/:"+ParamId, h.getReadMappingById)
	group.GET("/", h.getReadMappings)
	group.DELETE("/:"+ParamId, h.deleteReadMappingById)
	group.POST("/", h.createReadMapping)
}

// GetReadMappingById godoc
// @Summary Get Read Mapping by Id
// @Description Method to get read mapping by id
// @Tags Read Mapping
// @Accept json
// @Produce json
// @Param id path string true "Read Mapping Id"
// @Success 200 {object} domain.ReadMapping
// @Failure 500 {object} domain.Error
// @Router /mapping/{id} [get]
func (h MappingRestHandler) getReadMappingById(c *gin.Context) {
	id := c.Param(ParamId)
	result, err := h.processService.GetReadMappingById(c.Request.Context(), id)
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

// GetReadMappingById godoc
// @Summary Get Read Mappings
// @Description Method to get all read mappings
// @Tags Read Mapping
// @Accept json
// @Produce json
// @Success 200 {array} domain.ReadMapping
// @Failure 500 {object} domain.Error
// @Router /mapping [get]
func (h MappingRestHandler) getReadMappings(c *gin.Context) {
	results, err := h.processService.GetReadMappings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

// DeleteReadMappingById godoc
// @Summary Delete Read Mapping by Id
// @Description Method to delete read mapping by id
// @Tags Read Mapping
// @Accept json
// @Produce json
// @Param id path string true "Read Mapping Id"
// @Success 200
// @Failure 500 {object} domain.Error
// @Router /mapping/{id} [delete]
func (h MappingRestHandler) deleteReadMappingById(c *gin.Context) {
	id := c.Param(ParamId)
	err := h.processService.DeleteReadMappingById(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
	}
}

// CreateReadMapping godoc
// @Summary Create Read Mapping
// @Description Method to create read mapping
// @Tags Read Mapping
// @Accept json
// @Produce json
// @Param read_mapping body domain.ReadMapping true "Read Mapping (without id)"
// @Success 200 {object} domain.ReadMapping
// @Failure 500 {object} domain.Error
// @Router /mapping [post]
func (h MappingRestHandler) createReadMapping(c *gin.Context) {
	var obj domain.ReadMapping
	if err := c.BindJSON(&obj); err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	result, err := h.processService.CreateReadMapping(c.Request.Context(), &obj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
	}
	c.JSON(http.StatusOK, result)
}
