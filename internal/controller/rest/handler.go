package rest

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"caviar/internal/dto"
	"caviar/internal/dto/converter"
	"caviar/internal/models"
	"caviar/internal/types"
	"caviar/pkg/apperror"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type ProductService interface {
	Create(ctx context.Context, input *dto.ProductCreateDTO) error
	List(ctx context.Context, isAuthenticated bool, filter *types.ProductFilter) ([]*models.Product, error)
	Update(ctx context.Context, input *dto.ProductUpdateDTO) error
	Delete(ctx context.Context, id string) error
}

type OrderService interface {
	Create(ctx context.Context, input *dto.OrderCreateDTO) (*models.Order, error)
	GetByID(ctx context.Context, id string) (*models.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error)
	List(ctx context.Context, filter *types.OrderFilter) ([]*models.Order, int64, error)
	UpdateStatus(ctx context.Context, id string, status models.OrderStatus) error
	Delete(ctx context.Context, id string) error
	GetStatistics(ctx context.Context) (map[string]any, error)
}

type Handler struct {
	port            string
	authSecret      string
	productService  ProductService
	orderService    OrderService
	logger          *zap.Logger
	converter       *converter.Converter
	isProd          bool
}

func NewHandler(
	port string, 
	productService ProductService,
	orderService OrderService,
	logger *zap.Logger,
	isProd bool,
) *Handler {
	return &Handler{
		port:            port,
		productService:  productService,
		orderService:    orderService,
		logger:          logger,
		converter:       converter.NewConverter(),
		isProd:          isProd,
	}
}

func (h *Handler) RegisterAndRun(r *gin.Engine) {
	r.Use(h.LoggerMiddleware())
	r.Use(h.ErrorMiddleware())
	
	if !h.isProd {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	
	api := r.Group("/api/v1")
	
	api.GET("/health", h.healthCheck)
	if !h.isProd {
		api.GET("/debug/auth", h.debugAuth)
	}
	
	h.initProductRoutes(api)
	h.initOrderRoutes(api)

	if err := r.Run(":" + h.port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

type APIResponse struct {
	Success   bool        `json:"success"`
	Data      any `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *APIMeta    `json:"meta,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp string      `json:"timestamp"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type APIMeta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

func (h *Handler) handleError(c *gin.Context, err error) {
	requestID := h.getRequestID(c)
	
	h.logger.Error("API Error",
		zap.String("request_id", requestID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("ip", c.ClientIP()),
		zap.Error(err),
	)

	var httpStatus int
	var apiError *APIError

	if appErr, ok := err.(*apperror.AppError); ok {
		httpStatus = appErr.HTTPStatus()
		apiError = &APIError{
			Code:    string(appErr.Code),
			Message: appErr.Message,
		}
		
		if !h.isProd {
			if appErr.Err != nil {
				apiError.Details = appErr.Err.Error()
			} else {
				apiError.Details = fmt.Sprintf("%+v", appErr)
			}
		}
	} else {
		httpStatus = http.StatusInternalServerError
		apiError = &APIError{
			Code:    string(apperror.CodeInternal),
			Message: "An unexpected error occurred",
		}
		
		if !h.isProd {
			apiError.Details = err.Error()
		}
	}

	response := APIResponse{
		Success:   false,
		Error:     apiError,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(httpStatus, response)
}

func (h *Handler) handleValidationError(c *gin.Context, err error) {
	requestID := h.getRequestID(c)
	
	h.logger.Warn("Validation Error",
		zap.String("request_id", requestID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Error(err),
	)

	apiError := &APIError{
		Code:    "VALIDATION_ERROR",
		Message: "Invalid input data",
	}
	
	if !h.isProd {
		apiError.Details = err.Error()
	}

	response := APIResponse{
		Success:   false,
		Error:     apiError,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusBadRequest, response)
}

func (h *Handler) handleNotFound(c *gin.Context, resource string) {
	requestID := h.getRequestID(c)
	
	h.logger.Warn("Resource Not Found",
		zap.String("request_id", requestID),
		zap.String("resource", resource),
		zap.String("path", c.Request.URL.Path),
	)

	apiError := &APIError{
		Code:    string(apperror.CodeNotFound),
		Message: resource + " not found",
	}

	response := APIResponse{
		Success:   false,
		Error:     apiError,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusNotFound, response)
}

func (h *Handler) getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	
	if requestID := c.GetString("request_id"); requestID != "" {
		return requestID
	}
	
	return "req_" + time.Now().Format("20060102150405") + "_" + c.ClientIP()
}

func (h *Handler) healthCheck(c *gin.Context) {
	h.handleSuccess(c, http.StatusOK, map[string]any{
		"status":  "healthy",
		"service": "Caviar API",
		"version": "1.0.0",
	})
}

func (h *Handler) debugAuth(c *gin.Context) {
	h.handleSuccess(c, http.StatusOK, map[string]any{
		"auth_secret_set": h.authSecret != "",
		"auth_secret_length": len(h.authSecret),
		"auth_secret_preview": func() string {
			if h.authSecret == "" {
				return "EMPTY"
			}
			if len(h.authSecret) > 6 {
				return h.authSecret[:3] + "..." + h.authSecret[len(h.authSecret)-3:]
			}
			return "***"
		}(),
	})
}

// handleSuccess returns standardized success responses
func (h *Handler) handleSuccess(c *gin.Context, statusCode int, data any, meta ...*APIMeta) {
	requestID := h.getRequestID(c)
	
	// Log successful responses (optional, can be removed in production for performance)
	if !h.isProd {
		h.logger.Info("API Success",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status_code", statusCode),
		)
	}

	response := APIResponse{
		Success:   true,
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if len(meta) > 0 && meta[0] != nil {
		response.Meta = meta[0]
	}

	c.JSON(statusCode, response)
}

// handleCreated handles successful resource creation
func (h *Handler) handleCreated(c *gin.Context, data any, message ...string) {
	msg := "Resource created successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	
	h.handleSuccess(c, http.StatusCreated, map[string]any{
		"message": msg,
		"data":    data,
	})
}

// handleUpdated handles successful resource updates
func (h *Handler) handleUpdated(c *gin.Context, data any, message ...string) {
	msg := "Resource updated successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	
	h.handleSuccess(c, http.StatusOK, map[string]any{
		"message": msg,
		"data":    data,
	})
}

// handleDeleted handles successful resource deletion
func (h *Handler) handleDeleted(c *gin.Context, message ...string) {
	msg := "Resource deleted successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	
	h.handleSuccess(c, http.StatusOK, map[string]any{
		"message": msg,
	})
}

// handleList handles paginated list responses
func (h *Handler) handleList(c *gin.Context, data any, page, limit, total int) {
	totalPages := (total + limit - 1) / limit
	
	meta := &APIMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
	
	h.handleSuccess(c, http.StatusOK, data, meta)
}

// bindJSON safely binds JSON with validation error handling
func (h *Handler) bindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		h.handleValidationError(c, err)
		return false
	}
	return true
}

// getPathParam safely extracts path parameters with validation
func (h *Handler) getPathParam(c *gin.Context, key string, required bool) (string, bool) {
	value := c.Param(key)
	if required && value == "" {
		h.handleValidationError(c, 
			apperror.New("MISSING_PARAMETER", key + " parameter is required"))
		return "", false
	}
	return value, true
}

