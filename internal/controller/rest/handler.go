package rest

import (
	"context"
	"log"
	"net/http"

	"caviar/internal/dto"
	"caviar/internal/models"
	"caviar/internal/types"
	"caviar/pkg/apperror"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type ProductService interface {
	Create(ctx context.Context, input *dto.ProductCreateDTO) error
	List(ctx context.Context, isAuthenticated bool, filter *types.ProductFilter) ([]*models.Product, error)
	Update(ctx context.Context, input *dto.ProductUpdateDTO) error
	Delete(ctx context.Context, id string) error
}

type Handler struct {
	port string
	authSecret string
	productService ProductService
	isProd bool
}

func NewHandler(
	port string, 
	authSecret string, 
	productService ProductService,
	isProd bool,
) *Handler {
	return &Handler{
		port: port,
		authSecret: authSecret,
		productService: productService,
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
	
	h.initProductRoutes(api)

	if err := r.Run(":" + h.port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (h *Handler) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperror.AppError); ok {
		c.JSON(appErr.HTTPStatus(), gin.H{
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"code":    apperror.CodeInternal,
			"message": "Internal server error",
		},
	})
}

func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"message": "Caviar API is running",
	})
}

