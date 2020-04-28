package rest

import (
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

type OrderRestHandler struct {
	orderService domain.OrderService
}

func NewOrderRestHandler(orderService domain.OrderService) *OrderRestHandler {
	return &OrderRestHandler{orderService: orderService}
}

func (h OrderRestHandler) Register(router *gin.Engine) {
	group := router.Group("/order")
	group.GET("/:"+ParamId, h.getOrderById)
	group.GET("/", h.getOrders)
	group.POST("/:"+ParamProcessId, h.submitOrder)
}

// GetOrderById godoc
// @Summary Get Order by Id
// @Description Method to get Order by id
// @Tags Order
// @Accept json
// @Produce json
// @Param id path string true "Order Id"
// @Success 200 {object} domain.Order
// @Failure 500 {object} domain.Error
// @Router /order/{id} [get]
func (h OrderRestHandler) getOrderById(c *gin.Context) {
	id := c.Param(ParamId)
	result, err := h.orderService.GetOrderById(c.Request.Context(), id)
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

// GetOrderById godoc
// @Summary Get Orders
// @Description Method to get all orders
// @Tags Order
// @Accept json
// @Produce json
// @Success 200 {array} domain.Order
// @Failure 500 {object} domain.Error
// @Router /order [get]
func (h OrderRestHandler) getOrders(c *gin.Context) {
	results, err := h.orderService.GetOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	c.JSON(http.StatusOK, results)
}

// SubmitOrder godoc
// @Summary Submit Order
// @Description Method to submit order
// @Tags Order
// @Accept json
// @Produce json
// @Param process_id path string true "Process Id"
// @Param order body domain.Order true "Order (without id)"
// @Success 200 {object} domain.Order
// @Failure 500 {object} domain.Error
// @Router /order [post]
func (h OrderRestHandler) submitOrder(c *gin.Context) {
	processId := c.Param(ParamProcessId)
	var obj domain.Order
	if err := c.BindJSON(&obj); err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	result, err := h.orderService.SubmitOrder(c.Request.Context(), &obj, processId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	c.JSON(http.StatusOK, result)
}
