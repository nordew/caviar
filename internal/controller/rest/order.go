package rest

import (
	"net/http"
	"strconv"
	"time"

	"caviar/internal/dto"
	"caviar/internal/models"
	"caviar/internal/types"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initOrderRoutes(api *gin.RouterGroup) {
	orders := api.Group("/orders")
	orders.POST("", h.createOrder)

	ordersProtected := orders.Group("/", h.AuthMiddleware())
	ordersProtected.GET("", h.listOrders)
	ordersProtected.GET("/statistics", h.getOrderStatistics)
	ordersProtected.GET("/:id", h.getOrder)
	ordersProtected.GET("/number/:orderNumber", h.getOrderByNumber)
	ordersProtected.PUT("/:id/status", h.updateOrderStatus)
	ordersProtected.DELETE("/:id", h.deleteOrder)
}

// @Summary Create a new order
// @Description Create a new order with customer and delivery information
// @Tags orders
// @Accept json
// @Produce json
// @Param order body dto.OrderCreateDTO true "Order creation data"
// @Success 201 {object} dto.OrderResponseDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders [post]
func (h *Handler) createOrder(c *gin.Context) {
	var input dto.OrderCreateDTO
	if !h.bindJSON(c, &input) {
		return
	}

	order, err := h.orderService.Create(c.Request.Context(), &input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.converter.Order.ToResponseDTO(order)
	c.JSON(http.StatusCreated, response)
}

// @Summary Get order by ID
// @Description Retrieve an order by its ID
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} dto.OrderResponseDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders/{id} [get]
func (h *Handler) getOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "order ID is required",
			},
		})
		return
	}

	order, err := h.orderService.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.converter.Order.ToResponseDTO(order)
	c.JSON(http.StatusOK, response)
}

// @Summary Get order by order number
// @Description Retrieve an order by its order number
// @Tags orders
// @Produce json
// @Param orderNumber path string true "Order Number"
// @Success 200 {object} dto.OrderResponseDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders/number/{orderNumber} [get]
func (h *Handler) getOrderByNumber(c *gin.Context) {
	orderNumber := c.Param("orderNumber")
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "order number is required",
			},
		})
		return
	}

	order, err := h.orderService.GetByOrderNumber(c.Request.Context(), orderNumber)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.converter.Order.ToResponseDTO(order)
	c.JSON(http.StatusOK, response)
}

// @Summary List orders
// @Description Retrieve a list of orders with filtering and pagination
// @Tags orders
// @Produce json
// @Param status query string false "Filter by order status"
// @Param customer_phone query string false "Filter by customer phone"
// @Param country query string false "Filter by delivery country"
// @Param created_from query string false "Filter by creation date from (RFC3339 format)"
// @Param created_to query string false "Filter by creation date to (RFC3339 format)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.OrderListResponseDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders [get]
func (h *Handler) listOrders(c *gin.Context) {
	filter := &types.OrderFilter{}

	if status := c.Query("status"); status != "" {
		filter.Status = status
	}

	if phone := c.Query("customer_phone"); phone != "" {
		filter.CustomerPhone = phone
	}

	if country := c.Query("country"); country != "" {
		filter.Country = country
	}

	if createdFrom := c.Query("created_from"); createdFrom != "" {
		if t, err := time.Parse(time.RFC3339, createdFrom); err == nil {
			filter.CreatedFrom = t
		}
	}

	if createdTo := c.Query("created_to"); createdTo != "" {
		if t, err := time.Parse(time.RFC3339, createdTo); err == nil {
			filter.CreatedTo = t
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	filter.Offset = (page - 1) * limit
	filter.Limit = limit

	orders, total, err := h.orderService.List(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.converter.Order.ToListResponseDTO(orders, page, limit, total)

	c.JSON(http.StatusOK, response)
}

// @Summary Update order status
// @Description Update the status of an existing order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param status body dto.OrderStatusUpdateDTO true "Status update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders/{id}/status [put]
func (h *Handler) updateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "order ID is required",
			},
		})
		return
	}

	var input dto.OrderStatusUpdateDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": err.Error(),
			},
		})
		return
	}

	status := models.OrderStatus(input.Status)
	err := h.orderService.UpdateStatus(c.Request.Context(), id, status)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
		"status":  input.Status,
	})
}

// @Summary Delete order
// @Description Delete an order by ID
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders/{id} [delete]
func (h *Handler) deleteOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "order ID is required",
			},
		})
		return
	}

	err := h.orderService.Delete(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order deleted successfully",
	})
}

// @Summary Get order statistics
// @Description Get order statistics including counts by status and country
// @Tags orders
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders/statistics [get]
func (h *Handler) getOrderStatistics(c *gin.Context) {
	stats, err := h.orderService.GetStatistics(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

