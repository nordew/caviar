package rest

import (
	"net/http"

	"caviar/pkg/apperror"

	"github.com/gin-gonic/gin"
)

type Handler struct {
}

// handleError is a helper method to handle errors and write appropriate HTTP responses
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

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.Use(h.errorMiddleware())
	
	api := r.Group("/api/v1")
	
	api.GET("/health", h.healthCheck)
}

func (h *Handler) errorMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok {
			h.handleError(c, err)
		} else {
			h.handleError(c, apperror.New(apperror.CodeInternal, "Internal server error"))
		}
	})
}

func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"message": "Caviar API is running",
	})
}

