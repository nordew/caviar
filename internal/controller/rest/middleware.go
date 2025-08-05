package rest

import (
	"fmt"
	"net/http"

	"caviar/pkg/apperror"

	"github.com/gin-gonic/gin"
)

func (h *Handler) LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
		)
	})
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header is required",
				},
			})
			c.Abort()
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid authorization format. Use 'Bearer <token>'",
				},
			})
			c.Abort()
			return
		}

		token := authHeader[7:]
		if token != h.authSecret {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid authorization token",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (h *Handler) ErrorMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok {
			h.handleError(c, err)
		} else {
			h.handleError(c, apperror.New(apperror.CodeInternal, "Internal server error"))
		}
	})
} 