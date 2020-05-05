package rest

import (
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type MappingRestHandler struct {
	readMappingService domain.ReadMappingService
}

func NewMappingRestHandler(readMappingService domain.ReadMappingService) *MappingRestHandler {
	return &MappingRestHandler{readMappingService: readMappingService}
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
	var result domain.ReadMapping
	if err := h.readMappingService.GetById(c.Request.Context(), id, &result); err != nil {
		log.Error(err)
		if domain.ECode(err) == domain.ErrNotFound {
			c.Status(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusInternalServerError, E(err))
		return
	}
	c.JSON(http.StatusOK, result)
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
	var results []domain.ReadMapping
	if err := h.readMappingService.GetAll(c.Request.Context(), &results); err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, E(err))
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
	if err := h.readMappingService.DeleteById(c.Request.Context(), id); err != nil {
		log.Error(err)
		if domain.ECode(err) == domain.ErrNotFound {
			c.Status(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusInternalServerError, E(err))
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
		log.Error(err)
		c.JSON(http.StatusInternalServerError, E(err))
		return
	}
	if err := h.readMappingService.Create(c.Request.Context(), &obj); err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, E(err))
		return
	}
	c.JSON(http.StatusOK, obj)
}
